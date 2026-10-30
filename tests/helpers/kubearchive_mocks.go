package helpers

import (
	"fmt"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
)

const (
	TestK8sAPIServerURL = "https://api.test.example.com:6443"
)

// MockKubearchiveHealthCheckOK mocks a successful KubeArchive health check
func MockKubearchiveHealthCheckOK(baseURL string) {
	gock.New(baseURL).
		Get("/livez").
		Reply(200).
		BodyString("ok")
}

// MockKubearchiveHealthCheckFailure mocks a failed KubeArchive health check
func MockKubearchiveHealthCheckFailure(baseURL string) {
	gock.New(baseURL).
		Get("/livez").
		Reply(503).
		BodyString("service unavailable")
}

// MockQueryResources mocks a list query against a given base URL.
// Use with config.KubearchiveAPIURL for KubeArchive or TestK8sAPIServerURL for K8s API.
func MockQueryResources(baseURL, apiGroup, version, resource, namespace, labels string, response interface{}) {
	mock := gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, namespace, resource))

	if labels != "" {
		mock.MatchParam("labelSelector", labels)
	}

	mock.Reply(200).JSON(response)
}

// MockQueryResourcesEmpty mocks a list query returning no items.
func MockQueryResourcesEmpty(baseURL, apiGroup, version, resource, namespace string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, namespace, resource)).
		Reply(200).
		JSON(map[string]interface{}{
			"kind":     "List",
			"items":    []interface{}{},
			"metadata": map[string]interface{}{},
		})
}

// MockQueryResource mocks a single resource GET by name.
func MockQueryResource(baseURL, apiGroup, version, resource, namespace, name string, response interface{}) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s", apiGroup, version, namespace, resource, name)).
		Reply(200).
		JSON(response)
}

// MockQueryResourceNotFound mocks a 404 for a single resource GET.
func MockQueryResourceNotFound(baseURL, apiGroup, version, resource, namespace, name string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s", apiGroup, version, namespace, resource, name)).
		Reply(404).
		JSON(map[string]interface{}{"message": "resource not found"})
}

// MockQueryResourcesAuthFailure mocks a 401 Unauthorized response for a list query.
func MockQueryResourcesAuthFailure(baseURL, apiGroup, version, resource, namespace string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, namespace, resource)).
		Reply(401).
		JSON(map[string]interface{}{"message": "Unauthorized"})
}

// MockQueryResourcesForbidden mocks a 403 Forbidden response for a list query.
func MockQueryResourcesForbidden(baseURL, apiGroup, version, resource, namespace string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, namespace, resource)).
		Reply(403).
		JSON(map[string]interface{}{"message": "Forbidden"})
}

// MockGetPodLogs mocks a pod log retrieval. Uses the core API path (/api/v1/).
func MockGetPodLogs(baseURL, namespace, podName, container, response string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log", namespace, podName)).
		MatchParam("container", container).
		Reply(200).
		BodyString(response)
}

// MockGetPodLogsNotFound mocks a 404 for pod logs.
func MockGetPodLogsNotFound(baseURL, namespace, podName string) {
	gock.New(baseURL).
		Get(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log", namespace, podName)).
		Reply(404).
		JSON(map[string]interface{}{"message": "pod not found"})
}

// MockSelfSubjectReview mocks a successful SelfSubjectReview POST returning the given username.
func MockSelfSubjectReview(baseURL, username string) {
	gock.New(baseURL).
		Post("/apis/authentication.k8s.io/v1/selfsubjectreviews").
		Reply(201).
		JSON(map[string]interface{}{
			"apiVersion": "authentication.k8s.io/v1",
			"kind":       "SelfSubjectReview",
			"status": map[string]interface{}{
				"userInfo": map[string]interface{}{
					"username": username,
				},
			},
		})
}

// MockSelfSubjectReviewAuthFailure mocks a 401 on SelfSubjectReview.
func MockSelfSubjectReviewAuthFailure(baseURL string) {
	gock.New(baseURL).
		Post("/apis/authentication.k8s.io/v1/selfsubjectreviews").
		Reply(401).
		JSON(map[string]interface{}{"message": "Unauthorized"})
}

// MockQueryResourcesWithContinue mocks a paginated list query page.
// If continueToken is non-empty, the mock matches on the "continue" query param.
func MockQueryResourcesWithContinue(baseURL, apiGroup, version, resource, namespace, labels, continueToken string, response interface{}) {
	mock := gock.New(baseURL).
		Get(fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", apiGroup, version, namespace, resource))

	if labels != "" {
		mock.MatchParam("labelSelector", labels)
	}
	if continueToken != "" {
		mock.MatchParam("continue", continueToken)
	}

	mock.Reply(200).JSON(response)
}

// Convenience helpers using default test URLs

// MockKubearchiveReleases mocks a release list query against KubeArchive.
func MockKubearchiveReleases(labels string, response interface{}) {
	MockQueryResources(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, labels, response)
}

// MockK8sReleases mocks a release list query against the K8s API server.
func MockK8sReleases(labels string, response interface{}) {
	MockQueryResources(TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, labels, response)
}

// MockKubearchiveSnapshots mocks a snapshot list query against KubeArchive.
func MockKubearchiveSnapshots(labels string, response interface{}) {
	MockQueryResources(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant, labels, response)
}

// MockK8sSnapshots mocks a snapshot list query against the K8s API server.
func MockK8sSnapshots(labels string, response interface{}) {
	MockQueryResources(TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant, labels, response)
}

// MockKubearchivePipelineRuns mocks a pipelinerun list query against KubeArchive.
func MockKubearchivePipelineRuns(labels string, response interface{}) {
	MockQueryResources(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant, labels, response)
}

// MockK8sPipelineRuns mocks a pipelinerun list query against the K8s API server.
func MockK8sPipelineRuns(labels string, response interface{}) {
	MockQueryResources(TestK8sAPIServerURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant, labels, response)
}
