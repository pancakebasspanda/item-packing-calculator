package domain

// PackResult is the number of packs and the specific pack size
type PackResult struct {
	PackSize int `json:"packSize"`
	Quantity int `json:"quantity"`
}

// CalculationResult represents output needed for an order
type CalculationResult struct {
	OrderQuantity int          `json:"orderQuantity"`
	TotalItems    int          `json:"totalItems"`
	TotalPacks    int          `json:"totalPacks"`
	Packs         []PackResult `json:"packs"`
}
