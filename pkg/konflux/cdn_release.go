package konflux

import (
	"context"
	"fmt"
	"time"

	versionPkg "github.com/stackrox/relacs/cmd/relacs/version"
	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"
)

const ACSCdnArchLabel = "acs-cdn-arch"

func IsValidCDNArch(arch string) bool {
	_, ok := config.CDNArchProductIDs[arch]
	return ok
}

func RenderCdnRelease(snapshot kflx.Snapshot, arch, env string, attempt int, author string) (kflx.CdnRelease, error) {
	application := utils.Application{Name: snapshot.Metadata.Labels["appstudio.openshift.io/application"]}
	version, err := utils.NewVersion(snapshot.Metadata.Labels["acs-product-version"])
	if err != nil {
		return kflx.CdnRelease{}, fmt.Errorf("failed to get version from snapshot: %w", err)
	}

	releasePlan := renderCdnReleasePlanName(arch, env, version)
	name := renderCdnReleaseName(application, version, arch, env, attempt)

	return kflx.CdnRelease{
		Kind:       "Release",
		APIVersion: "appstudio.redhat.com/v1alpha1",
		Metadata: kflx.Metadata{
			Name:      name,
			Namespace: config.KonfluxTenant,
			Labels: map[string]string{
				ACSReleaseVersionLabel: version.String(),
				ACSCdnArchLabel:        arch,
				KonfluxAuthorLabel:     author,
			},
			Annotations: map[string]string{
				AuthorAnnotation:             author,
				GeneratedTimestampAnnotation: time.Now().Format(time.RFC3339),
				VersionAnnotation:            versionPkg.GetVersion(),
			},
		},
		Spec: kflx.CdnReleaseSpec{
			ReleasePlan: releasePlan,
			Snapshot:    snapshot.Metadata.Name,
			Data: kflx.CdnReleaseData{
				Mapping: kflx.CdnReleaseMapping{
					Components: []kflx.CdnReleaseComponent{
						{
							Name: fmt.Sprintf("main-%d-%d", version.Major, version.Minor),
							Staged: kflx.CdnReleaseStagedVersion{
								Version: version.String(),
							},
						},
					},
				},
				ReleaseNotes: kflx.CdnReleaseNotes{
					Type:       renderCdnReleaseType(version),
					Synopsis:   renderCdnSynopsis(version),
					References: renderCdnReferences(version, arch),
				},
			},
		},
	}, nil
}

func getArchFromCdnStageRelease(stageRelease kflx.Release) (string, error) {
	arch, ok := stageRelease.Metadata.Labels[ACSCdnArchLabel]
	if !ok {
		return "", fmt.Errorf("%s not found in stage release labels", ACSCdnArchLabel)
	}
	if !IsValidCDNArch(arch) {
		return "", fmt.Errorf("invalid arch %q in stage release labels, must be one of: %v", arch, config.GetCDNArchitectures())
	}
	return arch, nil
}

func PromoteCdnReleaseToProd(ctx context.Context, c *client.Client, stageRelease kflx.Release, applicationPrefix, org, repo, author string) (kflx.CdnRelease, error) {
	arch, err := getArchFromCdnStageRelease(stageRelease)
	if err != nil {
		return kflx.CdnRelease{}, fmt.Errorf("failed to get arch from stage release: %w", err)
	}

	tag, ok := stageRelease.Metadata.Labels[ACSReleaseVersionLabel]
	if !ok {
		return kflx.CdnRelease{}, fmt.Errorf("%s not found in stage release labels", ACSReleaseVersionLabel)
	}

	snapshot, err := GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return kflx.CdnRelease{}, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	return RenderCdnRelease(snapshot, arch, "prod", 1, author)
}

func renderCdnReleasePlanName(arch, env string, version utils.Version) string {
	return fmt.Sprintf("acs-cdn-%s-%s-%d-%d", arch, env, version.Major, version.Minor)
}

func renderCdnReleaseName(application utils.Application, version utils.Version, arch, env string, attempt int) string {
	target := fmt.Sprintf("cdn-%s-%s", env, arch)
	return fmt.Sprintf("%s--%d-%d-%d--%s-%d", application.Name, version.Major, version.Minor, version.Patch, target, attempt)
}

func renderCdnReleaseType(version utils.Version) string {
	if version.IsPatchRelease() {
		return "RHBA"
	}
	return "RHEA"
}

func renderCdnSynopsis(version utils.Version) string {
	return fmt.Sprintf("RHACS %s CLI update", version)
}

func renderCdnReferences(version utils.Version, arch string) []string {
	productID := config.CDNArchProductIDs[arch]
	return []string{
		config.RedHatDownloadsBaseURL + productID,
		renderDocsURL(version),
	}
}
