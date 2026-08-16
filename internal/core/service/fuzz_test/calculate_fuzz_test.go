package service

import (
	"context"
	"testing"

	"item-packing-calculator/internal/adapters/storage"
	"item-packing-calculator/internal/core/service"
)

// FuzzCalculate test programmatically calling calculate with many(millions) integers, it automatically checks for panics
func FuzzCalculate(f *testing.F) {
	ctx := context.Background()

	f.Add(1)
	f.Add(250)
	f.Add(251)
	f.Add(501)
	f.Add(0)
	f.Add(-1)

	// 2. Setup (runs once)
	initialSizes := []int{5000, 2000, 1000, 500, 250}
	store := storage.NewMemoryStorage(initialSizes)
	calculatorService := service.NewCalculatorService(store)

	// fuzz target
	f.Fuzz(func(t *testing.T, orderQuantity int) {
		result, err := calculatorService.Calculate(ctx, orderQuantity)

		// negatives and zero must always return an error
		if orderQuantity <= 0 {
			if err == nil {
				t.Errorf("Fuzz failure: core logic allowed an order quantity of %d without an error", orderQuantity)
			}
			return
		}

		// The packs returned must have enough capacity for the order
		if err == nil {
			totalCapacityCalculated := 0
			for _, pack := range result.Packs {
				totalCapacityCalculated += pack.PackSize * pack.Quantity
			}

			// verify the calculated packs can actually hold the requested items
			if totalCapacityCalculated < orderQuantity {
				t.Errorf("Fuzz failure: requested %d items, but calculated packs only hold %d", orderQuantity, totalCapacityCalculated)
			}

			// check capacity
			if result.TotalItems != totalCapacityCalculated {
				t.Errorf("Fuzz failure: result.TotalItems (%d) does not match actual pack capacity (%d)", result.TotalItems, totalCapacityCalculated)
			}
		}
	})
}
