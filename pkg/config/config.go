package config

import (
	"fmt"
	"sync"
)

// Config holds the application configuration
type Config struct {
	GitHubOrganization        string
	MainGitHubRepository      string
	KonfluxTenant             string
	KonfluxApplicationPrefix  string
	KonfluxHost               string
	ConformaDevScenarioPrefix string
	ImageTagSuffix            string
	IssueSource               string
	// JQL query template for fetching release data from Jira
	// Use with fmt.Sprintf, passing the fixVersion (tag) as the argument
	JiraReleaseDataQuery string
}

const (
	KonfluxTenant = "rh-acs-tenant"
	KonfluxHost   = "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com"
)

var (
	DefaultConfig = &Config{
		KonfluxTenant: KonfluxTenant,
		KonfluxHost:   KonfluxHost,
	}
	globalConfig     *Config
	globalConfigOnce sync.Once
	globalConfigErr  error
)

// Get returns the global Config instance, loading it on first call
func Get() (*Config, error) {
	globalConfigOnce.Do(func() {
		globalConfig, globalConfigErr = LoadConfig()
	})
	return globalConfig, globalConfigErr
}

// LoadConfig loads configuration from the Konflux ConfigMap
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig

	// Fetch ConfigMap from Konflux cluster
	configData, err := fetchConfigMap("relacs-config", cfg.KonfluxTenant)
	if err != nil {
		return nil, err
	}

	// Populate config from ConfigMap data
	cfg.GitHubOrganization = configData["github-organization"]
	cfg.MainGitHubRepository = configData["main-github-repository"]
	cfg.KonfluxApplicationPrefix = configData["konflux-application-prefix"]
	cfg.ConformaDevScenarioPrefix = configData["conforma-dev-scenario-prefix"]
	cfg.ImageTagSuffix = configData["image-tag-suffix"]
	cfg.IssueSource = configData["issue-source"]
	cfg.JiraReleaseDataQuery = configData["jira-release-data-query"]

	// Validate that all required fields are populated
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration fields are non-empty
func (c *Config) Validate() error {
	var missingFields []string

	if c.GitHubOrganization == "" {
		missingFields = append(missingFields, "github-organization")
	}
	if c.MainGitHubRepository == "" {
		missingFields = append(missingFields, "main-github-repository")
	}
	if c.KonfluxApplicationPrefix == "" {
		missingFields = append(missingFields, "konflux-application-prefix")
	}
	if c.ConformaDevScenarioPrefix == "" {
		missingFields = append(missingFields, "conforma-dev-scenario-prefix")
	}
	if c.ImageTagSuffix == "" {
		missingFields = append(missingFields, "image-tag-suffix")
	}
	if c.IssueSource == "" {
		missingFields = append(missingFields, "issue-source")
	}
	if c.JiraReleaseDataQuery == "" {
		missingFields = append(missingFields, "jira-release-data-query")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("ConfigMap is missing or has empty required fields: %v", missingFields)
	}

	return nil
}
