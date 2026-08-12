package productlifecycle

// ProductLifecycle represents the top-level response from the Red Hat Product Lifecycle API.
type ProductLifecycle struct {
	Data []ProductLifecycleData `json:"data"`
}

// ProductLifecycleData contains lifecycle information for a single product.
type ProductLifecycleData struct {
	Uuid     string           `json:"uuid"`
	Name     string           `json:"name"`
	Versions []ProductVersion `json:"versions"`
}

// ProductVersion represents a specific version of a product with its support type.
type ProductVersion struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
