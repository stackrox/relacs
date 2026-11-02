package kubearchive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strconv"

	"github.com/stackrox/relacs/pkg/config"
)

func queryResourcePage[T any](resp *http.Response, resource string) ([]T, string, error) {
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	var list KubeList[T]
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, "", fmt.Errorf("failed to decode %s list: %w", resource, err)
	}

	return list.Items, list.Metadata.Continue, nil
}

type doRequestFunc func(ctx context.Context, method, path string, params url.Values) (*http.Response, error)

func queryAllPages[T any](ctx context.Context, doReq doRequestFunc, path, resource string, params url.Values) ([]T, error) {
	var allItems []T
	p := cloneParams(params)
	p.Set("limit", strconv.Itoa(config.KubearchivePageSize))

	for {
		resp, err := doReq(ctx, "GET", path, p)
		if err != nil {
			return allItems, err
		}

		items, continueToken, err := queryResourcePage[T](resp, resource)
		if err != nil {
			return allItems, err
		}
		allItems = append(allItems, items...)

		if continueToken == "" {
			break
		}
		p.Set("continue", continueToken)
	}

	return allItems, nil
}

func cloneParams(p url.Values) url.Values {
	c := url.Values{}
	maps.Copy(c, p)
	return c
}

// QueryResources queries both the live K8s cluster and KubeArchive, merging
// results with deduplication by UID (live cluster wins on conflict).
func QueryResources[T HasUID](ctx context.Context, c *Client, apiGroup, version, resource string, opts QueryOptions) ([]T, error) {
	path := fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, c.namespace, resource)

	params := url.Values{}
	if opts.LabelSelector != "" {
		params.Set("labelSelector", opts.LabelSelector)
	}

	// Query live K8s cluster first (authoritative for current resources).
	var clusterItems []T
	if c.k8sAPIServerURL != "" {
		items, err := queryAllPages[T](ctx, c.doRequestToK8sAPI, path, resource, params)
		if err == nil {
			clusterItems = items
		} else if !isSkippableLiveError(err) {
			return nil, fmt.Errorf("failed to query live %s: %w", resource, err)
		}
	}

	// Query KubeArchive for archived resources.
	var archiveItems []T
	var archiveErr error
	items, err := queryAllPages[T](ctx, c.doRequest, path, resource, params)
	if err == nil {
		archiveItems = items
	} else {
		archiveErr = err
	}

	if len(clusterItems) == 0 && len(archiveItems) == 0 && archiveErr != nil {
		return nil, fmt.Errorf("failed to query %s: %w", resource, archiveErr)
	}

	return mergeByUID(clusterItems, archiveItems), nil
}

func isSkippableLiveError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return true
	}
	switch apiErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return true
	}
	return false
}

func querySingleResource[T any](resp *http.Response, resource, name string) (T, bool, error) {
	var zero T
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return zero, false, nil
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return zero, false, &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, false, fmt.Errorf("failed to decode %s/%s: %w", resource, name, err)
	}

	return result, true, nil
}

// QueryResource gets a single resource by name, trying the live cluster first,
// then falling back to KubeArchive.
func QueryResource[T any](ctx context.Context, c *Client, apiGroup, version, resource, name string) (T, error) {
	var zero T
	path := fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s", apiGroup, version, c.namespace, resource, name)

	// Try live K8s cluster first.
	if c.k8sAPIServerURL != "" {
		resp, err := c.doRequestToK8sAPI(ctx, "GET", path, nil)
		if err == nil {
			result, found, err := querySingleResource[T](resp, resource, name)
			if err != nil {
				return zero, err
			}
			if found {
				return result, nil
			}
		}
	}

	// Fall back to KubeArchive.
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return zero, fmt.Errorf("failed to query %s/%s: %w", resource, name, err)
	}

	result, found, err := querySingleResource[T](resp, resource, name)
	if err != nil {
		return zero, err
	}
	if !found {
		return zero, fmt.Errorf("%s %s not found", resource, name)
	}

	return result, nil
}

// mergeByUID combines items from both sources, deduplicating by UID.
// Items from the first slice (cluster) take precedence.
func mergeByUID[T HasUID](cluster, archive []T) []T {
	seen := make(map[string]bool, len(cluster))
	result := make([]T, 0, len(cluster)+len(archive))

	for _, item := range cluster {
		uid := item.GetUID()
		if uid != "" {
			seen[uid] = true
		}
		result = append(result, item)
	}

	for _, item := range archive {
		uid := item.GetUID()
		if uid != "" && seen[uid] {
			continue
		}
		result = append(result, item)
	}

	return result
}
