package kubearchive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type selfSubjectReview struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type selfSubjectReviewResponse struct {
	Status struct {
		UserInfo struct {
			Username string `json:"username"`
		} `json:"userInfo"`
	} `json:"status"`
}

func (c *Client) GetCurrentUsername() (string, error) {
	if c.k8sAPIServerURL == "" {
		return "", fmt.Errorf("kubernetes API server URL not configured — ensure kubeconfig is loaded")
	}

	body, err := json.Marshal(selfSubjectReview{
		APIVersion: "authentication.k8s.io/v1",
		Kind:       "SelfSubjectReview",
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal SelfSubjectReview: %w", err)
	}

	u := c.k8sAPIServerURL + "/apis/authentication.k8s.io/v1/selfsubjectreviews"
	req, err := http.NewRequest("POST", u, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query Kubernetes API: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("kubernetes API returned 401 Unauthorized — token may be expired, run 'oc login'")
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("kubernetes API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var review selfSubjectReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&review); err != nil {
		return "", fmt.Errorf("failed to decode SelfSubjectReview response: %w", err)
	}

	username := review.Status.UserInfo.Username
	if username == "" {
		return "", fmt.Errorf("empty username returned from Kubernetes API")
	}

	return username, nil
}
