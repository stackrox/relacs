package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFetchConfigMap_ParsesValidJSON(t *testing.T) {
	// This test would require mocking exec.Command which is complex
	// For now, we'll test the JSON unmarshaling logic separately
	validConfigMapJSON := `{
		"data": {
			"github-organization": "stackrox",
			"main-github-repository": "stackrox",
			"konflux-application-prefix": "acs",
			"conforma-dev-scenario-prefix": "acs-conforma-dev",
			"image-tag-suffix": "fast",
			"issue-source": "issues.redhat.com",
			"jira-release-data-query": "project = 'Test' AND fixVersion = %s"
		}
	}`

	var configMap struct {
		Data map[string]string `json:"data"`
	}

	err := json.Unmarshal([]byte(validConfigMapJSON), &configMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal valid ConfigMap JSON: %v", err)
	}

	if configMap.Data == nil {
		t.Fatal("ConfigMap data field is nil")
	}

	expectedValues := map[string]string{
		"github-organization":          "stackrox",
		"main-github-repository":       "stackrox",
		"konflux-application-prefix":   "acs",
		"conforma-dev-scenario-prefix": "acs-conforma-dev",
		"image-tag-suffix":             "fast",
		"issue-source":                 "issues.redhat.com",
		"jira-release-data-query":      "project = 'Test' AND fixVersion = %s",
	}

	for key, expectedValue := range expectedValues {
		if value, ok := configMap.Data[key]; !ok {
			t.Errorf("Missing key %s in ConfigMap data", key)
		} else if value != expectedValue {
			t.Errorf("For key %s: expected %s, got %s", key, expectedValue, value)
		}
	}
}

func TestFetchConfigMap_IncompleteConfigMapDetected(t *testing.T) {
	// Test that parsing a ConfigMap with missing fields is detected
	incompleteConfigMapJSON := `{
		"data": {
			"github-organization": "stackrox",
			"main-github-repository": "stackrox"
		}
	}`

	var configMap struct {
		Data map[string]string `json:"data"`
	}

	err := json.Unmarshal([]byte(incompleteConfigMapJSON), &configMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal ConfigMap JSON: %v", err)
	}

	// Simulate what LoadConfig does - populate a Config struct
	cfg := &Config{
		KonfluxTenant: "rh-acs-tenant",
		KonfluxHost:   "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com",
	}
	cfg.GitHubOrganization = configMap.Data["github-organization"]
	cfg.MainGitHubRepository = configMap.Data["main-github-repository"]
	cfg.KonfluxApplicationPrefix = configMap.Data["konflux-application-prefix"]
	cfg.ConformaDevScenarioPrefix = configMap.Data["conforma-dev-scenario-prefix"]
	cfg.ImageTagSuffix = configMap.Data["image-tag-suffix"]
	cfg.IssueSource = configMap.Data["issue-source"]
	cfg.JiraReleaseDataQuery = configMap.Data["jira-release-data-query"]

	// Validation should fail
	err = cfg.Validate()
	if err == nil {
		t.Error("Validate() should return error for incomplete ConfigMap")
	}

	// Should mention the missing fields
	errorMsg := err.Error()
	missingFields := []string{
		"konflux-application-prefix",
		"conforma-dev-scenario-prefix",
		"image-tag-suffix",
		"issue-source",
		"jira-release-data-query",
	}

	for _, field := range missingFields {
		if !strings.Contains(errorMsg, field) {
			t.Errorf("Error message should mention missing field '%s', got: %v", field, errorMsg)
		}
	}
}

func TestValidate_AllFieldsPresent(t *testing.T) {
	cfg := &Config{
		GitHubOrganization:        "stackrox",
		MainGitHubRepository:      "stackrox",
		KonfluxTenant:             "rh-acs-tenant",
		KonfluxApplicationPrefix:  "acs",
		KonfluxHost:               "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com",
		ConformaDevScenarioPrefix: "acs-conforma-dev",
		ImageTagSuffix:            "fast",
		IssueSource:               "issues.redhat.com",
		JiraReleaseDataQuery:      "project = 'Test' AND fixVersion = %s",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should not return error for valid config, got: %v", err)
	}
}

