package resource

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

func (g *GitHubClient) listReleasesV4() ([]*github.RepositoryRelease, error) {
	if g.clientV4 == nil {
		return nil, errors.New("github graphql is not been initialised")
	}
	var listReleases struct {
		Repository struct {
			Releases struct {
				Edges []struct {
					Node struct {
						ReleaseObject
					}
				} `graphql:"edges"`
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				} `graphql:"pageInfo"`
			} `graphql:"releases(first:$releasesCount, after: $releaseCursor, orderBy: {field: CREATED_AT, direction: DESC})"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}

	vars := map[string]interface{}{
		"repositoryOwner": githubv4.String(g.owner),
		"repositoryName":  githubv4.String(g.repository),
		"releaseCursor":   (*githubv4.String)(nil),
		"releasesCount":   githubv4.Int(100),
	}

	var allReleases []*github.RepositoryRelease
	for {
		if err := g.clientV4.Query(context.TODO(), &listReleases, vars); err != nil {
			return nil, err
		}

		for _, r := range listReleases.Repository.Releases.Edges {
			r := r
			publishedAt, _ := time.ParseInLocation(time.RFC3339, r.Node.PublishedAt.Time.Format(time.RFC3339), time.UTC)
			createdAt, _ := time.ParseInLocation(time.RFC3339, r.Node.CreatedAt.Time.Format(time.RFC3339), time.UTC)
			releaseID, _ := strconv.ParseInt(r.Node.ID, 10, 64)
			allReleases = append(allReleases, &github.RepositoryRelease{
				ID:          &releaseID,
				TagName:     &r.Node.TagName,
				Name:        &r.Node.Name,
				Prerelease:  &r.Node.IsPrerelease,
				Draft:       &r.Node.IsDraft,
				URL:         &r.Node.URL,
				PublishedAt: &github.Timestamp{Time: publishedAt},
				CreatedAt:   &github.Timestamp{Time: createdAt},
			})
		}

		if !listReleases.Repository.Releases.PageInfo.HasNextPage {
			break
		}
		vars["releaseCursor"] = listReleases.Repository.Releases.PageInfo.EndCursor
	}

	return allReleases, nil
}
