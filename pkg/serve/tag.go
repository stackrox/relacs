package serve

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
	"github.com/stackrox/relacs/pkg/konflux"
	"github.com/stackrox/relacs/pkg/utils"
)

type PipelineGroup struct {
	// Name is the pipeline name (the pipelinesascode.tekton.dev/original-prname
	// label); it is the grouping key and drives the milestone membership.
	Name string
	// Component is the group's display name: the
	// appstudio.openshift.io/component label, falling back to the pipeline
	// name when the runs carry no component label.
	Component    string
	PipelineRuns []PipelineRunInfo
	AllPassed    bool
	// Running reports whether the group's latest run is still running.
	Running bool
}

// PipelineMilestone sorts the pipeline run groups into the release
// milestones, in order: everything else, the operator bundle, and the
// operator index and custom snapshot.
type PipelineMilestone struct {
	Name   string
	Status string // "passed", "running", "failed"
	Groups []PipelineGroup
	// FailedPipelines lists the pipeline names of the groups whose latest
	// run failed; they are the candidates for a /konflux-retest.
	FailedPipelines []string
	// Retestable reports whether the milestone's retest button is enabled:
	// all groups of the previous milestone succeeded.
	Retestable bool
}

// pipelineMilestoneOrder lists the milestones in display order. A milestone
// with no group names collects all groups that belong to no other milestone.
var pipelineMilestoneOrder = []struct {
	Name  string
	Names []string
}{
	{Name: "Operands & Operator"},
	{Name: "Operator Bundle", Names: []string{"operator-bundle-on-push"}},
	{Name: "Custom Snapshot & Index Image", Names: []string{"operator-index-on-push", "create-custom-snapshot"}},
}

// groupPipelineMilestones sorts the groups into the release milestones and
// computes each milestone's combined status from the latest run of each of
// its groups. Empty milestones are dropped.
func groupPipelineMilestones(groups []PipelineGroup) []PipelineMilestone {
	membership := map[string]int{}
	for i, milestone := range pipelineMilestoneOrder {
		for _, name := range milestone.Names {
			membership[name] = i
		}
	}

	byMilestone := make([][]PipelineGroup, len(pipelineMilestoneOrder))
	for _, g := range groups {
		idx, ok := membership[g.Name]
		if !ok {
			idx = 0
		}
		byMilestone[idx] = append(byMilestone[idx], g)
	}

	var milestones []PipelineMilestone
	for i, milestone := range pipelineMilestoneOrder {
		if len(byMilestone[i]) == 0 {
			continue
		}
		// A milestone is retestable when all groups of the previous
		// milestone succeeded; an empty previous milestone counts as passed.
		retestable := i == 0 || milestoneStatus(byMilestone[i-1]) == "passed"

		var failedPipelines []string
		for _, g := range byMilestone[i] {
			if len(g.PipelineRuns) > 0 && !g.PipelineRuns[0].Succeeded && !g.PipelineRuns[0].Running {
				failedPipelines = append(failedPipelines, g.Name)
			}
		}

		milestones = append(milestones, PipelineMilestone{
			Name:            milestone.Name,
			Status:          milestoneStatus(byMilestone[i]),
			Groups:          byMilestone[i],
			FailedPipelines: failedPipelines,
			Retestable:      retestable,
		})
	}
	return milestones
}

// milestoneStatus is the combined status of a milestone's groups: "failed"
// when the latest run of any group did not succeed and is not running,
// "running" while any latest run is still running, and "passed" otherwise.
func milestoneStatus(groups []PipelineGroup) string {
	failed, running := false, false
	for _, g := range groups {
		if len(g.PipelineRuns) == 0 {
			continue
		}
		latest := g.PipelineRuns[0]
		switch {
		case latest.Succeeded:
		case latest.Running:
			running = true
		default:
			failed = true
		}
	}
	switch {
	case failed:
		return "failed"
	case running:
		return "running"
	default:
		return "passed"
	}
}

type PipelineRunInfo struct {
	Name              string
	CreationTimestamp time.Time
	EventType         string
	Succeeded         bool
	Running           bool
	Reason            string
	URL               string
}

type Violation struct {
	Image     string
	Arch      string
	Violation string
}

// ViolationGroup groups the conforma violations for one component version,
// e.g. all violations reported for images matching "scanner-4-10*".
type ViolationGroup struct {
	Name       string
	Archs      []string
	Violations []Violation
}

// ConformaRun is a conforma pipeline run that did not contribute to the
// violations: it is either still in progress or its report could not be
// retrieved or parsed (Reason).
type ConformaRun struct {
	Name   string
	URL    string
	Reason string
}

