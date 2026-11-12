# Migration Plan: KubeArchive CLI to Tekton Results API

## Context

Currently, `relacs` uses the KubeArchive CLI (`kubectl ka`) to query historical Kubernetes resources (Releases, Snapshots, PipelineRuns) from the Konflux cluster. This requires:
- The `KUBECTL_PLUGIN_KA_HOST` environment variable to be set
- The KubeArchive kubectl plugin to be installed
- Subprocess execution for every query (performance overhead)
- Hard-coded pagination limit of 1000 items

The goal is to replace this with direct REST API calls to the Tekton Results API service running in the cluster. This migration will:
- Eliminate the KubeArchive CLI dependency
- Use KUBECONFIG for authentication (consistent with current kubectl approach)
- Improve performance (direct HTTP vs subprocess)
- Provide better pagination control
- Simplify the client architecture

This plan focuses on migrating **GetRelease** and **GetReleases** methods first, as a foundation for migrating other methods later.

---

## Architecture Design

### 1. Configuration Changes

**File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/config/config.go`

Add new configuration constant for Tekton Results API endpoint:

```go
// Tekton Results API endpoint (configurable via env var or use the known endpoint)
const TektonResultsAPIEndpoint = "TEKTON_RESULTS_API_ENDPOINT"
const DefaultTektonResultsEndpoint = "https://tekton-results-tekton-results.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com"
```

**Environment Variables:**
- `TEKTON_RESULTS_API_ENDPOINT` (optional) - Custom endpoint URL
- Authentication uses the current kubectl/oc context from KUBECONFIG

### 2. New Tekton Results HTTP Client

**New File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/client.go`

Create a new HTTP-based client following the pattern from `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/jira/client.go`:

**Client Structure:**
```go
type Client struct {
    httpClient *http.Client
    baseURL    string
    token      string
    namespace  string
}
```

**Authentication Strategy:**

Use KUBECONFIG authentication by extracting the bearer token from the current kubectl/oc context. This approach:
- Leverages existing kubectl/oc authentication setup
- Works with any auth method configured in kubeconfig (tokens, exec plugins, etc.)
- Consistent with the current KubeArchive approach

**Token Extraction Options:**
1. Execute `kubectl config view --minify -o jsonpath='{.users[0].user.token}'` to get token directly
2. If that fails, execute `kubectl config view --raw` and parse the kubeconfig YAML to extract the token
3. Support for exec-based auth plugins by running the exec command if configured

**Initialization Logic:**
```go
func NewClient() (*Client, error) {
    // 1. Determine endpoint (env var or default)
    endpoint := os.Getenv(config.TektonResultsAPIEndpoint)
    if endpoint == "" {
        endpoint = config.DefaultTektonResultsEndpoint
    }

    // 2. Extract token from kubeconfig
    token, err := getKubeconfigToken()
    if err != nil {
        return nil, fmt.Errorf("failed to get kubeconfig token: %w", err)
    }

    return &Client{
        httpClient: &http.Client{Timeout: 180 * time.Second}, // API is very slow, needs generous timeout
        baseURL:    endpoint,
        token:      token,
        namespace:  config.KonfluxTenant,
    }, nil
}

func (c *Client) addAuthHeader(req *http.Request) {
    req.Header.Set("Authorization", "Bearer " + c.token)
    req.Header.Set("Content-Type", "application/json")
}
```

### 3. Tekton Results API Integration

**Understanding the Tekton Results API:**

The Tekton Results API stores Tekton resources (including custom resources like Release) with the following structure:
- **Endpoint:** `GET /apis/results.tekton.dev/v1alpha2/parents/{parent}/results/-/records`
  - Note: Use `/records` endpoint directly (not `/results` then fetch records)
- **Parent:** Namespace-scoped parent ID (e.g., `rh-acs-tenant/results/-`)
- **Filtering:** CEL expressions via query parameter `filter`
  - Example: `filter=data.metadata.labels['appstudio.openshift.io/application']=='acs'`
  - Example: `filter=data.metadata.name=='rhacs--4-10-1--stage-1'`
- **Pagination:** `page_token` and `page_size` parameters (min=6, max=10000)

**API Response Format:**
```json
{
  "records": [
    {
      "name": "rh-acs-tenant/records/release-xyz",
      "id": "release-xyz",
      "data": {
        "@type": "type.googleapis.com/tekton.dev/Resource",
        "kind": "Release",
        "apiVersion": "appstudio.redhat.com/v1alpha1",
        "metadata": {...},
        "spec": {...},
        "status": {...}
      }
    }
  ],
  "nextPageToken": "..."
}
```

