package konflux

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"

	versionPkg "github.com/stackrox/relacs/cmd/relacs/version"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

const (
	BaseAnnotation               = "relacs.rhacs"
	AuthorAnnotation             = BaseAnnotation + "/" + "author"
	GeneratedTimestampAnnotation = BaseAnnotation + "/" + "generated-timestamp"
	VersionAnnotation            = BaseAnnotation + "/" + "version"
	ACSReleaseVersionLabel       = "acs-release-version"
	ACSStageReleaseLabel         = "acs-stage-release"
)

func GetRelease(ctx context.Context, c *client.Client, name string) (kflx.Release, error) {
	release, err := c.GetKonflux().GetRelease(ctx, name)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to retrieve release: %w", err)
	}
	return release, nil
}

// isReleaseOrRerunSucceeded checks if a release or any of its reruns succeeded.
// Reruns follow the naming pattern: {original-name}-rerun-{5chars}
// For example: acs-4-11--4-11-7--cdn-stage-s390x-1-rerun-wtz8k
func isReleaseOrRerunSucceeded(ctx context.Context, c *client.Client, release kflx.Release) (bool, string, error) {
	// First check if the original release succeeded
	if release.IsSucceeded() {
		return true, release.Metadata.Name, nil
	}

	// If not, look for successful reruns
	// Get the version label to query for related releases
	versionLabel, ok := release.Metadata.Labels[ACSReleaseVersionLabel]
	if !ok {
		return false, "", fmt.Errorf("%s not found in release labels", ACSReleaseVersionLabel)
	}

	// Query all releases with the same version label
	labels := []string{fmt.Sprintf("%s=%s", ACSReleaseVersionLabel, versionLabel)}

	// Add arch label for CDN releases if present
	if archLabel, ok := release.Metadata.Labels[ACSCdnArchLabel]; ok {
		labels = append(labels, fmt.Sprintf("%s=%s", ACSCdnArchLabel, archLabel))
	}

	releases, err := c.GetKonflux().GetReleases(ctx, labels)
	if err != nil {
		return false, "", fmt.Errorf("failed to query releases: %w", err)
	}

	// Pattern for reruns: {original-name}-rerun-{5 alphanumeric chars}
	rerunPattern := regexp.MustCompile(fmt.Sprintf("^%s-rerun-[a-z0-9]{5}$", regexp.QuoteMeta(release.Metadata.Name)))

	var successfulReruns []kflx.Release
	for _, r := range releases {
		if rerunPattern.MatchString(r.Metadata.Name) && r.IsSucceeded() {
			successfulReruns = append(successfulReruns, r)
		}
	}

	if len(successfulReruns) == 0 {
		return false, "", nil
	}

	// Sort by creation timestamp to get the most recent successful rerun
	sort.Slice(successfulReruns, func(i, j int) bool {
		if successfulReruns[i].Metadata.CreationTimestamp == nil {
			return false
		}
		if successfulReruns[j].Metadata.CreationTimestamp == nil {
			return true
		}
		return successfulReruns[i].Metadata.CreationTimestamp.Before(*successfulReruns[j].Metadata.CreationTimestamp)
	})

	mostRecentRerun := successfulReruns[len(successfulReruns)-1]
	return true, mostRecentRerun.Metadata.Name, nil
}

// EnsureSuccessfulRelease verifies that the given release or one of its reruns succeeded,
// and returns the successful release object. If the original release failed but a rerun
// succeeded, it fetches and returns the successful rerun, logging a note to stderr.
func EnsureSuccessfulRelease(ctx context.Context, c *client.Client, release kflx.Release, originalName string) (kflx.Release, error) {
	succeeded, successfulReleaseName, err := isReleaseOrRerunSucceeded(ctx, c, release)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to check if release %s succeeded: %w", originalName, err)
	}
	if !succeeded {
		return kflx.Release{}, fmt.Errorf("stage release %s and its reruns did not succeed, cannot promote to prod", originalName)
	}

	if successfulReleaseName != originalName {
		successfulRelease, err := GetRelease(ctx, c, successfulReleaseName)
		if err != nil {
			return kflx.Release{}, fmt.Errorf("failed to get successful rerun %s: %w", successfulReleaseName, err)
		}
		fmt.Fprintf(os.Stderr, "Note: Using successful rerun %s for promotion of %s\n", successfulReleaseName, originalName)
		return successfulRelease, nil
	}

	return release, nil
}

func GetOperatorBundleImageForRelease(ctx context.Context, c *client.Client, release string) (string, error) {
	releaseObj, err := GetRelease(ctx, c, release)
	if err != nil {
		return "", fmt.Errorf("failed to get release: %w", err)
	}
	advisory, err := kflx.GetAdvisoryFromRelease(releaseObj)
	if err != nil {
		return "", fmt.Errorf("failed to get advisory for release: %w", err)
	}
	application, ok := releaseObj.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return "", fmt.Errorf("application not found in release labels")
	}
	operatorBundleImage, err := advisory.GetOperatorBundleImage(application)
	if err != nil {
		return "", fmt.Errorf("failed to get operator bundle image for release %s and application %s: %w", release, application, err)
	}
	return operatorBundleImage, nil
}