func convertConformaRuns(runs []konflux.ConformaRunInfo) []ConformaRun {
	out := make([]ConformaRun, len(runs))
	for i, run := range runs {
		out[i] = ConformaRun{Name: run.Name, URL: run.URL, Reason: run.Reason}
	}
	return out
}

// ConformaResults holds the results of one conforma scenario for a tag.
type ConformaResults struct {
	Violations     []Violation
	Groups         []ViolationGroup
	InProgressRuns []ConformaRun
	SkippedRuns    []ConformaRun
	Error          bool
	// Waiting reports that the checks are still running and nothing has been
	// verified yet.
	Waiting bool
}

func convertConformaResult(result konflux.ConformaViolations) ConformaResults {
	violations := convertViolations(result.Violations)
	return ConformaResults{
		Violations:     violations,
		Groups:         groupViolations(violations),
		InProgressRuns: convertConformaRuns(result.InProgressRuns),
		SkippedRuns:    convertConformaRuns(result.SkippedRuns),
		Waiting:        len(result.InProgressRuns) > 0 && len(result.Violations.Violations) == 0 && len(result.SkippedRuns) == 0,
	}
}

// ConformaTabData is the data for one conforma widget's tab content,
// identified by a prefix that namespaces its element ids.
type ConformaTabData struct {
	Prefix  string
	Results ConformaResults
}

func (t TagData) DevTabs() ConformaTabData {
	return ConformaTabData{Prefix: "dev", Results: t.ConformaDev}
}

func (t TagData) PreviewTabs() ConformaTabData {
	return ConformaTabData{Prefix: "preview", Results: t.ConformaReleasePreview}
}

type TagData struct {
	Tag  string
	Repo string

	OverallStatus      string // "passed", "failed", "warning"
	AllPipelinesPassed bool
	// PipelinesRunning reports whether build pipeline runs are still running
	// and none failed; it drives the Pipeline Runs widget's running status.
	PipelinesRunning bool
	HasViolations    bool

	CommitSHA           string
	CommitURL           string
	SnapshotName        string
	SnapshotURL         string
	OperatorBundleImage string
	OperatorIndexImage  string

	PipelineGroups     []PipelineGroup
	PipelineMilestones []PipelineMilestone
	PipelinesError     bool

	// ConformaDev holds the acs-conforma-dev results, ConformaReleasePreview
	// the acs-conforma-prod-like results (main repository only).
	ConformaDev            ConformaResults
	ConformaReleasePreview ConformaResults

	Errors []string
}

func computeTagOverallStatus(data *TagData, isMainRepo bool) {
	data.AllPipelinesPassed = len(data.PipelineGroups) > 0
	for _, g := range data.PipelineGroups {
		if !g.AllPassed {
			data.AllPipelinesPassed = false
			break
		}
	}
	data.HasViolations = len(data.ConformaDev.Violations) > 0

	// The pipeline widget's running status is independent of the overall
	// status below, so compute it before the error early-return.
	buildFailed, buildRunning := false, false
	for _, g := range data.PipelineGroups {
		if len(g.PipelineRuns) == 0 {
			continue
		}
		latest := g.PipelineRuns[0]
		if latest.Succeeded {
			continue
		}
		if latest.Running {
			buildRunning = true
		} else {
			buildFailed = true
		}
	}
	data.PipelinesRunning = buildRunning && !buildFailed

	// The overall status is "error" when any data could not be retrieved,
	// "failed" when prod-like violations were found, prod-like runs were
	// skipped or the latest run of a build pipeline run group did not
	// succeed, "waiting" while conforma or build pipeline runs are still in
	// progress, and "green" only when every build pipeline run group's latest
	// run succeeded and, for the main repository, the custom snapshot and the
	// operator bundle and index images are present and there are no prod-like
	// violations. For other repositories, green build pipeline runs are
	// sufficient.
	if data.PipelinesError || data.ConformaDev.Error ||
		(isMainRepo && (data.ConformaReleasePreview.Error || data.SnapshotName == "" ||
			data.OperatorBundleImage == "" || data.OperatorIndexImage == "")) {
		data.OverallStatus = "error"
		return
	}

	failed := buildFailed || (isMainRepo && (len(data.ConformaReleasePreview.Violations) > 0 || len(data.ConformaReleasePreview.SkippedRuns) > 0))
	waiting := buildRunning || (isMainRepo && len(data.ConformaReleasePreview.InProgressRuns) > 0)

	switch {
	case failed:
		data.OverallStatus = "failed"
	case waiting:
		data.OverallStatus = "waiting"
	default:
		data.OverallStatus = "green"
	}
}

