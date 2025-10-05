package github

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"encore.dev/beta/errs"
	"resty.dev/v3"
)

// GitHubStarsResponse represents the API response for GitHub stars
type GitHubStarsResponse struct {
	Stars int `json:"stars"`
}

// Regular expression to extract star count from GitHub Link header
var githubStargazersLinkRegex = regexp.MustCompile(`^.+,.+page=(\d+).+$`)

//encore:api public method=GET path=/:owner/:repo
func GetStars(ctx context.Context, owner, repo string) (*GitHubStarsResponse, error) {
	stars, err := getGitHubStars(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &GitHubStarsResponse{Stars: stars}, nil
}

func getGitHubStars(ctx context.Context, owner, repo string) (int, error) {
	// Create Resty client
	client := resty.New()
	defer client.Close()

	// Define error response struct
	var githubError struct {
		Message string `json:"message"`
	}

	// Build the stargazers URL with per_page=1 and make the request
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/stargazers", url.PathEscape(owner), url.PathEscape(repo))

	resp, err := client.R().
		SetContext(ctx).
		SetQueryParam("per_page", "1").
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		SetError(&githubError).
		Get(apiURL)

	if err != nil {
		return 0, &errs.Error{
			Code:    errs.Unavailable,
			Message: "failed to fetch from GitHub API",
		}
	}

	// Handle non-OK responses
	if resp.StatusCode() == 404 {
		return 0, &errs.Error{
			Code:    errs.NotFound,
			Message: "repository not found",
		}
	}

	if !resp.IsSuccess() {
		// Check if we have a GitHub error response
		if githubError.Message != "" {
			return 0, &errs.Error{
				Code:    errs.FailedPrecondition,
				Message: fmt.Sprintf("GitHub API error: %s", githubError.Message),
			}
		}

		return 0, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: fmt.Sprintf("GitHub API returned status %d", resp.StatusCode()),
		}
	}

	// Get the Link header
	linkHeader := resp.Header().Get("Link")

	// Parse the Link header to extract star count
	// The regex looks for the last page number in pagination links
	matches := githubStargazersLinkRegex.FindStringSubmatch(linkHeader)
	if len(matches) < 2 {
		return 0, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to parse Link header for star count",
		}
	}

	// Convert the extracted page number to int
	stars, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to parse star count from Link header",
		}
	}

	return stars, nil
}
