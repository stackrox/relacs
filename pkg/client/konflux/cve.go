package konflux

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/stackrox/relacs/pkg/utils"
)

type CVEInformation struct {
	Component string   `json:"component"`
	Key       string   `json:"key"`
	Severity  Severity `yaml:"-"`
	Year      string   `yaml:"-"`
}

func (c CVEInformation) GetSeverity() (Severity, error) {
	url := fmt.Sprintf("https://security.access.redhat.com/data/csaf/v2/vex/%s/%s.json", c.Year, strings.ToLower(c.Key))
	response, err := http.Get(url)
	if err != nil {
		return Severity_Low, fmt.Errorf("failed to download VEX data: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return Severity_Low, fmt.Errorf("failed to download VEX data: %s", response.Status)
	}

	//nolint:errcheck
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Severity_Low, fmt.Errorf("failed to read VEX data: %w", err)
	}
	vexData := VexData{}
	err = json.Unmarshal(responseBytes, &vexData)
	if err != nil {
		return Severity_Low, fmt.Errorf("failed to unmarshal VEX data: %w", err)
	}
	return vexData.GetSeverity()
}

func NewCVEInformation(application utils.Application, key string, component string) (CVEInformation, error) {
	formattedCVE, year, err := extractKeyAndYearFromCVE(key)
	if err != nil {
		return CVEInformation{}, fmt.Errorf("failed to format CVE key %s: %w", key, err)
	}

	componentWithVersion, err := application.GetResourceWithVersion(component)
	if err != nil {
		return CVEInformation{}, fmt.Errorf("failed to get versioned component name %s for CVE key %s: %w", component, key, err)
	}

	cve := CVEInformation{
		Component: componentWithVersion,
		Key:       formattedCVE,
		Year:      year,
	}

	cve.Severity, err = cve.GetSeverity()
	if err != nil {
		return CVEInformation{}, fmt.Errorf("failed to get severity for CVE %s: %w", cve.Key, err)
	}
	return cve, nil
}

func extractKeyAndYearFromCVE(cve string) (string, string, error) {
	cvePattern := regexp.MustCompile(`(cve|CVE)-(\d{4})-(\d+)`)
	matches := cvePattern.FindStringSubmatch(cve)
	if len(matches) != 4 {
		return "", "", fmt.Errorf("invalid CVE ID: %s", cve)
	}
	return strings.ToUpper(matches[0]), matches[2], nil
}
