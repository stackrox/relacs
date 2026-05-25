package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

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
	JobName         string              `json:"jobName" yaml:"jobName"`
	Conclusion      string              `json:"conclusion" yaml:"conclusion"`
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

	// PerPage: 100 is the GitHub API max. Current scan workflows have ~30 jobs.
	jobs, _, err := c.GetGitHub().Actions.ListWorkflowJobs(ctx, org, repo, runID, &gh.ListWorkflowJobsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	var entries []ImageScanEntry
	for _, job := range jobs.Jobs {
		if job.GetName() == "Run parameters" {
			continue
		}

		entry := ImageScanEntry{
			JobName:    job.GetName(),
			Conclusion: job.GetConclusion(),
		}

		if job.GetConclusion() == "failure" {
			logURL, _, err := c.GetGitHub().Actions.GetWorkflowJobLogs(ctx, org, repo, job.GetID(), 2)
			if err != nil {
				return nil, fmt.Errorf("failed to get logs for job %s: %w", job.GetName(), err)
			}

			logText, err := downloadLogs(logURL.String())
			if err != nil {
				return nil, fmt.Errorf("failed to download logs for job %s: %w", job.GetName(), err)
			}

			scanResult, err := parseScanResultFromLogs(logText)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: could not parse scan results for %s: %v\n", job.GetName(), err)
			} else {
				entry.Summary = scanResult.Summary
				entry.Vulnerabilities = scanResult.Vulnerabilities
			}
		}

		entries = append(entries, entry)
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

func downloadLogs(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch logs: HTTP %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}
	return string(body), nil
}

var timestampPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T[\d:.]+Z `)

func parseScanResultFromLogs(logText string) (*ScanResult, error) {
	lines := strings.Split(logText, "\n")

	var strippedLines []string
	for _, line := range lines {
		strippedLines = append(strippedLines, timestampPrefix.ReplaceAllString(line, ""))
	}
	text := strings.Join(strippedLines, "\n")

	start := strings.Index(text, `"result"`)
	if start < 0 {
		return nil, fmt.Errorf("no scan result JSON found in logs")
	}

	// Walk backwards to find the opening brace.
	braceStart := strings.LastIndex(text[:start], "{")
	if braceStart < 0 {
		return nil, fmt.Errorf("no scan result JSON found in logs")
	}

	// Walk forward from the opening brace to find the matching close.
	depth := 0
	for i := braceStart; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				jsonStr := text[braceStart : i+1]
				var wrapper scanResultWrapper
				if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
					return nil, fmt.Errorf("failed to parse scan result JSON: %w", err)
				}
				return &wrapper.Result, nil
			}
		}
	}

	return nil, fmt.Errorf("no scan result JSON found in logs")
}
