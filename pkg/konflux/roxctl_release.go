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

const ACSRoxctlArchLabel = "acs-roxctl-arch"

func RenderRoxctlRelease(snapshot kflx.Snapshot, arch, env string, attempt int, author string) (kflx.RoxctlRelease, error) {
	application := utils.Application{Name: snapshot.Metadata.Labels["appstudio.openshift.io/application"]}
	version, err := utils.NewVersion(snapshot.Metadata.Labels["acs-product-version"])
	if err != nil {
		return kflx.RoxctlRelease{}, fmt.Errorf("failed to get version from snapshot: %w", err)
	}

	releasePlan := renderRoxctlReleasePlanName(arch, env, version)
	name := renderRoxctlReleaseName(application, version, arch, env, attempt)

	return kflx.RoxctlRelease{
		Kind:       "Release",
		APIVersion: "appstudio.redhat.com/v1alpha1",
		Metadata: kflx.Metadata{
			Name:      name,
			Namespace: config.KonfluxTenant,
			Labels: map[string]string{
				ACSReleaseVersionLabel: version.String(),
				ACSRoxctlArchLabel:     arch,
			},
			Annotations: map[string]string{
				AuthorAnnotation:             author,
				GeneratedTimestampAnnotation: time.Now().Format(time.RFC3339),
				VersionAnnotation:            versionPkg.GetVersion(),
			},
		},
		Spec: kflx.RoxctlReleaseSpec{
			ReleasePlan: releasePlan,
			Snapshot:    snapshot.Metadata.Name,
		},
	}, nil
}

func GetValidArchFromRoxctlStageRelease(stageRelease kflx.Release) (string, error) {
	arch, ok := stageRelease.Metadata.Labels[ACSRoxctlArchLabel]
	if !ok {
		return "", fmt.Errorf("%s not found in stage release labels", ACSRoxctlArchLabel)
	}
	return arch, nil
}

func PromoteRoxctlReleaseToProd(ctx context.Context, c *client.Client, stageRelease kflx.Release, applicationPrefix, org, repo, author string) (kflx.RoxctlRelease, error) {
	arch, err := GetValidArchFromRoxctlStageRelease(stageRelease)
	if err != nil {
		return kflx.RoxctlRelease{}, fmt.Errorf("failed to get arch from stage release: %w", err)
	}

	tag, ok := stageRelease.Metadata.Labels[ACSReleaseVersionLabel]
	if !ok {
		return kflx.RoxctlRelease{}, fmt.Errorf("%s not found in stage release labels", ACSReleaseVersionLabel)
	}

	snapshot, err := GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return kflx.RoxctlRelease{}, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	return RenderRoxctlRelease(snapshot, arch, "prod", 1, author)
}

func renderRoxctlReleasePlanName(arch, env string, version utils.Version) string {
	return fmt.Sprintf("acs-cdn-%s-%s-%d-%d", arch, env, version.Major, version.Minor)
}

func renderRoxctlReleaseName(application utils.Application, version utils.Version, arch, env string, attempt int) string {
	target := fmt.Sprintf("cdn-%s-%s", env, arch)
	return fmt.Sprintf("%s--%d-%d-%d--%s-%d", application.Name, version.Major, version.Minor, version.Patch, target, attempt)
}
