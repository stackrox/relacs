package konflux

import (
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/utils"
)

type CVEInformation struct {
	Key       string   `json:"key"`
	Component string   `json:"component"`
	Severity  Severity `yaml:"-"`
	Year      string   `yaml:"-"`
}

// NewCVEInformation creates a CVEInformation from Jira CVE data
func NewCVEInformation(application utils.Application, cveKey string, cveYear string, component string, severityString string) (CVEInformation, error) {
	componentWithVersion, err := application.GetResourceWithVersion(component)
	if err != nil {
		return CVEInformation{}, fmt.Errorf("failed to get versioned component name %s for CVE key %s: %w", component, cveKey, err)
	}

	severity, err := parseSeverity(severityString)
	if err != nil {
		return CVEInformation{}, fmt.Errorf("failed to parse severity %q for CVE %s: %w", severityString, cveKey, err)
	}

	cve := CVEInformation{
		Key:       cveKey,
		Component: componentWithVersion,
		Severity:  severity,
		Year:      cveYear,
	}

	return cve, nil
}

// parseSeverity converts severity string to Severity enum
func parseSeverity(severityString string) (Severity, error) {
	severityMap := map[string]Severity{
		"Low":       Severity_Low,
		"Moderate":  Severity_Moderate,
		"Important": Severity_Important,
		"Critical":  Severity_Critical,
	}

	severity, ok := severityMap[severityString]
	if !ok {
		validSeverities := []string{"Low", "Moderate", "Important", "Critical"}
		return Severity_Low, fmt.Errorf("unexpected severity: %s (expected one of: %s)", severityString, strings.Join(validSeverities, ", "))
	}
	return severity, nil
}
