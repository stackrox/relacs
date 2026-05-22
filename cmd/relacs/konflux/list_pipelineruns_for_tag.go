package konflux

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
	"github.com/stackrox/relacs/pkg/utils"
)

type PipelinesForTagOutput struct {
	Organization string              `json:"organization" yaml:"organization"`
	Repository   string              `json:"repository" yaml:"repository"`
	Tag          string              `json:"tag" yaml:"tag"`
	PipelineRuns []PipelineRunOutput `json:"pipelineRuns" yaml:"pipelineRuns"`
}

func (p PipelinesForTagOutput) String() string {
	// Output this in a table format, like the github list-prs command.
	// If there are retest-comment or test-comment pipeline runs,
	// warn the user that we cannot guarantee that they are for the given tag.
	output := fmt.Sprintf("Input: %s/%s@%s\n", p.Organization, p.Repository, p.Tag)
	output += fmt.Sprintln("Pipeline runs:")
	output += renderPipelineRunsTable(p.PipelineRuns)
	return output
}

type PipelineRunOutput struct {
	Name              string    `json:"name" yaml:"name"`
	CreationTimestamp time.Time `json:"creationTimestamp" yaml:"creationTimestamp"`
	EventType         string    `json:"eventType" yaml:"eventType"`
	Succeeded         bool      `json:"succeeded" yaml:"succeeded"`
	Reason            string    `json:"reason" yaml:"reason"`
	URL               string    `json:"url" yaml:"url"`
}

func convertPipelineRunsToOutput(cfg *config.Config, pipelineRuns []kflx.PipelineRun) ([]PipelineRunOutput, error) {
	output := []PipelineRunOutput{}
	for _, pipelineRun := range pipelineRuns {
		converted, err := convertPipelineRunToOutput(cfg, pipelineRun)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pipeline run to output: %w", err)
		}
		output = append(output, converted)
	}
	return output, nil
}

func convertPipelineRunToOutput(cfg *config.Config, pipelineRun kflx.PipelineRun) (PipelineRunOutput, error) {
	var condition kflx.PipelineRunCondition
	for _, c := range pipelineRun.Status.Conditions {
		if c.Type == "Succeeded" {
			condition = c
			break
		}
	}
	if condition.Type == "" {
		return PipelineRunOutput{}, fmt.Errorf("no Succeeded condition found for pipeline run %s", pipelineRun.Metadata.Name)
	}
	return PipelineRunOutput{
		Name:              pipelineRun.Metadata.Name,
		CreationTimestamp: *pipelineRun.Metadata.CreationTimestamp,
		EventType:         pipelineRun.Metadata.Labels["pipelinesascode.tekton.dev/event-type"],
		Succeeded:         condition.Status == "True",
		Reason:            condition.Reason,
		URL:               utils.Application{Name: pipelineRun.Metadata.Labels["appstudio.openshift.io/application"]}.GetPipelineRunURL(cfg, pipelineRun.Metadata.Name),
	}, nil
}

var listPipelineRunsForTagCmd = &cobra.Command{
	Use:   "list-pipelineruns-for-tag",
	Short: "List the pipeline runs for a given Git tag",
	Long:  `Retrieve the pipeline runs associated with a specific Git tag`,
	RunE:  common.RunCommand(runListPipelineRunsForTag),
}

func init() {
	listPipelineRunsForTagCmd.Flags().String("org", "", "GitHub organization (optional)")
	listPipelineRunsForTagCmd.Flags().String("repo", "", "GitHub repository (optional)")
	listPipelineRunsForTagCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	listPipelineRunsForTagCmd.MarkFlagRequired("tag")
}

func runListPipelineRunsForTag(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	cfg := c.GetConfig()

	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	if org == "" {
		org = cfg.GitHubOrganization
	}
	repo, _ := cmd.Flags().GetString("repo")
	if repo == "" {
		repo = cfg.MainGitHubRepository
	}
	tag, _ := cmd.Flags().GetString("tag")

	// Get the pipelines for the tag
	pipelineRuns, err := konflux.ListPipelineRunsForTag(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines for tag: %w", err)
	}

	fmt.Fprintln(os.Stderr, "NOTE: This output includes pipeline runs for `/retest tag:<tag>` comments. Retest comments in the form `/retest branch:<release branch>` will not be included.")

	convertedPipelineRuns, err := convertPipelineRunsToOutput(cfg, pipelineRuns)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pipeline runs to output: %w", err)
	}

	return PipelinesForTagOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		PipelineRuns: convertedPipelineRuns,
	}, nil
}

func renderPipelineRunsTable(pipelineRuns []PipelineRunOutput) string {
	if len(pipelineRuns) == 0 {
		return "  No pipeline runs found."
	}

	var buf bytes.Buffer
	columnHeaders := getColumnHeaders()
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(columnHeaders))

	for _, pipelineRun := range pipelineRuns {
		row := getTableRow(pipelineRun)
		//nolint:errcheck
		table.Append(row)
	}

	//nolint:errcheck
	table.Render()
	return buf.String()
}

func getColumnHeaders() []string {
	return []string{"Name", "Creation Timestamp", "Event Type", "Succeeded?", "Reason", "URL"}
}

func getTableRow(pipelineRun PipelineRunOutput) []string {
	return []string{
		pipelineRun.Name,
		pipelineRun.CreationTimestamp.Format(time.RFC3339),
		pipelineRun.EventType,
		fmt.Sprintf("%t", pipelineRun.Succeeded),
		pipelineRun.Reason,
		pipelineRun.URL,
	}
}
