package kubearchive

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

func (c *Client) GetPodLogs(podName string, opts LogOptions) (string, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log", c.namespace, podName)

	params := url.Values{}
	if opts.Container != "" {
		params.Set("container", opts.Container)
	}
	if opts.TailLines > 0 {
		params.Set("tailLines", fmt.Sprintf("%d", opts.TailLines))
	}

	resp, err := c.doRequest("GET", path, params)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", podName, err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read logs for pod %s: %w", podName, err)
	}

	return strings.TrimSpace(string(body)), nil
}
