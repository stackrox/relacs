package konflux

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
	"github.com/stackrox/relacs/pkg/config"
)

type Client struct {
	kubearchive *kubearchive.Client
	once        sync.Once
	initErr     error
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) init() error {
	c.once.Do(func() {
		info, err := kubearchive.GetKubeconfigInfo()
		if err != nil {
			c.initErr = fmt.Errorf("failed to get kubeconfig info: %w", err)
			return
		}

		kaClient, err := kubearchive.NewClient(config.KubearchiveAPIURL, info.Token, config.KonfluxTenant)
		if err != nil {
			c.initErr = fmt.Errorf("failed to create KubeArchive client: %w", err)
			return
		}

		kaClient.SetK8sAPIServerURL(info.K8sAPIServerURL)
		c.kubearchive = kaClient
	})
	return c.initErr
}

func (c *Client) getKubearchive() (*kubearchive.Client, error) {
	if c.kubearchive != nil {
		return c.kubearchive, nil
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kubearchive, nil
}

func NewTestClient(httpClient *http.Client) *Client {
	return &Client{
		kubearchive: kubearchive.NewTestClient(httpClient, config.KubearchiveAPIURL, config.KonfluxTenant),
	}
}
