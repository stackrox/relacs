package productlifecycle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/stackrox/relacs/pkg/config"
)

// FetchProductLifecycleDataForProduct retrieves product lifecycle information
// from Red Hat's Product Lifecycle API for the specified product name.
// It returns an error if the product is not found or if multiple products match.
func FetchProductLifecycleDataForProduct(productName string) (*ProductLifecycleData, error) {
	if productName == "" {
		return nil, fmt.Errorf("product name cannot be empty")
	}

	baseURL, err := url.Parse(config.ProductLifecycleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	query := baseURL.Query()
	query.Set("name", productName)
	baseURL.RawQuery = query.Encode()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	response, err := client.Get(baseURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get product lifecycle: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var productLifecycle ProductLifecycle
	err = json.Unmarshal(body, &productLifecycle)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product lifecycle: %w", err)
	}

	if len(productLifecycle.Data) != 1 {
		return nil, fmt.Errorf("expected exactly 1 product, got %d for product: %s", len(productLifecycle.Data), productName)
	}

	return &productLifecycle.Data[0], nil
}

// GetSupportedProductStreams filters product versions based on their support type
// and returns a list of version names that are eligible for merge window tracking.
// The returned versions are sorted in ascending order by semantic version.
func GetSupportedProductStreams(productLifecycleData *ProductLifecycleData) []string {
	supportedProductStreams := []string{}
	for _, productStream := range productLifecycleData.Versions {
		if slices.Contains(config.ProductLifecycleVersionTypesForMergeWindows, productStream.Type) {
			supportedProductStreams = append(supportedProductStreams, productStream.Name)
		}
	}

	// Sort by semantic version in ascending order
	sort.Slice(supportedProductStreams, func(i, j int) bool {
		vi, errI := semver.NewVersion(supportedProductStreams[i])
		vj, errJ := semver.NewVersion(supportedProductStreams[j])

		// If either version fails to parse, fall back to string comparison
		if errI != nil || errJ != nil {
			return supportedProductStreams[i] < supportedProductStreams[j]
		}

		return vi.LessThan(vj)
	})

	return supportedProductStreams
}
