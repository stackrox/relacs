package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	gh "github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client"
)

const scanWorkflowFile = "scan-image-vulnerabilities.yaml"

type scanResultWrapper struct {
	Result ScanResult `json:"result"`
}

type ScanResult struct {
	Summary         ScanSummary         `json:"summary"`
	Vulnerabilities []ScanVulnerability `json:"vulnerabilities"`
}

type ScanSummary struct {
	Critical             int `json:"CRITICAL"`
	Important            int `json:"IMPORTANT"`
	Moderate             int `json:"MODERATE"`
	Low                  int `json:"LOW"`
	TotalComponents      int `json:"TOTAL-COMPONENTS"`
	TotalVulnerabilities int `json:"TOTAL-VULNERABILITIES"`
}

type ScanVulnerability struct {
	CVEId                 string  `json:"cveId"`
	CVESeverity           string  `json:"cveSeverity"`
	CVECVSS               float64 `json:"cveCVSS"`
	CVEInfo               string  `json:"cveInfo"`
	ComponentName         string  `json:"componentName"`
	ComponentVersion      string  `json:"componentVersion"`
	ComponentFixedVersion string  `json:"componentFixedVersion"`
}

type ImageScanEntry struct {
	ArtifactName    string              `json:"artifactName" yaml:"artifactName"`
	Summary         ScanSummary         `json:"summary" yaml:"summary"`
	Vulnerabilities []ScanVulnerability `json:"vulnerabilities,omitempty" yaml:"vulnerabilities,omitempty"`
}

func ListScanVulnerabilities(ctx context.Context, c *client.Client, org, repo, tag string) ([]ImageScanEntry, error) {
	commitInfo, err := GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	runID, err := findScanWorkflowRun(ctx, c, org, repo, commitInfo.SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to find scan workflow run: %w", err)
	}

	// PerPage: 100 is the GitHub API max. Current scan workflows produce ~30 artifacts.
	artifacts, _, err := c.GetGitHub().Actions.ListWorkflowRunArtifacts(ctx, org, repo, runID, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	var entries []ImageScanEntry
	for _, artifact := range artifacts.Artifacts {
		downloadURL, _, err := c.GetGitHub().Actions.DownloadArtifact(ctx, org, repo, artifact.GetID(), 2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not get download URL for artifact %s: %v\n", artifact.GetName(), err)
			continue
		}

		scanResult, err := downloadAndParseArtifact(downloadURL.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not parse artifact %s: %v\n", artifact.GetName(), err)
			continue
		}

		entries = append(entries, ImageScanEntry{
			ArtifactName:    artifact.GetName(),
			Summary:         scanResult.Summary,
			Vulnerabilities: scanResult.Vulnerabilities,
		})
	}
	return entries, nil
}

func findScanWorkflowRun(ctx context.Context, c *client.Client, org, repo, commitSHA string) (int64, error) {
	runs, _, err := c.GetGitHub().Actions.ListWorkflowRunsByFileName(ctx, org, repo, scanWorkflowFile, &gh.ListWorkflowRunsOptions{
		HeadSHA: commitSHA,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	if len(runs.WorkflowRuns) == 0 {
		return 0, fmt.Errorf("no scan workflow runs found for commit %s", commitSHA)
	}

	return runs.WorkflowRuns[0].GetID(), nil
}

func downloadAndParseArtifact(url string) (*ScanResult, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download artifact: HTTP %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to open artifact zip: %w", err)
	}

	for _, f := range zipReader.File {
		if f.Name != "scan-result.json" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open scan-result.json: %w", err)
		}
		defer rc.Close() //nolint:errcheck

		var wrapper scanResultWrapper
		if err := json.NewDecoder(rc).Decode(&wrapper); err != nil {
			return nil, fmt.Errorf("failed to parse scan-result.json: %w", err)
		}
		return &wrapper.Result, nil
	}

	return nil, fmt.Errorf("scan-result.json not found in artifact")
}
