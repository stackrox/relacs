package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type RPASConfig struct {
	CDN CDNConfig `yaml:"cdn"`
}

type CDNConfig struct {
	Architectures  []ArchitectureConfig `yaml:"architectures"`
	ActiveVersions []VersionConfig      `yaml:"active_versions"`
	Targets        []TargetConfig       `yaml:"targets"`
	Timeout        string               `yaml:"timeout"`
}

type VersionConfig struct {
	Version    string `yaml:"version"`
	RHELTarget string `yaml:"rhel_target"`
}

type ArchitectureConfig struct {
	Name              string     `yaml:"name"`
	ProductID         int        `yaml:"product_id"`
	ProductNameSuffix string     `yaml:"product_name_suffix"`
	SourceArch        string     `yaml:"source_arch"`
	OS                []OSConfig `yaml:"os"`
}

type OSConfig struct {
	Name             string `yaml:"name"`
	ArchiveExtension string `yaml:"archive_extension"`
	SourceOS         string `yaml:"source_os"`
	PrettyName       string `yaml:"pretty_name"`
}

type TargetConfig struct {
	Name               string `yaml:"name"`
	CDNEnv             string `yaml:"cdn_env"`
	Intention          string `yaml:"intention"`
	ServiceAccountName string `yaml:"serviceAccountName"`
	GitBranch          string `yaml:"gitBranch"`
}

func main() {
	rpasConfig, err := parseRPASConfigYAML("rpas.config.yaml")
	if err != nil {
		log.Fatalf("Failed to parse rpas.config.yaml: %v", err)
	}

	templateFile, err := readTemplateFile("template.cdn-rpa.yaml")
	if err != nil {
		log.Fatalf("Failed to read template.cdn-rpa.yaml: %v", err)
	}

	// for each active version x architecture x target, generate an RPA file
	for _, architecture := range rpasConfig.CDN.Architectures {
		for _, target := range rpasConfig.CDN.Targets {
			for _, version := range rpasConfig.CDN.ActiveVersions {
				generateRPAFile(templateFile, architecture, target, version, rpasConfig.CDN.Timeout)
			}
		}
	}
}

// Parse rpas.config.yaml file
func parseRPASConfigYAML(filename string) (RPASConfig, error) {
	var rpasConfig RPASConfig
	yamlData, err := os.ReadFile(filename)
	if err != nil {
		return RPASConfig{}, err
	}
	err = yaml.Unmarshal(yamlData, &rpasConfig)
	if err != nil {
		return RPASConfig{}, err
	}
	return rpasConfig, nil
}

func generateRPAFile(templateFile string, architecture ArchitectureConfig, target TargetConfig, version VersionConfig, timeout string) {
	parts := strings.Split(version.Version, ".")
	renderedTemplate, err := renderTemplate(templateFile, map[string]interface{}{
		"architecture":               architecture,
		"target":                     target,
		"version":                    version,
		"timeout":                    timeout,
		"major_version":              parts[0],
		"major_minor_version":        parts[0] + "." + parts[1],
		"dashed_major_minor_version": parts[0] + "-" + parts[1],
	})
	if err != nil {
		log.Fatalf("failed to render template: %v", err)
	}

	filename := fmt.Sprintf("acs-cdn-%s-%s-%s-%s.yaml",
		architecture.Name,
		target.Name,
		parts[0],
		parts[1],
	)
	filepath := "../../../gitlab.cee.redhat.com/releng/konflux-release-data/config/stone-prd-rh01.pg1f.p1/product/ReleasePlanAdmission/rh-acs"
	writeRPAFile(filepath, filename, renderedTemplate)
}

func readTemplateFile(filename string) (string, error) {
	yamlData, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

func renderTemplate(templateFile string, data any) (string, error) {
	tmpl, err := template.New("template").Funcs(template.FuncMap{
		"capitalize": strings.Title,
		"dashToUnderscore": func(s string) string {
			return strings.ReplaceAll(s, "-", "_")
		},
	}).Parse(templateFile)
	if err != nil {
		return "", err
	}
	var renderedTemplate bytes.Buffer
	err = tmpl.Execute(&renderedTemplate, data)
	if err != nil {
		return "", err
	}
	return renderedTemplate.String(), nil
}

func writeRPAFile(filepath, filename, content string) error {
	return os.WriteFile(filepath+"/"+filename, []byte(content), 0644)
}
