package ports

import (
	"context"
	"item-packing-calculator/internal/core/domain"
)

// defines the contracts of our application

// PackStorage defines methods for acting with our saved pack sizes
type PackStorage interface {
	GetSizes(ctx context.Context) []int
	SaveSize(ctx context.Context, size int) error
	DeleteSize(ctx context.Context, size int) error
}

// PackService defines core operations for order calculation and pack size management
type PackService interface {
	Calculate(ctx context.Context, orderQuantity int) (domain.CalculationResult, error)
	AddSize(ctx context.Context, size int) error
	RemoveSize(ctx context.Context, size int) error
}
