package helpers

import (
	"net/http"
	"strconv"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/client"
)

// SetupTestClient creates a test client with gock-intercepted HTTP client
func SetupTestClient() (*client.Client, error) {
	httpClient := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(httpClient)
	return NewTestClient(httpClient)
}

// TearDownGock cleans up gock mocks
func TearDownGock() {
	gock.Off()
	gock.Clean()
}

// MockLightweightTag mocks the complete flow for a lightweight tag
// This mocks both the tag ref lookup and the commit lookup
func MockLightweightTag(org, repo, tag, commitSHA string) {
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/ref/tags/" + tag).
		Reply(200).
		JSON(map[string]interface{}{
			"ref": "refs/tags/" + tag,
			"object": map[string]interface{}{
				"type": "commit",
				"sha":  commitSHA,
			},
		})

	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/commits/" + commitSHA).
		Reply(200).
		JSON(map[string]interface{}{
			"sha": commitSHA,
		})
}

// MockAnnotatedTag mocks the complete flow for an annotated tag
// This mocks the tag ref lookup, tag object lookup, and commit lookup
func MockAnnotatedTag(org, repo, tag, tagObjectSHA, commitSHA string) {
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/ref/tags/" + tag).
		Reply(200).
		JSON(map[string]interface{}{
			"ref": "refs/tags/" + tag,
			"object": map[string]interface{}{
				"type": "tag",
				"sha":  tagObjectSHA,
			},
		})

	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/tags/" + tagObjectSHA).
		Reply(200).
		JSON(map[string]interface{}{
			"sha": tagObjectSHA,
			"object": map[string]interface{}{
				"sha": commitSHA,
			},
		})

	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/commits/" + commitSHA).
		Reply(200).
		JSON(map[string]interface{}{
			"sha": commitSHA,
		})
}

// MockGitCommitBySHA mocks getting a commit by short SHA (for git-describe versions)
func MockGitCommitBySHA(org, repo, shortSHA, fullSHA string) {
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/commits/" + shortSHA).
		Reply(200).
		JSON(map[string]interface{}{
			"sha": fullSHA,
		})
}

// MockTagNotFound mocks a 404 response for a tag that doesn't exist
func MockTagNotFound(org, repo, tag string) {
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/git/ref/tags/" + tag).
		Reply(404).
		JSON(map[string]interface{}{
			"message": "Not Found",
		})
}

// PRData represents a pull request for mocking
type PRData struct {
	Number  int
	Title   string
	Author  string
	HeadSHA string
	IsDraft bool
	HTMLURL string
}

// CheckRunData represents a check run for mocking
type CheckRunData struct {
	Conclusion string // "success", "failure", "timed_out", "action_required", ""
	Status     string // "completed", "in_progress", "queued"
}

// ReviewData represents a review for mocking
type ReviewData struct {
	State string // "APPROVED", "CHANGES_REQUESTED", "COMMENTED"
}

