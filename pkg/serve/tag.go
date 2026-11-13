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
	Name         string
	PipelineRuns []PipelineRunInfo
	AllPassed    bool
}

type PipelineRunInfo struct {
	Name              string
	CreationTimestamp time.Time
	EventType         string
	Succeeded         bool
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

type TagData struct {
	Tag  string
	Repo string

	OverallStatus      string // "passed", "failed", "warning"
	AllPipelinesPassed bool
	HasViolations      bool

	CommitSHA           string
	CommitURL           string
	SnapshotName        string
	SnapshotURL         string
	OperatorBundleImage string
	OperatorIndexImage  string

	PipelineGroups  []PipelineGroup
	PipelinesError  bool
	Violations      []Violation
	ViolationGroups []ViolationGroup
	ViolationsError bool
	Errors          []string
}

func computeTagOverallStatus(data *TagData) {
	data.AllPipelinesPassed = len(data.PipelineGroups) > 0
	for _, g := range data.PipelineGroups {
		if !g.AllPassed {
			data.AllPipelinesPassed = false
			break
		}
	}
	data.HasViolations = len(data.Violations) > 0

	if data.PipelinesError || data.ViolationsError || !data.AllPipelinesPassed {
		data.OverallStatus = "failed"
	} else if data.HasViolations {
		data.OverallStatus = "warning"
	} else {
		data.OverallStatus = "passed"
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
		Reason:            condition.Reason,
		URL:               utils.Application{Name: plr.Metadata.Labels["appstudio.openshift.io/application"]}.GetPipelineRunURL(plr.Metadata.Name),
	}, nil
}

func groupPipelineRuns(plrs []kflx.PipelineRun) ([]PipelineGroup, error) {
	groupMap := map[string][]PipelineRunInfo{}
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
		}
		groupMap[name] = append(groupMap[name], info)
	}

	sort.Strings(groupOrder)
	groups := make([]PipelineGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		groups = append(groups, PipelineGroup{
			Name:         name,
			PipelineRuns: groupMap[name],
		})
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
		viols    konflux.Violations
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
			data.ViolationsError = true
			mu.Unlock()
			return
		}
		viols = result.Violations
	}()

	if isMainRepo {
		wg.Add(2)

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
			}
		}
	}

	data.Violations = convertViolations(viols)
	data.ViolationGroups = groupViolations(data.Violations)

	computeTagOverallStatus(&data)

	return data
}
