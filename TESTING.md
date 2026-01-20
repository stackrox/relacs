# E2E Testing Guide for relacs CLI

This guide explains how to write end-to-end (e2e) tests for relacs CLI commands and how to practice test-driven development (TDD).

## Overview

E2E tests in relacs verify that CLI commands work correctly by:
- Executing command functions directly (not as subprocesses)
- Mocking external API responses using [gock](https://github.com/h2non/gock)
- Testing the full command flow from input to output
- Verifying error handling and edge cases

## Test Structure

Tests are organized by command group in the `tests/` directory:

```
tests/
├── helpers/
│   ├── client.go      # Test client creation helpers
│   └── command.go     # Command mocking helpers
└── github/
    └── get_commit_for_tag_test.go  # Example: GitHub command tests
```

## Running Tests

### Run all e2e tests
```bash
make e2e-test
```

### Run tests for a specific package
```bash
go test ./tests/github/... -v
```

### Run a specific test
```bash
go test ./tests/github/... -v -run TestGetCommitForTag_LightweightTag
```

## Adding New E2E Tests

### Step 1: Create Test File

Create a test file in the appropriate directory under `tests/`:

```go
package github

import (
    "context"
    "net/http"
    "testing"

    "github.com/h2non/gock"
    "github.com/spf13/cobra"
    "github.com/stackrox/relacs/cmd/relacs/github"
    "github.com/stackrox/relacs/pkg/client"
    "github.com/stackrox/relacs/tests/helpers"
    "github.com/stretchr/testify/suite"
)

func TestYourCommandTestSuite(t *testing.T) {
    suite.Run(t, new(YourCommandTestSuite))
}

type YourCommandTestSuite struct {
    suite.Suite
}
```

### Step 2: Set Up Test Suite

Add teardown to clean up mocks:

```go
func (s *YourCommandTestSuite) TearDownTest() {
    // Clean up gock mocks after each test
    gock.Off()
    gock.Clean()
}
```

### Step 3: Create Helper Methods

Create reusable helper methods for your test suite:

#### Setup Test Client
```go
func (s *YourCommandTestSuite) setupTestClient() *client.Client {
    httpClient := &http.Client{Transport: &http.Transport{}}
    gock.InterceptClient(httpClient)
    testClient, err := helpers.NewTestClient(httpClient)
    s.Require().NoError(err)
    return testClient
}
```

#### Create Mock Command
```go
func (s *YourCommandTestSuite) createCommand(flag1, flag2 string) *cobra.Command {
    cmd := helpers.NewMockCommand()
    s.Require().NoError(helpers.SetFlag(cmd, "flag1", flag1))
    s.Require().NoError(helpers.SetFlag(cmd, "flag2", flag2))
    return cmd
}
```

#### Mock API Responses
```go
func (s *YourCommandTestSuite) mockSuccessResponse() {
    gock.New("https://api.example.com").
        Get("/endpoint").
        Reply(200).
        JSON(map[string]interface{}{
            "field": "value",
        })
}

func (s *YourCommandTestSuite) mockErrorResponse() {
    gock.New("https://api.example.com").
        Get("/endpoint").
        Reply(404).
        JSON(map[string]interface{}{
            "message": "Not Found",
        })
}
```

#### Verify Output
```go
func (s *YourCommandTestSuite) verifyOutput(output any, expectedField string) {
    result, ok := output.(YourOutputType)
    s.True(ok)
    s.Equal(expectedField, result.Field)
}
```

### Step 4: Write Test Cases

Write test cases following the Arrange-Act-Assert pattern:

```go
func (s *YourCommandTestSuite) TestYourCommand_Success() {
    // Arrange: Set up mocks and test data
    s.mockSuccessResponse()
    testClient := s.setupTestClient()
    cmd := s.createCommand("value1", "value2")

    // Act: Execute the command function
    output, err := github.RunYourCommand(cmd, []string{}, testClient, context.Background())

    // Assert: Verify results
    s.NoError(err)
    s.NotNil(output)
    s.verifyOutput(output, "expected")
    s.True(gock.IsDone()) // Verify all mocks were called
}

func (s *YourCommandTestSuite) TestYourCommand_Error() {
    // Arrange
    s.mockErrorResponse()
    testClient := s.setupTestClient()
    cmd := s.createCommand("invalid", "values")

    // Act
    output, err := github.RunYourCommand(cmd, []string{}, testClient, context.Background())

    // Assert
    s.Error(err)
    s.Nil(output)
    s.Contains(err.Error(), "expected error message")
    s.True(gock.IsDone())
}
```

### Step 5: Export Command Function for Testing

If the command function is not exported, add an exported wrapper:

```go
// In cmd/relacs/github/your_command.go

// RunYourCommand is the exported version for testing purposes
func RunYourCommand(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
    return runYourCommand(cmd, args, c, ctx)
}

func runYourCommand(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
    // Implementation
}
```

## Test-Driven Development (TDD) Workflow

### 1. Write Test First (Red)

Start by writing a failing test that describes the desired behavior:

```go
func (s *YourCommandTestSuite) TestNewFeature() {
    // This test will fail because the feature doesn't exist yet
    s.mockNewFeatureResponse()
    testClient := s.setupTestClient()
    cmd := s.createCommand("input")

    output, err := github.RunNewFeature(cmd, []string{}, testClient, context.Background())

    s.NoError(err)
    s.NotNil(output)
    // ... assertions
}
```

Run the test to confirm it fails:
```bash
go test ./tests/github/... -v -run TestNewFeature
```

### 2. Implement Minimum Code (Green)

Implement the minimal code needed to make the test pass:

```go
func runNewFeature(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
    // Minimal implementation to pass the test
    return NewFeatureOutput{}, nil
}
```

Run the test again to confirm it passes:
```bash
go test ./tests/github/... -v -run TestNewFeature
```

### 3. Refactor (Refactor)

Improve the code while keeping tests passing:

- Extract helper functions
- Improve error handling
- Add edge case tests
- Refactor for better maintainability

### 4. Add More Test Cases

Add tests for edge cases and error scenarios:

```go
func (s *YourCommandTestSuite) TestNewFeature_EdgeCase() { ... }
func (s *YourCommandTestSuite) TestNewFeature_ErrorHandling() { ... }
```

## Best Practices

### 1. Keep Tests DRY (Don't Repeat Yourself)

- Extract common setup into helper methods
- Reuse mock functions for similar scenarios
- Create verification helpers for common assertions

### 2. Use Descriptive Test Names

Good:
```go
TestGetCommitForTag_LightweightTag
TestGetCommitForTag_TagNotFound
TestGetCommitForTag_InvalidRepository
```

Bad:
```go
TestGetCommitForTag1
TestGetCommitForTag2
TestError
```

### 3. Follow Arrange-Act-Assert Pattern

```go
func (s *Suite) TestSomething() {
    // Arrange: Set up test data and mocks
    s.mockAPI()
    client := s.setupClient()

    // Act: Execute the code under test
    result, err := function(client)

    // Assert: Verify the results
    s.NoError(err)
    s.Equal(expected, result)
}
```

### 4. Verify All Mocks Were Called

Always check that all gock mocks were used:

```go
s.True(gock.IsDone())
```

### 5. Test Both Success and Failure Cases

- Happy path scenarios
- Error conditions (404, 500, network errors)
- Edge cases (empty inputs, boundary values)
- Invalid inputs

### 6. Mock External Dependencies

Never make real API calls in tests. Always use gock to mock:
- GitHub API responses
- Konflux API responses
- Any external HTTP services

### 7. Keep Tests Independent

Each test should be able to run independently:
- Don't rely on test execution order
- Clean up mocks in `TearDownTest()`
- Use unique test data for each test

## Example: Complete Test Suite

See `tests/github/get_commit_for_tag_test.go` for a complete example that demonstrates:
- Test suite structure
- Helper method organization
- Multiple test scenarios
- Mock API responses
- Output verification

## Common Patterns

### Mocking GitHub API

```go
gock.New("https://api.github.com").
    Get("/repos/org/repo/endpoint").
    Reply(200).
    JSON(map[string]interface{}{
        "field": "value",
    })
```

### Mocking Error Responses

```go
gock.New("https://api.github.com").
    Get("/repos/org/repo/endpoint").
    Reply(404).
    JSON(map[string]interface{}{
        "message": "Not Found",
    })
```

### Testing Command Flags

```go
cmd := helpers.NewMockCommand()
s.Require().NoError(helpers.SetFlag(cmd, "flag-name", "flag-value"))
```

### Verifying Output Structure

```go
output, ok := result.(ExpectedType)
s.True(ok, "Output should be of type ExpectedType")
s.Equal(expectedValue, output.Field)
```

## Troubleshooting

### Tests fail with "gock: cannot match any request"

- Ensure `gock.InterceptClient(httpClient)` is called before creating the client
- Verify the HTTP client is passed to `helpers.NewTestClient()`
- Check that mock URLs match exactly (including protocol and path)

### Tests fail with "flag redefined"

- Don't define flags multiple times in the same command
- Use `helpers.SetFlag()` which handles both definition and setting

### Mocks not being called

- Ensure `gock.IsDone()` is checked at the end of each test
- Verify the HTTP client used by the command is the one intercepted by gock
- Check that the base URL in mocks matches the actual API calls

## CI Integration

E2E tests run automatically in CI via the `e2e-test` job in `.github/workflows/build.yaml`. The tests run in parallel with unit tests and must pass before code can be merged.

## Additional Resources

- [testify documentation](https://github.com/stretchr/testify)
- [gock documentation](https://github.com/h2non/gock)
- [Go testing best practices](https://golang.org/doc/effective_go#testing)
