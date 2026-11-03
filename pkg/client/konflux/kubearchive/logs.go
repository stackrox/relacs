package kubearchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

func (c *Client) GetPodLogs(ctx context.Context, podName string, opts LogOptions) (string, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log", c.namespace, podName)

	params := url.Values{}
	if opts.Container != "" {
		params.Set("container", opts.Container)
	}
	if opts.TailLines > 0 {
		params.Set("tailLines", fmt.Sprintf("%d", opts.TailLines))
	}

	// Try live K8s cluster first, fall back to KubeArchive.
	if c.k8sAPIServerURL != "" {
		logs, err := c.fetchPodLogs(ctx, c.doRequestToK8sAPI, podName, path, params)
		if err == nil {
			return logs, nil
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) &&
			!errors.Is(err, ErrNotFound) &&
			!errors.Is(err, ErrUnauthorized) &&
			!errors.Is(err, ErrForbidden) {
			return "", err
		}
	}

	logs, err := c.fetchPodLogs(ctx, c.doRequest, podName, path, params)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("logs for pod %s: %w", podName, ErrNotFound)
		}
		return "", err
	}
	return logs, nil
}

func (c *Client) fetchPodLogs(ctx context.Context, doReq doRequestFunc, podName, path string, params url.Values) (string, error) {
	resp, err := doReq(ctx, "GET", path, params)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", podName, err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", ErrNotFound
	}
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