// violationImageGroup extracts the "<component>-<major>-<minor>" prefix of an
// image name, e.g. "scanner-4-10" from "scanner-4-10-s390x". Components may
// themselves contain hyphens ("scanner-db-4-10"). Images without a numeric
// major-minor version form their own group.
func violationImageGroup(image string) string {
	parts := strings.Split(image, "-")
	for i := 1; i < len(parts)-1; i++ {
		if isNumeric(parts[i]) && isNumeric(parts[i+1]) {
			return strings.Join(parts[:i+2], "-")
		}
	}
	return image
}

// violationImageArch extracts the architecture of an image relative to its
// group prefix. Image names look like
// "operator-5-0-sha256:<digest>-ppc64le", so the remainder after the group is
// "<digest>[-<arch>]": the digest contains no hyphens, and the arch may
// ("x86-64"). Images without an arch suffix are the multi-arch image.
func violationImageArch(image, group string) string {
	remainder := strings.TrimPrefix(strings.TrimPrefix(image, group), "-")
	_, arch, found := strings.Cut(remainder, "-")
	if !found || arch == "" {
		return "multi-arch"
	}
	return arch
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// groupViolations groups violations by their image's component version prefix
// and annotates each violation with its image's architecture. Groups are
// sorted by name, and each group lists its distinct architectures sorted.
func groupViolations(violations []Violation) []ViolationGroup {
	groupMap := map[string][]Violation{}
	var groupOrder []string
	for _, v := range violations {
		name := violationImageGroup(v.Image)
		if _, exists := groupMap[name]; !exists {
			groupOrder = append(groupOrder, name)
		}
		groupMap[name] = append(groupMap[name], v)
	}
	sort.Strings(groupOrder)

	groups := make([]ViolationGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		seen := map[string]bool{}
		var archs []string
		for i := range groupMap[name] {
			groupMap[name][i].Arch = violationImageArch(groupMap[name][i].Image, name)
			if !seen[groupMap[name][i].Arch] {
				seen[groupMap[name][i].Arch] = true
				archs = append(archs, groupMap[name][i].Arch)
			}
		}
		sort.Strings(archs)
		groups = append(groups, ViolationGroup{
			Name:       name,
			Archs:      archs,
			Violations: groupMap[name],
		})
	}
	return groups
}

// ComputeGroupStatus sets AllPassed on each group based on the most recently
// created PipelineRun. Runs within each group are sorted newest-first.
func ComputeGroupStatus(groups []PipelineGroup) {
	for i := range groups {
		runs := groups[i].PipelineRuns
		sort.Slice(runs, func(a, b int) bool {
			return runs[a].CreationTimestamp.After(runs[b].CreationTimestamp)
		})
		groups[i].PipelineRuns = runs
		if len(runs) > 0 {
			groups[i].AllPassed = runs[0].Succeeded
			groups[i].Running = !runs[0].Succeeded && runs[0].Running
		}
	}
}

func convertPipelineRun(plr kflx.PipelineRun) (PipelineRunInfo, error) {
	var condition kflx.PipelineRunCondition
	for _, c := range plr.Status.Conditions {
		if c.Type == "Succeeded" {
			condition = c
			break
		}
	}
	if condition.Type == "" {
		return PipelineRunInfo{}, fmt.Errorf("no Succeeded condition found for pipeline run %s", plr.Metadata.Name)
	}
	return PipelineRunInfo{
		Name:              plr.Metadata.Name,
		CreationTimestamp: *plr.Metadata.CreationTimestamp,
		EventType:         plr.Metadata.Labels["pipelinesascode.tekton.dev/event-type"],
		Succeeded:         condition.Status == "True",
		Running:           condition.Status == "Unknown",
		Reason:            condition.Reason,
		URL:               utils.Application{Name: plr.Metadata.Labels["appstudio.openshift.io/application"]}.GetPipelineRunURL(plr.Metadata.Name),
	}, nil
}

func groupPipelineRuns(plrs []kflx.PipelineRun) ([]PipelineGroup, error) {
	groupMap := map[string][]PipelineRunInfo{}
	components := map[string]string{}
	var groupOrder []string

	for _, plr := range plrs {
		info, err := convertPipelineRun(plr)
		if err != nil {
			return nil, err
		}
		name := plr.Metadata.Labels["pipelinesascode.tekton.dev/original-prname"]
		if name == "" {
			name = plr.Metadata.Name
		}
		if _, exists := groupMap[name]; !exists {
			groupOrder = append(groupOrder, name)
			components[name] = plr.Metadata.Labels["appstudio.openshift.io/component"]
		}
		groupMap[name] = append(groupMap[name], info)
	}

	sort.Strings(groupOrder)
	groups := make([]PipelineGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		group := PipelineGroup{
			Name:         name,
			Component:    name,
			PipelineRuns: groupMap[name],
		}
		if component := components[name]; component != "" {
			group.Component = component
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func convertViolations(v konflux.Violations) []Violation {
	out := make([]Violation, len(v.Violations))
	for i, viol := range v.Violations {
		out[i] = Violation{Image: viol.Image, Violation: viol.Violation}
	}
	return out
}

func FetchTagData(ctx context.Context, c *client.Client, repo, tag string) TagData {
	org := config.GitHubOrganization
	appPrefix := config.KonfluxApplicationPrefix
	isMainRepo := repo == config.MainGitHubRepository

	data := TagData{
		Tag:  tag,
		Repo: repo,
	}

	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		commit   github.CommitInfo
		snap     kflx.Snapshot
		plrs     []kflx.PipelineRun
		indexImg konflux.OperatorIndexImage
	)

	addError := func(msg string) {
		mu.Lock()
		data.Errors = append(data.Errors, msg)
		mu.Unlock()
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		var err error
		commit, err = github.GetCommitForTag(ctx, c, org, repo, tag)
		if err != nil {
			addError(fmt.Sprintf("Failed to get commit for tag: %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		plrs, err = konflux.ListPipelineRunsForTag(ctx, c, org, repo, tag)
		if err != nil {
			addError(fmt.Sprintf("Failed to list pipeline runs: %v", err))
			mu.Lock()
			data.PipelinesError = true
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		result, violErr := konflux.ListConformaViolations(ctx, c, org, repo, tag, konflux.ConformaScenarioDev)
		if violErr != nil {
			addError(fmt.Sprintf("Failed to list conforma violations: %v", violErr))
			mu.Lock()
			data.ConformaDev.Error = true
			mu.Unlock()
			return
		}
		mu.Lock()
		data.ConformaDev = convertConformaResult(result)
		mu.Unlock()
	}()

	if isMainRepo {
		wg.Add(3)

		go func() {
			defer wg.Done()
			result, violErr := konflux.ListConformaViolations(ctx, c, org, repo, tag, konflux.ConformaScenarioProdLike)
			if violErr != nil {
				addError(fmt.Sprintf("Failed to list conforma release preview violations: %v", violErr))
				mu.Lock()
				data.ConformaReleasePreview.Error = true
				mu.Unlock()
				return
			}
			mu.Lock()
			data.ConformaReleasePreview = convertConformaResult(result)
			mu.Unlock()
		}()

		go func() {
			defer wg.Done()
			var err error
			indexImg, err = konflux.GetOperatorIndexImage(ctx, c, org, repo, tag)
			if err != nil {
				addError(fmt.Sprintf("Failed to get operator index image: %v", err))
			}
		}()

		go func() {
			defer wg.Done()
			var err error
			snap, err = konflux.GetCustomSnapshot(ctx, c, appPrefix, org, repo, tag)
			if err != nil {
				addError(fmt.Sprintf("Failed to get custom snapshot: %v", err))
			}
		}()
	}

	wg.Wait()

	if indexImg.ImageURL != "" {
		data.OperatorIndexImage = indexImg.ImageURL + "@" + indexImg.ImageDigest
	}

	if commit.SHA != "" {
		data.CommitSHA = commit.SHA
		data.CommitURL = fmt.Sprintf("https://github.com/%s/%s/commit/%s", org, repo, commit.SHA)
	}

	if snap.Metadata.Name != "" {
		data.SnapshotName = snap.Metadata.Name
		snapshotURL, err := snap.GetURL()
		if err == nil {
			data.SnapshotURL = snapshotURL
		}

		bundleImage, err := snap.GetOperatorBundleImage(tag)
		if err != nil {
			addError(fmt.Sprintf("Failed to get operator bundle image: %v", err))
		} else {
			data.OperatorBundleImage = bundleImage
		}
	}

	if plrs != nil {
		if len(plrs) == 0 {
			addError("No pipeline runs found for tag")
			data.PipelinesError = true
		} else {
			groups, err := groupPipelineRuns(plrs)
			if err != nil {
				addError(fmt.Sprintf("Failed to process pipeline runs: %v", err))
				data.PipelinesError = true
			} else {
				ComputeGroupStatus(groups)
				data.PipelineGroups = groups
				data.PipelineMilestones = groupPipelineMilestones(groups)
			}
		}
	}

	computeTagOverallStatus(&data, isMainRepo)

	return data
}
