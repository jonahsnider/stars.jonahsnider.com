# stars.jonahsnider.com

A fast, simple API for getting the number of stars for a GitHub repository.

## Usage

```
GET https://stars.jonahsnider.com/{owner}/{repo}
```

**Example:**

```bash
curl https://stars.jonahsnider.com/facebook/react
```

**Response:**

```json
{
  "stars": 228000
}
```

## How it works

The GitHub pagination API exposes the total number of stars for a repository in the `Link` header.
Instead of fetching the whole JSON blob of data for a repository, you can fetch a single stargazer for the repository and extract the stargazer count from that header.
