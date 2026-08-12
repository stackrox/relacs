package helpers

import "github.com/h2non/gock"

// ProductVersionData represents a product version for mocking
type ProductVersionData struct {
	Name string
	Type string
}

// MockProductLifecycleAPI mocks the Red Hat Product Lifecycle API
func MockProductLifecycleAPI(productName string, versions []ProductVersionData) {
	versionsJSON := make([]map[string]interface{}, len(versions))
	for i, v := range versions {
		versionsJSON[i] = map[string]interface{}{
			"name": v.Name,
			"type": v.Type,
		}
	}

	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		JSON(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":     "test-uuid",
					"name":     productName,
					"versions": versionsJSON,
				},
			},
		})
}

// MockProductLifecycleAPINotFound mocks a 404 or empty response from Product Lifecycle API
func MockProductLifecycleAPINotFound(productName string) {
	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		JSON(map[string]interface{}{
			"data": []map[string]interface{}{},
		})
}

// MockProductLifecycleAPIError mocks an error response from Product Lifecycle API
func MockProductLifecycleAPIError(productName string) {
	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(500).
		JSON(map[string]interface{}{
			"message": "Internal Server Error",
		})
}

// MockProductLifecycleAPIMultipleProducts mocks a response with multiple products (ambiguous result)
func MockProductLifecycleAPIMultipleProducts(productName string, versions []ProductVersionData) {
	versionsJSON := make([]map[string]interface{}, len(versions))
	for i, v := range versions {
		versionsJSON[i] = map[string]interface{}{
			"name": v.Name,
			"type": v.Type,
		}
	}

	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		JSON(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":     "test-uuid-1",
					"name":     productName,
					"versions": versionsJSON,
				},
				{
					"uuid":     "test-uuid-2",
					"name":     productName + " Duplicate",
					"versions": versionsJSON,
				},
			},
		})
}

// MockProductLifecycleAPIMissingDataField mocks a response with missing data field entirely
func MockProductLifecycleAPIMissingDataField(productName string) {
	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		JSON(map[string]interface{}{
			// No "data" field at all
			"status": "ok",
		})
}

// MockProductLifecycleAPINullData mocks a response with null data field
func MockProductLifecycleAPINullData(productName string) {
	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		BodyString(`{"data": null}`)
}

// MockProductLifecycleAPIMalformedJSON mocks a malformed JSON response
func MockProductLifecycleAPIMalformedJSON(productName string) {
	gock.New("https://access.redhat.com").
		Get("/product-life-cycles/api/v2/products").
		MatchParam("name", productName).
		Reply(200).
		BodyString(`{"data": [invalid json}`)
}
