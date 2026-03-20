package konflux

import (
	"context"
	"fmt"
	"sort"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func GetRelease(ctx context.Context, c *client.Client, name string) (kflx.Release, error) {
	release, err := c.GetKonflux().GetRelease(name)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to retrieve release: %w", err)
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

func RenderStageRelease(snapshot kflx.Snapshot, issues []string, cves []jira.CVE, attempt int) (kflx.Release, error) {
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
				"acs-release-version": version.String(),
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
		// Pass base component name; NewCVEInformation will append version via GetResourceWithVersion
		cveInfo, err := kflx.NewCVEInformation(application, cve.Code, cve.ComponentName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CVE: %w", err)
		}

		err = validateComponentForCVE(snapshot, cveInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to validate component %s for %s: %w", cve.ComponentName, cve.Code, err)
		}

		cveList = append(cveList, cveInfo)
	}
	return cveList, nil
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

func renderReferences(version utils.Version) []string {
	// References should include the release notes found on docs.redhat.com.
	docsRedhatURL := fmt.Sprintf(
		"https://docs.redhat.com/en/documentation/red_hat_advanced_cluster_security_for_kubernetes/%d.%d/html-single/release_notes/index",
		version.Major, version.Minor)
	// For patch releases, we need to append the anchor that points directly to the release notes for that patch release.
	if version.IsPatchRelease() {
		docsRedhatURL += fmt.Sprintf("#about-this-release-%d%d%d_release-notes-%d%d",
			version.Major, version.Minor, version.Patch, version.Major, version.Minor)
	}
	return []string{docsRedhatURL}
}

func PromoteReleaseToProd(release kflx.Release, snapshot kflx.Snapshot) (kflx.Release, error) {
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

	prodRelease, err := renderProdRelease(release, application, version, snapshot)
	if err != nil {
		return kflx.Release{}, fmt.Errorf("failed to render prod release: %w", err)
	}
	return prodRelease, nil
}

func renderProdRelease(release kflx.Release, application utils.Application, version utils.Version, snapshot kflx.Snapshot) (kflx.Release, error) {
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
			"acs-release-version": version.String(),
			"acs-stage-release":   release.Metadata.Name,
		},
	}
	prodRelease.Spec.ReleasePlan = releasePlan
	prodRelease.Spec.Snapshot = snapshot.Metadata.Name
	prodRelease.Status = nil
	return prodRelease, nil
}
