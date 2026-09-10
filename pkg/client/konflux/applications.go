package konflux

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/Masterminds/semver/v3"
)

// GetActiveProductStreamsFromApplications retrieves product streams by querying Konflux applications
// via kubectl. It filters applications matching the pattern ^acs-(\d+)-(\d+)$ and extracts
// the major and minor versions, returning them sorted in ascending order.
func (c *Client) GetActiveProductStreamsFromApplications() ([]ProductStream, error) {
	// Use direct kubectl command (not kubearchive) to get live applications
	cmd, err := c.directK8sCommandJSON("get", "application")
	if err != nil {
		return nil, fmt.Errorf("failed to create kubectl command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}

	var appList ApplicationList
	if err := json.Unmarshal(output, &appList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal applications: %w", err)
	}

	// Regular expression to match acs-<major>-<minor>
	re := regexp.MustCompile(`^acs-(\d+)-(\d+)$`)

	productStreams := []ProductStream{}
	for _, app := range appList.Items {
		matches := re.FindStringSubmatch(app.Metadata.Name)
		if matches == nil {
			continue
		}

		major, err := strconv.Atoi(matches[1])
		if err != nil {
			continue // Skip invalid major version
		}

		minor, err := strconv.Atoi(matches[2])
		if err != nil {
			continue // Skip invalid minor version
		}

		productStreams = append(productStreams, ProductStream{
			Major: major,
			Minor: minor,
		})
	}

	// Sort product streams by semantic version
	sort.Slice(productStreams, func(i, j int) bool {
		vi, errI := semver.NewVersion(productStreams[i].String())
		vj, errJ := semver.NewVersion(productStreams[j].String())

		// If either version fails to parse, fall back to numeric comparison
		if errI != nil || errJ != nil {
			if productStreams[i].Major != productStreams[j].Major {
				return productStreams[i].Major < productStreams[j].Major
			}
			return productStreams[i].Minor < productStreams[j].Minor
		}

		return vi.LessThan(vj)
	})

	return productStreams, nil
}
