package kubearchive

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	httpClient      *http.Client
	baseURL         string
	token           string
	namespace       string
	k8sAPIServerURL string
}

func NewClient(baseURL, token, namespace string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}

	c := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
		namespace:  namespace,
	}

	if err := c.healthCheck(); err != nil {
		return nil, fmt.Errorf("KubeArchive API unreachable at %s: %w", baseURL, err)
	}

	return c, nil
}

func NewTestClient(httpClient *http.Client, baseURL, namespace string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		httpClient:      httpClient,
		baseURL:         baseURL,
		token:           "test-token",
		namespace:       namespace,
		k8sAPIServerURL: "https://api.test.example.com:6443",
	}
}

func (c *Client) healthCheck() error {
	resp, err := c.doRequest("GET", "/livez", nil)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) doRequest(method, path string, params url.Values) (*http.Response, error) {
	return c.doRequestTo(c.baseURL, method, path, params)
}

func (c *Client) doRequestToK8sAPI(method, path string, params url.Values) (*http.Response, error) {
	return c.doRequestTo(c.k8sAPIServerURL, method, path, params)
}

func (c *Client) doRequestTo(baseURL, method, path string, params url.Values) (*http.Response, error) {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Unauthorized — token may be expired, run 'oc login'",
		}
	}
	if resp.StatusCode == http.StatusForbidden {
		_ = resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Forbidden — insufficient permissions, check RBAC for the KubeArchive namespace",
		}
	}

	return resp, nil
}

type kubeconfigInfo struct {
	Token           string
	K8sAPIServerURL string
}

func GetKubeconfigInfo() (*kubeconfigInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if restConfig.BearerToken == "" {
		return nil, fmt.Errorf("no bearer token found in kubeconfig — run 'oc login' to authenticate")
	}

	return &kubeconfigInfo{
		Token:           restConfig.BearerToken,
		K8sAPIServerURL: restConfig.Host,
	}, nil
}

func (c *Client) SetK8sAPIServerURL(url string) {
	c.k8sAPIServerURL = url
}