func TestValidate_MissingFields(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *Config
		expectedError string
	}{
		{
			name: "missing github-organization",
			cfg: &Config{
				MainGitHubRepository:      "stackrox",
				KonfluxApplicationPrefix:  "acs",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				ImageTagSuffix:            "fast",
				IssueSource:               "issues.redhat.com",
				JiraReleaseDataQuery:      "test",
			},
			expectedError: "github-organization",
		},
		{
			name: "missing main-github-repository",
			cfg: &Config{
				GitHubOrganization:        "stackrox",
				KonfluxApplicationPrefix:  "acs",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				ImageTagSuffix:            "fast",
				IssueSource:               "issues.redhat.com",
				JiraReleaseDataQuery:      "test",
			},
			expectedError: "main-github-repository",
		},
		{
			name: "missing konflux-application-prefix",
			cfg: &Config{
				GitHubOrganization:        "stackrox",
				MainGitHubRepository:      "stackrox",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				ImageTagSuffix:            "fast",
				IssueSource:               "issues.redhat.com",
				JiraReleaseDataQuery:      "test",
			},
			expectedError: "konflux-application-prefix",
		},
		{
			name: "missing conforma-dev-scenario-prefix",
			cfg: &Config{
				GitHubOrganization:       "stackrox",
				MainGitHubRepository:     "stackrox",
				KonfluxApplicationPrefix: "acs",
				ImageTagSuffix:           "fast",
				IssueSource:              "issues.redhat.com",
				JiraReleaseDataQuery:     "test",
			},
			expectedError: "conforma-dev-scenario-prefix",
		},
		{
			name: "missing image-tag-suffix",
			cfg: &Config{
				GitHubOrganization:        "stackrox",
				MainGitHubRepository:      "stackrox",
				KonfluxApplicationPrefix:  "acs",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				IssueSource:               "issues.redhat.com",
				JiraReleaseDataQuery:      "test",
			},
			expectedError: "image-tag-suffix",
		},
		{
			name: "missing issue-source",
			cfg: &Config{
				GitHubOrganization:        "stackrox",
				MainGitHubRepository:      "stackrox",
				KonfluxApplicationPrefix:  "acs",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				ImageTagSuffix:            "fast",
				JiraReleaseDataQuery:      "test",
			},
			expectedError: "issue-source",
		},
		{
			name: "missing jira-release-data-query",
			cfg: &Config{
				GitHubOrganization:        "stackrox",
				MainGitHubRepository:      "stackrox",
				KonfluxApplicationPrefix:  "acs",
				ConformaDevScenarioPrefix: "acs-conforma-dev",
				ImageTagSuffix:            "fast",
				IssueSource:               "issues.redhat.com",
			},
			expectedError: "jira-release-data-query",
		},
		{
			name: "multiple missing fields",
			cfg: &Config{
				GitHubOrganization: "stackrox",
				ImageTagSuffix:     "fast",
			},
			expectedError: "main-github-repository",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.cfg.Validate()
			if err == nil {
				t.Error("Validate() should return error for missing fields")
			}
			if err != nil && !strings.Contains(err.Error(), test.expectedError) {
				t.Errorf("Validate() error should mention '%s', got: %v", test.expectedError, err)
			}
		})
	}
}

func TestValidate_EmptyStrings(t *testing.T) {
	// Empty strings should be treated the same as missing fields
	cfg := &Config{
		GitHubOrganization:        "",
		MainGitHubRepository:      "",
		KonfluxApplicationPrefix:  "",
		ConformaDevScenarioPrefix: "",
		ImageTagSuffix:            "",
		IssueSource:               "",
		JiraReleaseDataQuery:      "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() should return error for all empty fields")
	}

	// Should list all missing fields
	errorMsg := err.Error()
	requiredFields := []string{
		"github-organization",
		"main-github-repository",
		"konflux-application-prefix",
		"conforma-dev-scenario-prefix",
		"image-tag-suffix",
		"issue-source",
		"jira-release-data-query",
	}

	for _, field := range requiredFields {
		if !strings.Contains(errorMsg, field) {
			t.Errorf("Error message should mention missing field '%s', got: %v", field, errorMsg)
		}
	}
}