**Important API Constraints:**
- Minimum `page_size` is 6 (API returns error if < 6)
- Maximum `page_size` is 10000
- **API is very slow** - queries can take 60+ seconds to respond
- HTTP client timeout must be generous (180s recommended)
- Empty replies from server indicate query timeout on API side

Each record in the response contains the full Kubernetes resource in the `data` field.

### 4. Implementation of GetRelease and GetReleases

**File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/releases.go`

**GetReleases Implementation:**
```go
func (c *Client) GetReleases(labels []string) ([]konflux.Release, error) {
    // Build CEL filter expression
    filter := buildCELFilter(labels)

    // Query Tekton Results API records endpoint directly with pagination
    parent := fmt.Sprintf("%s/results/-", c.namespace)
    baseURL := fmt.Sprintf("%s/apis/results.tekton.dev/v1alpha2/parents/%s/records",
        c.baseURL, parent)

    var allReleases []konflux.Release
    pageToken := ""

    for {
        // Build paginated request URL
        params := url.Values{}
        params.Add("filter", filter)
        params.Add("page_size", "100") // Min=6, Max=10000, use 100 for reasonable batches
        if pageToken != "" {
            params.Add("page_token", pageToken)
        }
        
        requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

        // Execute HTTP request with auth
        req, err := http.NewRequest("GET", requestURL, nil)
        if err != nil {
            return nil, fmt.Errorf("failed to create request: %w", err)
        }
        c.addAuthHeader(req)

        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, fmt.Errorf("failed to execute request: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
        }

        // Parse response
        var recordsResp RecordsResponse
        if err := json.NewDecoder(resp.Body).Decode(&recordsResp); err != nil {
            return nil, fmt.Errorf("failed to decode response: %w", err)
        }

        // Extract Release objects from records
        for _, record := range recordsResp.Records {
            if record.Data.Kind != "Release" {
                continue
            }
            release := konflux.Release{
                Kind:       record.Data.Kind,
                APIVersion: record.Data.APIVersion,
                Metadata:   record.Data.Metadata,
                Spec:       record.Data.Spec,
                Status:     record.Data.Status,
            }
            allReleases = append(allReleases, release)
        }

        // Check for more pages
        if recordsResp.NextPageToken == "" {
            break
        }
        pageToken = recordsResp.NextPageToken
    }

    return allReleases, nil
}
```

**GetRelease Implementation:**
```go
func (c *Client) GetRelease(name string) (konflux.Release, error) {
    // Query by name using CEL filter
    filter := fmt.Sprintf("data.metadata.name=='%s'", name)
    
    // Query directly with filter
    parent := fmt.Sprintf("%s/results/-", c.namespace)
    params := url.Values{}
    params.Add("filter", filter)
    params.Add("page_size", "10") // Expect only 1 result
    
    requestURL := fmt.Sprintf("%s/apis/results.tekton.dev/v1alpha2/parents/%s/records?%s",
        c.baseURL, parent, params.Encode())

    req, err := http.NewRequest("GET", requestURL, nil)
    if err != nil {
        return konflux.Release{}, fmt.Errorf("failed to create request: %w", err)
    }
    c.addAuthHeader(req)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return konflux.Release{}, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return konflux.Release{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
    }

    var recordsResp RecordsResponse
    if err := json.NewDecoder(resp.Body).Decode(&recordsResp); err != nil {
        return konflux.Release{}, fmt.Errorf("failed to decode response: %w", err)
    }

    if len(recordsResp.Records) == 0 {
        return konflux.Release{}, fmt.Errorf("release %s not found", name)
    }

    record := recordsResp.Records[0]
    return konflux.Release{
        Kind:       record.Data.Kind,
        APIVersion: record.Data.APIVersion,
        Metadata:   record.Data.Metadata,
        Spec:       record.Data.Spec,
        Status:     record.Data.Status,
    }, nil
}
```

**CEL Filter Builder:**
```go
func buildCELFilter(labels []string) string {
    // Convert kubectl-style labels to Tekton Results CEL filter format
    // Input: ["pac.test.appstudio.openshift.io/sha=abc123", "appstudio.openshift.io/application=acs"]
    // Output: "data.metadata.labels['pac.test.appstudio.openshift.io/sha']=='abc123' && data.metadata.labels['appstudio.openshift.io/application']=='acs'"

    filters := make([]string, len(labels))
    for i, label := range labels {
        parts := strings.SplitN(label, "=", 2)
        if len(parts) == 2 {
            // CEL syntax: data.metadata.labels['key']=='value'
            filters[i] = fmt.Sprintf("data.metadata.labels['%s']=='%s'", parts[0], parts[1])
        }
    }
    return strings.Join(filters, " && ")
}
```

### 5. Update Konflux Client

**File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/client.go`

