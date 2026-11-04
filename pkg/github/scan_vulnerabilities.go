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
	"sort"
	"strings"

	gh "github.com/google/go-github/v79/github"
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
	ImageName       string              `json:"imageName" yaml:"imageName"`
	Source          string              `json:"source" yaml:"source"`
	Conclusion      string              `json:"conclusion" yaml:"conclusion"`
	Summary         ScanSummary         `json:"summary" yaml:"summary"`
	Vulnerabilities []ScanVulnerability `json:"vulnerabilities,omitempty" yaml:"vulnerabilities,omitempty"`
}

func ListScanVulnerabilities(ctx context.Context, ghClient *gh.Client, org, repo, tag string) ([]ImageScanEntry, error) {
	commitInfo, err := GetCommitForTag(ctx, ghClient, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	runID, err := findScanWorkflowRun(ctx, ghClient, org, repo, commitInfo.SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to find scan workflow run: %w", err)
	}

	// PerPage: 100 is the GitHub API max. Current scan workflows have ~30 jobs.
	jobs, _, err := ghClient.Actions.ListWorkflowJobs(ctx, org, repo, runID, &gh.ListWorkflowJobsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	artifactMap, err := buildArtifactMap(ctx, ghClient, org, repo, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to build artifact map: %w", err)
	}

	var entries []ImageScanEntry
	for _, job := range jobs.Jobs {
		if job.GetName() == "Run parameters" {
			continue
		}
		source, imageName := parseJobName(job.GetName())
		if source == "" {
			fmt.Fprintf(os.Stderr, "Skipping job with unexpected name format: %q\n", job.GetName())
			continue
		}

		conclusion := job.GetConclusion()
		scanResult, hasArtifact := artifactMap[job.GetName()]
		if conclusion == "failure" && !hasArtifact {
			conclusion = "scan error"
		}

		entry := ImageScanEntry{
			ImageName:  imageName,
			Source:     source,
			Conclusion: conclusion,
		}

		if hasArtifact {
			entry.Summary = scanResult.Summary
			entry.Vulnerabilities = scanResult.Vulnerabilities
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Source != entries[j].Source {
			return entries[i].Source == "Upstream"
		}
		return entries[i].ImageName < entries[j].ImageName
	})

	return entries, nil
}

// parseJobName extracts source and image name from "Konflux: release-central-db:4.10.3".
func parseJobName(jobName string) (source, imageName string) {
	parts := strings.SplitN(jobName, ": ", 2)
	if len(parts) != 2 {
		return "", jobName
	}
	source = parts[0]
	imageWithVersion := parts[1]
	if idx := strings.LastIndex(imageWithVersion, ":"); idx >= 0 {
		imageName = imageWithVersion[:idx]
	} else {
		imageName = imageWithVersion
	}
	return source, imageName
}

// buildArtifactMap downloads all artifacts and returns a map from job name to scan result.
func buildArtifactMap(ctx context.Context, ghClient *gh.Client, org, repo string, runID int64) (map[string]*ScanResult, error) {
	// PerPage: 100 is the GitHub API max. Current scan workflows produce ~30 artifacts.
	artifacts, _, err := ghClient.Actions.ListWorkflowRunArtifacts(ctx, org, repo, runID, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	result := make(map[string]*ScanResult)
	for _, artifact := range artifacts.Artifacts {
		downloadURL, _, err := ghClient.Actions.DownloadArtifact(ctx, org, repo, artifact.GetID(), 2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not get download URL for artifact %s: %v\n", artifact.GetName(), err)
			continue
		}

		scanResult, err := downloadAndParseArtifact(downloadURL.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not parse artifact %s: %v\n", artifact.GetName(), err)
			continue
		}

		jobName := artifactNameToJobName(artifact.GetName())
		result[jobName] = scanResult
	}
	return result, nil
}

// artifactNameToJobName converts "konflux_quay_io_rhacs-eng_release-central-db_4.10.3.json"
// to "Konflux: release-central-db:4.10.3".
func artifactNameToJobName(artifactName string) string {
	name := strings.TrimSuffix(artifactName, ".json")
	parts := strings.SplitN(name, "_quay_io_rhacs-eng_", 2)
	if len(parts) != 2 {
		return artifactName
	}

	source := parts[0]
	switch source {
	case "konflux":
		source = "Konflux"
	case "upstream":
		source = "Upstream"
	}

	imageAndVersion := parts[1]
	if idx := strings.LastIndex(imageAndVersion, "_"); idx >= 0 {
		imageAndVersion = imageAndVersion[:idx] + ":" + imageAndVersion[idx+1:]
	}

	return source + ": " + imageAndVersion
}

// findScanWorkflowRun finds the most recent scan workflow run for a commit.
// A commit may have multiple runs (tag push, branch push, re-runs).
func findScanWorkflowRun(ctx context.Context, ghClient *gh.Client, org, repo, commitSHA string) (int64, error) {
	runs, _, err := ghClient.Actions.ListWorkflowRunsByFileName(ctx, org, repo, scanWorkflowFile, &gh.ListWorkflowRunsOptions{
		HeadSHA: commitSHA,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	if len(runs.WorkflowRuns) == 0 {
		return 0, fmt.Errorf("no scan workflow runs found for commit %s", commitSHA)
	}

	latest := runs.WorkflowRuns[0]
	for _, run := range runs.WorkflowRuns[1:] {
		if run.GetCreatedAt().After(latest.GetCreatedAt().Time) {
			latest = run
		}
	}

	if len(runs.WorkflowRuns) > 1 {
		fmt.Fprintf(os.Stderr, "Found %d scan workflow runs for commit %s, using the latest\n", len(runs.WorkflowRuns), commitSHA[:12])
	}

	return latest.GetID(), nil
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
