package konflux

import (
	"fmt"
	"net/http"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
	"github.com/stackrox/relacs/pkg/config"
)

type Client struct {
	kubearchive *kubearchive.Client
}

func NewClient() (*Client, error) {
	info, err := kubearchive.GetKubeconfigInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig info: %w", err)
	}

	kaClient, err := kubearchive.NewClient(config.KubearchiveAPIURL, info.Token, config.KonfluxTenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create KubeArchive client: %w", err)
	}

	kaClient.SetK8sAPIServerURL(info.K8sAPIServerURL)

	return &Client{kubearchive: kaClient}, nil
}

func NewTestClient(httpClient *http.Client) *Client {
	return &Client{
		kubearchive: kubearchive.NewTestClient(httpClient, config.KubearchiveAPIURL, config.KonfluxTenant),
	}
}
