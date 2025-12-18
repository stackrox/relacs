package konflux

import (
	"context"
	"fmt"

	"github.com/stackrox/relacs/pkg/client"

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