// MockListPullRequests mocks the PR list endpoint
func MockListPullRequests(org, repo, baseBranch string, prs []PRData) {
	prList := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		prList[i] = map[string]interface{}{
			"number":   pr.Number,
			"title":    pr.Title,
			"draft":    pr.IsDraft,
			"html_url": pr.HTMLURL,
			"user": map[string]interface{}{
				"login": pr.Author,
			},
			"head": map[string]interface{}{
				"sha": pr.HeadSHA,
			},
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls").
		MatchParam("base", baseBranch).
		MatchParam("state", "open").
		MatchParam("direction", "asc").
		MatchParam("per_page", "100").
		Reply(200).
		JSON(prList)
}

// MockListPullRequestsPaginated mocks paginated PR list
func MockListPullRequestsPaginated(org, repo, baseBranch string, page1PRs, page2PRs []PRData) {
	// Page 1 (no page parameter in first request)
	prList1 := make([]map[string]interface{}, len(page1PRs))
	for i, pr := range page1PRs {
		prList1[i] = map[string]interface{}{
			"number":   pr.Number,
			"title":    pr.Title,
			"draft":    pr.IsDraft,
			"html_url": pr.HTMLURL,
			"user": map[string]interface{}{
				"login": pr.Author,
			},
			"head": map[string]interface{}{
				"sha": pr.HeadSHA,
			},
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls").
		MatchParam("base", baseBranch).
		MatchParam("state", "open").
		MatchParam("direction", "asc").
		MatchParam("per_page", "100").
		Reply(200).
		SetHeader("Link", `<https://api.github.com/repos/`+org+`/`+repo+`/pulls?base=`+baseBranch+`&state=open&direction=asc&per_page=100&page=2>; rel="next"`).
		JSON(prList1)

	// Page 2
	prList2 := make([]map[string]interface{}, len(page2PRs))
	for i, pr := range page2PRs {
		prList2[i] = map[string]interface{}{
			"number":   pr.Number,
			"title":    pr.Title,
			"draft":    pr.IsDraft,
			"html_url": pr.HTMLURL,
			"user": map[string]interface{}{
				"login": pr.Author,
			},
			"head": map[string]interface{}{
				"sha": pr.HeadSHA,
			},
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls").
		MatchParam("base", baseBranch).
		MatchParam("state", "open").
		MatchParam("direction", "asc").
		MatchParam("per_page", "100").
		MatchParam("page", "2").
		Reply(200).
		JSON(prList2)
}

// MockCombinedStatus mocks the combined status endpoint
func MockCombinedStatus(org, repo, sha, state string) {
	totalCount := 0
	if state != "" {
		totalCount = 1
	}
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/commits/" + sha + "/status").
		Reply(200).
		JSON(map[string]interface{}{
			"state":       state,
			"total_count": totalCount,
		})
}

// MockCheckRuns mocks the check runs endpoint
func MockCheckRuns(org, repo, sha string, checkRuns []CheckRunData) {
	checkRunsList := make([]map[string]interface{}, len(checkRuns))
	for i, run := range checkRuns {
		checkRunsList[i] = map[string]interface{}{
			"conclusion": run.Conclusion,
			"status":     run.Status,
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/commits/" + sha + "/check-runs").
		Reply(200).
		JSON(map[string]interface{}{
			"total_count": len(checkRuns),
			"check_runs":  checkRunsList,
		})
}

// MockListReviews mocks the reviews endpoint
func MockListReviews(org, repo string, prNumber int, reviews []ReviewData) {
	reviewsList := make([]map[string]interface{}, len(reviews))
	for i, review := range reviews {
		reviewsList[i] = map[string]interface{}{
			"state": review.State,
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls/"+strconv.Itoa(prNumber)+"/reviews").
		MatchParam("per_page", "100").
		Reply(200).
		JSON(reviewsList)
}

// MockListReviewsPaginated mocks paginated reviews
func MockListReviewsPaginated(org, repo string, prNumber int, page1Reviews, page2Reviews []ReviewData) {
	// Page 1
	reviewsList1 := make([]map[string]interface{}, len(page1Reviews))
	for i, review := range page1Reviews {
		reviewsList1[i] = map[string]interface{}{
			"state": review.State,
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls/"+strconv.Itoa(prNumber)+"/reviews").
		MatchParam("per_page", "100").
		MatchParam("page", "1").
		Reply(200).
		SetHeader("Link", `<https://api.github.com/repos/`+org+`/`+repo+`/pulls/`+strconv.Itoa(prNumber)+`/reviews?per_page=100&page=2>; rel="next"`).
		JSON(reviewsList1)

	// Page 2
	reviewsList2 := make([]map[string]interface{}, len(page2Reviews))
	for i, review := range page2Reviews {
		reviewsList2[i] = map[string]interface{}{
			"state": review.State,
		}
	}

	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/pulls/"+strconv.Itoa(prNumber)+"/reviews").
		MatchParam("per_page", "100").
		MatchParam("page", "2").
		Reply(200).
		JSON(reviewsList2)
}

// MockAddTeamMembership mocks a successful PUT to add a user to a team
func MockAddTeamMembership(org, team, user, role, state string) {
	gock.New("https://api.github.com").
		Put("/orgs/" + org + "/teams/" + team + "/memberships/" + user).
		Reply(200).
		JSON(map[string]interface{}{
			"role":  role,
			"state": state,
			"url":   "https://api.github.com/orgs/" + org + "/teams/" + team + "/memberships/" + user,
		})
}

// MockAddTeamMembershipNotFound mocks a 404 when the team does not exist
func MockAddTeamMembershipNotFound(org, team, user string) {
	gock.New("https://api.github.com").
		Put("/orgs/" + org + "/teams/" + team + "/memberships/" + user).
		Reply(404).
		JSON(map[string]interface{}{
			"message": "Not Found",
		})
}

// MockAddTeamMembershipUnauthorized mocks a 403 when the token lacks permissions
func MockAddTeamMembershipUnauthorized(org, team, user string) {
	gock.New("https://api.github.com").
		Put("/orgs/" + org + "/teams/" + team + "/memberships/" + user).
		Reply(403).
		JSON(map[string]interface{}{
			"message": "Must have admin rights to Repository.",
		})
}

// MockAddTeamMembershipUnprocessable mocks a 422 when the request is invalid
func MockAddTeamMembershipUnprocessable(org, team, user string) {
	gock.New("https://api.github.com").
		Put("/orgs/" + org + "/teams/" + team + "/memberships/" + user).
		Reply(422).
		JSON(map[string]interface{}{
			"message": "Validation Failed",
		})
}

// MockRepositoryNotFound mocks a 404 response for a repository that doesn't exist
func MockRepositoryNotFound(org, repo string) {
	gock.New("https://api.github.com").
		Get("/repos/" + org + "/" + repo + "/pulls").
		Reply(404).
		JSON(map[string]interface{}{
			"message": "Not Found",
		})
}

// MockGetFileContents mocks the GetContents API endpoint for fetching a file
func MockGetFileContents(org, repo, path, ref, content string) {
	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/contents/"+path).
		MatchParam("ref", ref).
		Reply(200).
		JSON(map[string]interface{}{
			"type":     "file",
			"encoding": "base64",
			"name":     path,
			"path":     path,
			"content":  content,
		})
}

// MockFileNotFound mocks a 404 response for a file that doesn't exist
func MockFileNotFound(org, repo, path, ref string) {
	gock.New("https://api.github.com").
		Get("/repos/"+org+"/"+repo+"/contents/"+path).
		MatchParam("ref", ref).
		Reply(404).
		JSON(map[string]interface{}{
			"message": "Not Found",
		})
}