func RenderStageRelease(snapshot kflx.Snapshot, issues []string, cves []jira.CVE, attempt int, author string) (kflx.Release, error) {
	application := utils.Application{Name: snapshot.Metadata.Labels["appstudio.openshift.io/application"]}
	version, err := utils.NewVersion(snapshot.Metadata.Labels["acs-product-version"])
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to get version: %w", err)
	}

	releasePlan, err := application.GetResourceWithVersion("acs-stage")
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to get release plan: %w", err)
	}

	renderedCVEs, err := renderCVEs(snapshot, application, cves)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to render CVEs: %w", err)
	}
	releaseType := determineReleaseType(renderedCVEs, version)
	highestSeverity := determineHighestSeverity(renderedCVEs)
	synosis, err := renderSynopsis(releaseType, version, issues, highestSeverity)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to render synopsis: %w", err)
	}

	renderedIssues := renderIssues(issues)

	release := kflx.Release{
		Kind:       "Release",
		APIVersion: "appstudio.redhat.com/v1alpha1",
		Metadata: kflx.Metadata{
			Name:      renderReleaseName(application, version, "stage", attempt),
			Namespace: config.KonfluxTenant,
			Labels: map[string]string{
				ACSReleaseVersionLabel: version.String(),
			},
			Annotations: map[string]string{
				AuthorAnnotation:             author,
				GeneratedTimestampAnnotation: time.Now().Format(time.RFC3339),
				VersionAnnotation:            versionPkg.GetVersion(),
			},
		},
		Spec: kflx.ReleaseSpec{
			ReleasePlan: releasePlan,
			Snapshot:    snapshot.Metadata.Name,
			Data: kflx.ReleaseData{
				Mapping: kflx.ReleaseMapping{
					Defaults: kflx.ReleaseMappingDefaults{
						Tags: version.RenderTags(),
					},
				},
				ReleaseNotes: kflx.ReleaseNotes{
					Type:       releaseType.String(),
					Synopsis:   synosis,
					References: renderReferences(version),
					Issues: kflx.ReleaseIssues{
						Fixed: renderedIssues,
					},
					CVEs: renderedCVEs,
				},
			},
		},
	}
	return release, nil
}

func renderCVEs(snapshot kflx.Snapshot, application utils.Application, cves []jira.CVE) ([]kflx.CVEInformation, error) {
	cveList := []kflx.CVEInformation{}

	for _, cve := range cves {
		// Pass parsed CVE data; NewCVEInformation will append version via GetResourceWithVersion
		cveInfo, err := kflx.NewCVEInformation(application, cve.Key, cve.Year, cve.ComponentName, cve.Severity)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CVE: %w", err)
		}

		err = validateComponentForCVE(snapshot, cveInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to validate component %s for %s: %w", cve.ComponentName, cve.Key, err)
		}

		cveList = append(cveList, cveInfo)
	}

	sort.Slice(cveList, cveLess(cveList))
	return cveList, nil
}

// cveLess returns a sort comparator that orders by component, then year, then numeric CVE ID.
func cveLess(cveList []kflx.CVEInformation) func(i, j int) bool {
	return func(i, j int) bool {
		if cveList[i].Component != cveList[j].Component {
			return cveList[i].Component < cveList[j].Component
		}
		if cveList[i].Year != cveList[j].Year {
			return cveList[i].Year < cveList[j].Year
		}
		return cveNumericID(cveList[i].Key) < cveNumericID(cveList[j].Key)
	}
}

func cveNumericID(key string) int {
	lastDash := strings.LastIndex(key, "-")
	if lastDash < 0 {
		return 0
	}
	id, _ := strconv.Atoi(key[lastDash+1:])
	return id
}

func validateComponentForCVE(snapshot kflx.Snapshot, cveInfo kflx.CVEInformation) error {
	for _, component := range snapshot.Spec.Components {
		if component.Name == cveInfo.Component {
			return nil
		}
	}
	return fmt.Errorf("component not found in snapshot: %s", cveInfo.Component)
}

func determineReleaseType(cves []kflx.CVEInformation, version utils.Version) kflx.ReleaseType {
	// Major releases are always RHEA advisories because they are treated as new product releases.
	if version.IsMajorRelease() {
		return kflx.ReleaseType_RHEA
	}
	if len(cves) > 0 {
		return kflx.ReleaseType_RHSA
	}
	if version.IsPatchRelease() {
		return kflx.ReleaseType_RHBA
	}
	return kflx.ReleaseType_RHEA
}

func determineHighestSeverity(cves []kflx.CVEInformation) kflx.Severity {
	highestSeverity := kflx.Severity_Low
	for _, cve := range cves {
		if cve.Severity > highestSeverity {
			highestSeverity = cve.Severity
		}
	}
	return highestSeverity
}

