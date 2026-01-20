package helpers

import (
	"net/http"

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
