package konflux

import (
	"fmt"
	"io"
	"net/http"

	"github.com/stackrox/relacs/pkg/utils"
	"gopkg.in/yaml.v3"
)

func GetAdvisoryFromRelease(release Release) (Advisory, error) {
	advisory := Advisory{}
	url := release.GetAdvisoryInternalURL()
	response, err := http.Get(url)
	if err != nil {
		return advisory, fmt.Errorf("failed to download advisory content: %w", err)
	}
	//nolint:errcheck
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return advisory, fmt.Errorf("failed to read advisory content: %w", err)
	}
	err = yaml.Unmarshal(responseBytes, &advisory)
	if err != nil {
		return advisory, fmt.Errorf("failed to unmarshal advisory content: %w", err)
	}
	return advisory, nil
}

func (a Advisory) GetOperatorBundleImage(application string) (string, error) {
	operatorBundleComponentName, err := utils.GetResourceWithVersion("operator-bundle", application)
	if err != nil {
		return "", fmt.Errorf("failed to get operator bundle component name: %w", err)
	}
	for _, image := range a.Spec.Content.Images {
		if image.Component == operatorBundleComponentName {
			return image.ContainerImage, nil
		}
	}
	return "", fmt.Errorf("operator bundle image not found in advisory")
}