**Replace the Client struct:**
```go
type Client struct {
    tektonResultsClient *tektonresults.Client
}

func NewClient() (*Client, error) {
    trClient, err := tektonresults.NewClient()
    if err != nil {
        return nil, fmt.Errorf("failed to create Tekton Results client: %w", err)
    }
    return &Client{tektonResultsClient: trClient}, nil
}
```

**File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/releases.go`

**Update GetRelease and GetReleases to delegate:**
```go
func (c *Client) GetReleases(labels []string) ([]Release, error) {
    return c.tektonResultsClient.GetReleases(labels)
}

func (c *Client) GetRelease(name string) (Release, error) {
    return c.tektonResultsClient.GetRelease(name)
}
```

**Remove these functions entirely:**
- `getK8sCommand()` - No longer needed
- `runCommand()` - No longer needed
- `getK8sCommandWithJSONOutput()` - No longer needed
- `getK8sCommandWithPlainOutput()` - No longer needed
- `synchronizedBuffer` - No longer needed

**Keep temporarily (for other methods not yet migrated):**
- We need to keep KubeArchive integration for other methods not yet migrated.

### 6. Update Data Models

**File:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/models.go`

Define Tekton Results API response structures:

```go
type RecordsResponse struct {
    Records       []Record `json:"records"`
    NextPageToken string   `json:"nextPageToken"`
}

type Record struct {
    Name string     `json:"name"`
    ID   string     `json:"id"`
    Data RecordData `json:"data"`
}

type RecordData struct {
    Type       string                 `json:"@type"`
    Kind       string                 `json:"kind"`
    APIVersion string                 `json:"apiVersion"`
    Metadata   konflux.Metadata       `json:"metadata"`
    Spec       konflux.ReleaseSpec    `json:"spec"`
    Status     *konflux.ReleaseStatus `json:"status,omitempty"`
}
```

**No changes needed to:** `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/models.go`
- Release, ReleaseSpec, ReleaseStatus, Metadata structures remain the same
- Helper methods (GetState, GetTarget, GetURL, etc.) unchanged

---

## Critical Files to Modify

1. **Configuration:**
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/config/config.go` - Add Tekton Results endpoint constants

2. **New Files (Tekton Results Client):**
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/client.go` - HTTP client with kubeconfig auth
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/releases.go` - GetRelease/GetReleases implementation
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/models.go` - Tekton Results API models
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/client_test.go` - Unit tests for client
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/tektonresults/releases_test.go` - Unit tests for releases

3. **Refactor Konflux Client:**
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/client.go` - Remove subprocess code, delegate to tektonresults
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/releases.go` - Update to delegate to Tekton Results client

4. **Update Tests:**
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/client_test.go` - Update to test new HTTP-based approach
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/konflux/releases_test.go` - Ensure helper methods still work

5. **Update Main Client:**
   - `/Users/tmartens/projects/src/github.com/stackrox/relacs/pkg/client/client.go` - Update error messages (remove KubeArchive references)

---

## Testing Strategy

### Unit Tests

**1. Tekton Results Client Tests** (`client_test.go`):
- Test kubeconfig token extraction
- Test endpoint configuration (env var vs default)
- Test HTTP client initialization
- Test auth header construction

**2. Tekton Results Releases Tests** (`releases_test.go`):
- Mock Tekton Results API responses using `gock` (already used in codebase)
- Test GetReleases with label filters
- Test GetRelease by name
- Test pagination handling
- Test error cases (404, 401, network errors)

**3. Konflux Client Tests** (update existing):
- Test delegation to Tekton Results client
- Remove subprocess execution tests
- Keep Release helper method tests (GetState, GetTarget, etc.)

### Integration Tests

**Manual Testing Checklist:**
1. ✅ Set `TEKTON_RESULTS_API_ENDPOINT` to cluster endpoint
2. ✅ Run `relacs konflux get-release <name>` - verify Release retrieved
3. ✅ Run with valid kubeconfig context - verify auth works
4. ✅ Test label filtering with multiple labels
5. ✅ Test pagination with >1000 releases
6. ✅ Verify all existing commands still work (no breaking changes to Release model)

### Mock Examples

