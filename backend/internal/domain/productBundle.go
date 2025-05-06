package domain

type ProductBundle struct {
	BundleID    int `json:"bundle_id"`
	ProductID   int `json:"product_id"`
	LayerNumber int `json:"layer_no"`
}