func renderSynopsis(releaseType kflx.ReleaseType, version utils.Version, issues []string, highestSeverity kflx.Severity) (string, error) {
	updateType := ""
	switch releaseType {
	case kflx.ReleaseType_RHSA:
		updateType = "security"
		if version.IsPatchRelease() {
			if len(issues) > 0 {
				updateType += " and bug fix"
			}
		} else {
			updateType += " and enhancement"
		}
	case kflx.ReleaseType_RHBA:
		updateType = "bug fix"
	case kflx.ReleaseType_RHEA:
		updateType = "enhancement"
	default:
		return "", fmt.Errorf("invalid release type: %v", releaseType)
	}

	synosis := fmt.Sprintf("RHACS %s %s update", version, updateType)
	if releaseType == kflx.ReleaseType_RHSA {
		synosis = fmt.Sprintf("%s: %s", highestSeverity, synosis)
	}
	return synosis, nil
}

func renderIssues(issues []string) []map[string]string {
	renderedIssues := []map[string]string{}
	sort.Strings(issues)
	seen := make(map[string]struct{})
	for _, issue := range issues {
		if _, ok := seen[issue]; !ok {
			seen[issue] = struct{}{}
			renderedIssues = append(renderedIssues, map[string]string{
				"id":     issue,
				"source": config.IssueSource,
			})
		}
	}
	return renderedIssues
}

func renderReleaseName(application utils.Application, version utils.Version, target string, attempt int) string {
	return fmt.Sprintf("%s--%d-%d-%d--%s-%d", application.Name, version.Major, version.Minor, version.Patch, target, attempt)
}

func renderDocsURL(version utils.Version) string {
	// References should include the release notes found on docs.redhat.com.
	docsRedhatURL := fmt.Sprintf(config.RedHatDocsReleaseNotesURLTemplate, version.Major, version.Minor)
	// For patch releases, we need to append the anchor that points directly to the release notes for that patch release.
	if version.IsPatchRelease() {
		docsRedhatURL += fmt.Sprintf("#about-this-release-%d%d%d_release-notes-%d%d",
			version.Major, version.Minor, version.Patch, version.Major, version.Minor)
	}
	return docsRedhatURL
}

func renderReferences(version utils.Version) []string {
	return []string{renderDocsURL(version)}
}

func GetValidTagFromStageRelease(stageRelease kflx.Release) (string, error) {
	tag, ok := stageRelease.Metadata.Labels[ACSReleaseVersionLabel]
	if !ok {
		return "", fmt.Errorf("%s not found in stage release labels", ACSReleaseVersionLabel)
	}

	err := validateProdTagFormat(tag)
	if err != nil {
		return "", fmt.Errorf("failed to validate prod tag format: %w", err)
	}

	return tag, nil
}

func validateProdTagFormat(tag string) error {
	tagPattern := regexp.MustCompile(`^[1-9]\d*\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$`)
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("tag must be in x.y.z format (e.g., 1.2.3), got: %s", tag)
	}
	return nil
}

func PromoteReleaseToProd(release kflx.Release, snapshot kflx.Snapshot, author string) (kflx.Release, error) {
	applicationName, ok := release.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return kflx.Release{}, fmt.Errorf("appstudio.openshift.io/application not found in release labels")
	}
	application := utils.Application{Name: applicationName}

	applicationInSnapshot := utils.Application{Name: snapshot.Metadata.Labels["appstudio.openshift.io/application"]}
	if application.Name != applicationInSnapshot.Name {
		return kflx.Release{}, fmt.Errorf("application name mismatch between stage release and snapshot application name: %s != %s", application.Name, applicationInSnapshot.Name)
	}

	version, err := utils.NewVersion(snapshot.Metadata.Labels["acs-product-version"])
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to get version: %w", err)
	}

	prodRelease, err := renderProdRelease(release, application, version, snapshot, author)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to render prod release: %w", err)
	}
	return prodRelease, nil
}

func renderProdRelease(release kflx.Release, application utils.Application, version utils.Version, snapshot kflx.Snapshot, author string) (kflx.Release, error) {
	releasePlan, err := application.GetResourceWithVersion("acs-prod")
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to get release plan: %w", err)
	}

	// We expect there to be only one attempt to create a production Release.
	// Anything else should not be handled by relacs, but the release engineer.
	attempt := 1

	prodRelease, err := release.DeepCopy()
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to deep copy release: %w", err)
	}
	prodRelease.Metadata = kflx.Metadata{
		Name:      renderReleaseName(application, version, "prod", attempt),
		Namespace: config.KonfluxTenant,
		Labels: map[string]string{
			ACSReleaseVersionLabel: version.String(),
			ACSStageReleaseLabel:   release.Metadata.Name,
		},
		Annotations: map[string]string{
			AuthorAnnotation:             author,
			GeneratedTimestampAnnotation: time.Now().Format(time.RFC3339),
			VersionAnnotation:            versionPkg.GetVersion(),
		},
	}
	prodRelease.Spec.ReleasePlan = releasePlan
	prodRelease.Spec.Snapshot = snapshot.Metadata.Name
	prodRelease.Status = nil
	return prodRelease, nil
}