```go
// Example gock mock for GetReleases
gock.New("https://tekton-results-tekton-results.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com").
    Get("/apis/results.tekton.dev/v1alpha2/parents/rh-acs-tenant/results/-/records").
    MatchParam("filter", "data.metadata.labels['appstudio.openshift.io/application']=='acs'").
    MatchParam("page_size", "100").
    Reply(200).
    JSON(tektonresults.RecordsResponse{
        Records: []tektonresults.Record{
            {
                Name: "rh-acs-tenant/records/release-xyz",
                ID: "release-xyz",
                Data: tektonresults.RecordData{
                    Kind: "Release",
                    APIVersion: "appstudio.redhat.com/v1alpha1",
                    Metadata: konflux.Metadata{Name: "rhacs--4-10-1--stage-1"},
                    Spec: konflux.ReleaseSpec{...},
                    Status: &konflux.ReleaseStatus{...},
                },
            },
        },
        NextPageToken: "",
    })

// Example gock mock for GetRelease by name
gock.New("https://tekton-results-tekton-results.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com").
    Get("/apis/results.tekton.dev/v1alpha2/parents/rh-acs-tenant/results/-/records").
    MatchParam("filter", "data.metadata.name=='rhacs--4-10-1--stage-1'").
    MatchParam("page_size", "10").
    Reply(200).
    JSON(tektonresults.RecordsResponse{
        Records: []tektonresults.Record{
            {
                Name: "rh-acs-tenant/records/release-xyz",
                ID: "release-xyz",
                Data: tektonresults.RecordData{
                    Kind: "Release",
                    APIVersion: "appstudio.redhat.com/v1alpha1",
                    Metadata: konflux.Metadata{Name: "rhacs--4-10-1--stage-1"},
                    Spec: konflux.ReleaseSpec{...},
                    Status: &konflux.ReleaseStatus{...},
                },
            },
        },
        NextPageToken: "",
    })
```

---

## Migration Path

### Phase 1: GetRelease and GetReleases (This Plan)
1. ✅ Add Tekton Results endpoint configuration
2. ✅ Implement Tekton Results HTTP client with kubeconfig auth
3. ✅ Implement GetRelease and GetReleases
4. ✅ Write comprehensive unit tests
5. ✅ Update Konflux client to delegate
6. ✅ Remove KubeArchive subprocess code
7. ✅ Manual integration testing

### Phase 2: Other Methods (Future Work)
- GetSnapshots
- GetPipelineRun / GetPipelineRuns
- GetPodLogs (may require different API endpoint)
- GetCurrentUsername (refactor to use Kube API directly)

---

## Known Issues & Mitigations

### API Performance

The Tekton Results API has known performance issues:
- Queries can take 60+ seconds, especially for large result sets
- Unfiltered queries or broad filters may timeout with "Empty reply from server"
- Filtering by specific labels (e.g., release name, application) improves performance

**Mitigations:**
1. Always use specific filters when possible (name, commit SHA, etc.)
2. Use reasonable `page_size` (100 recommended, not max 10000)
3. Implement retry logic with exponential backoff for transient failures
4. Consider caching frequently accessed releases
5. HTTP client configured with 180s timeout (generous but necessary)

### Future Optimization Opportunities

If API performance remains an issue:
- Add local caching layer (in-memory or Redis)
- Implement background refresh for commonly accessed data
- Consider incremental sync strategy (only fetch recent releases)
- Monitor API latency and add metrics/alerting

---

## Rollback Plan

If Tekton Results API integration fails:
- Revert changes to konflux client
- Restore subprocess-based implementation
- Keep `KUBECTL_PLUGIN_KA_HOST` requirement

**Note:** Since this is a hard cutover, rollback requires reverting the full PR. Test thoroughly before merging.

---

## Dependencies & Prerequisites

**Required:**
- Tekton Results API service running in the cluster at configurable endpoint
- Valid kubeconfig with authenticated context (same as current kubectl/oc usage)

**Optional:**
- `TEKTON_RESULTS_API_ENDPOINT` environment variable (defaults to in-cluster service)

**Removed:**
- `KUBECTL_PLUGIN_KA_HOST` environment variable (no longer needed)
- KubeArchive kubectl plugin dependency

---

## Success Criteria

- ✅ GetRelease and GetReleases work identically to KubeArchive version
- ✅ Authentication works with kubeconfig (consistent with current kubectl usage)
- ✅ Pagination works correctly for >1000 releases
- ✅ All existing unit tests pass
- ✅ New Tekton Results client has >80% test coverage
- ✅ No breaking changes to Release model or helper methods
- ✅ Documentation updated (remove KubeArchive references)
- ✅ All `relacs` CLI commands using GetRelease/GetReleases continue working
