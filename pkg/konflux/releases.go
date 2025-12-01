package konflux

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/stackrox/relacs/pkg/utils"
	"github.com/stackrox/relacs/pkg/vpn"
	"gopkg.in/yaml.v3"
)

func getRelease(ctx context.Context, name, namespace string) (string, error) {
	command := fmt.Sprintf("kubectl ka get release %s -n %s -o yaml", name, namespace)
	output, err := exec.CommandContext(ctx, "sh", "-c", command).Output()
	if err != nil {
		if len(output) > 0 {
			return "", fmt.Errorf("failed to get release: %s", string(output))
		}
		return "", nil
	}
	return string(output), nil
}

func GetOperatorBundleImageRefForRelease(ctx context.Context, name, namespace string) (string, error) {
	if err := vpn.CheckConnected(); err != nil {
		return "", fmt.Errorf("failed to check VPN connection: %w", err)
	}

	releaseYAML, err := getRelease(ctx, name, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get release: %w", err)
	}

	release, err := parseReleaseFromReleaseList(releaseYAML)
	if err != nil {
		return "", fmt.Errorf("failed to parse release: %w", err)
	}

	applicationName := getApplicationFromRelease(release)
	version, err := utils.GetVersionFromApplicationName(applicationName)
	if err != nil {
		return "", fmt.Errorf("failed to get version from application name: %w", err)
	}
	internalAdvisoryUrl := getInternalAdvisoryUrlFromRelease(release)

	advisory, err := downloadAdvisoryContent(internalAdvisoryUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download advisory content: %w", err)
	}

	operatorBundleImageRef, err := getOperatorBundleImageRefFromAdvisory(advisory, version)
	if err != nil {
		return "", fmt.Errorf("failed to get operator bundle image ref from advisory: %w", err)
	}
	return operatorBundleImageRef, nil
}

func parseReleaseFromReleaseList(releaseYAML string) (Release, error) {
	var releaseList ReleaseList
	err := yaml.Unmarshal([]byte(releaseYAML), &releaseList)
	if err != nil {
		return Release{}, fmt.Errorf("failed to unmarshal release list: %w", err)
	}
	if len(releaseList.Items) == 0 {
		return Release{}, fmt.Errorf("no releases found in the YAML output")
	}
	return releaseList.Items[0], nil
}

func getApplicationFromRelease(release Release) string {
	return release.Metadata.Labels["appstudio.openshift.io/application"]
}

func getInternalAdvisoryUrlFromRelease(release Release) string {
	return release.Status.Artifacts.Advisory.InternalURL
}

func downloadAdvisoryContent(url string) (Advisory, error) {
	response, err := http.Get(url)
	if err != nil {
		return Advisory{}, fmt.Errorf("failed to download advisory content: %w", err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Advisory{}, fmt.Errorf("failed to read advisory content: %w", err)
	}
	var advisory Advisory
	err = yaml.Unmarshal(responseBytes, &advisory)
	if err != nil {
		return Advisory{}, fmt.Errorf("failed to unmarshal advisory content: %w", err)
	}
	return advisory, nil
}

func getOperatorBundleImageRefFromAdvisory(advisory Advisory, version utils.Version) (string, error) {
	for _, image := range advisory.Spec.Content.Images {
		if image.Component == fmt.Sprintf("operator-bundle-%d-%d", version.Major, version.Minor) {
			return image.ContainerImage, nil
		}
	}
	return "", fmt.Errorf("no operator bundle image ref found for version %s in advisory %s", version, advisory.Metadata.Name)
}
