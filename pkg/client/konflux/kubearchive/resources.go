package kubearchive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const defaultLimit = 1000

func queryResourceList[T any](resp *http.Response, resource string) ([]T, error) {
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	var list KubeList[T]
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode %s list: %w", resource, err)
	}

	return list.Items, nil
}

// QueryResources queries both the live K8s cluster and KubeArchive, merging
// results with deduplication by UID (live cluster wins on conflict).
func QueryResources[T HasUID](c *Client, apiGroup, version, resource string, opts QueryOptions) ([]T, error) {
	path := fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, c.namespace, resource)

	params := url.Values{}
	if opts.LabelSelector != "" {
		params.Set("labelSelector", opts.LabelSelector)
	}
	limit := defaultLimit
	if opts.Limit > 0 {
		limit = opts.Limit
	}
	params.Set("limit", strconv.Itoa(limit))

	// Query live K8s cluster first (authoritative for current resources).
	var clusterItems []T
	if c.k8sAPIServerURL != "" {
		resp, err := c.doRequestToK8sAPI("GET", path, params)
		if err == nil {
			items, err := queryResourceList[T](resp, resource)
			if err == nil {
				clusterItems = items
			}
		}
	}

	// Query KubeArchive for archived resources.
	var archiveItems []T
	resp, err := c.doRequest("GET", path, params)
	if err == nil {
		items, err := queryResourceList[T](resp, resource)
		if err == nil {
			archiveItems = items
		}
	}

	if len(clusterItems) == 0 && len(archiveItems) == 0 && err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", resource, err)
	}

	return mergeByUID(clusterItems, archiveItems), nil
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
func QueryResource[T any](c *Client, apiGroup, version, resource, name string) (T, error) {
	var zero T
	path := fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s", apiGroup, version, c.namespace, resource, name)

	// Try live K8s cluster first.
	if c.k8sAPIServerURL != "" {
		resp, err := c.doRequestToK8sAPI("GET", path, nil)
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
	resp, err := c.doRequest("GET", path, nil)
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
