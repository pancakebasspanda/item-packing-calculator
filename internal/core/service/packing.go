package service

import (
	"context"
	"fmt"
	"item-packing-calculator/internal/adapters/api"
	"maps"
	"math"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"item-packing-calculator/internal/core/domain"
	"item-packing-calculator/internal/core/ports"
)

var tracer = otel.Tracer("calculator-service")

// check CalculatorService implements ports.PackService
var _ ports.PackService = (*CalculatorService)(nil)

// PackingSolution holds the best state for a given target inside the recursive function
type PackingSolution struct {
	totalItems int
	totalPacks int
	counts     map[int]int
}

// CalculatorService implements the calculation method defined in the PackService interface
type CalculatorService struct {
	store ports.PackStorage
}

// NewCalculatorService instantiates a new instance of the CalculatorService
func NewCalculatorService(store ports.PackStorage) *CalculatorService {
	return &CalculatorService{
		store: store,
	}
}

// Calculate determines the optimal number of packs for a given order quantity.
func (s *CalculatorService) Calculate(ctx context.Context, orderQuantity int) (domain.CalculationResult, error) {
	ctx, span := tracer.Start(ctx, "CalculatorService.Calculate")
	defer span.End()

	requestedQuantity := orderQuantity // preserve for the response

	span.SetAttributes(attribute.Int("order.requested_quantity", orderQuantity))

	if orderQuantity <= 0 {
		err := fmt.Errorf("order quantity must be greater than zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return domain.CalculationResult{}, err
	}

	if orderQuantity > api.MaxOrderQuantity {
		err := fmt.Errorf("order quantity exceeds maximum supported limit")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return domain.CalculationResult{}, err
	}

	sizes := s.store.GetSizes(ctx)
	if len(sizes) == 0 {
		err := fmt.Errorf("no pack sizes available in the system")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return domain.CalculationResult{}, err
	}

	largestSize := sizes[0] // since sizes is sorted descending

	// optimize large orders by allocating the bulk of the items into the biggest packs
	orderQuantity, biggestAvailablePacks := s.preCalculateBulkPacks(orderQuantity, largestSize)

	// find the remaining number of packs needed as per the rules
	knownSolutions := make(map[int]*PackingSolution)
	remainderSolution := s.findOptimalPacks(orderQuantity, sizes, knownSolutions)

	// merge calculations and format the response
	result := s.buildFinalResult(requestedQuantity, sizes, biggestAvailablePacks, remainderSolution)

	span.SetAttributes(
		attribute.Int("order.shipped_items", result.TotalItems),
		attribute.Int("order.shipped_packs", result.TotalPacks),
	)

	return result, nil
}

// findOptimalPacks uses a recursive memoized DFS to find the combination that
// minimizes items first, then minimizes packs.
func (s *CalculatorService) findOptimalPacks(target int, sizes []int, knownSolutions map[int]*PackingSolution) *PackingSolution {
	if target <= 0 {
		return &PackingSolution{
			totalItems: 0,
			totalPacks: 0,
			counts:     make(map[int]int),
		}
	}

	// return last optimized state
	if val, ok := knownSolutions[target]; ok {
		return val
	}

	// set worst case scenario / max
	best := &PackingSolution{
		totalItems: math.MaxInt32,
		totalPacks: math.MaxInt32,
	}

	// use recursion to find the optimal packs
	for _, size := range sizes {
		branchSolution := s.findOptimalPacks(target-size, sizes, knownSolutions)

		newTotalItems := branchSolution.totalItems + size
		newTotalPacks := branchSolution.totalPacks + 1

		// cater for rule 2 and 3
		isBetter := newTotalItems < best.totalItems || (newTotalItems == best.totalItems && newTotalPacks < best.totalPacks)

		if isBetter {
			best.totalItems = newTotalItems
			best.totalPacks = newTotalPacks
			best.counts = maps.Clone(branchSolution.counts)
			best.counts[size]++
		}
	}

	// store in memo map for future lookups
	knownSolutions[target] = best
	return best
}

// preCalculateBulkPacks reduces the order target by safely stripping out the maximum
// amount of the largest pack size, preventing stack overflow on massive numbers.
func (s *CalculatorService) preCalculateBulkPacks(orderQuantity, largestSize int) (int, map[int]int) {
	biggestAvailablePacks := make(map[int]int) // size of pack mapped to the number of them to use

	// if the order is large, we want to reduce it down as to not run out of memory or encounter stack issues
	if orderQuantity > 2*largestSize {
		numLargest := (orderQuantity - largestSize) / largestSize
		biggestAvailablePacks[largestSize] = numLargest

		// reduce the order quantity by number of the largest size packs used
		orderQuantity -= numLargest * largestSize
	}

	return orderQuantity, biggestAvailablePacks
}

// buildFinalResult merges the pre-calculated bulk items with the exact remainder calculation and formats it into the domain model.
func (s *CalculatorService) buildFinalResult(requestedQuantity int, sizes []int, biggestAvailablePacks map[int]int, remainderSolution *PackingSolution) domain.CalculationResult {
	// combine backpacks count with the remaining best combination of packs
	finalCounts := maps.Clone(biggestAvailablePacks) // maps.copy could overwrite keys so we clone instead

	for size, count := range remainderSolution.counts {
		finalCounts[size] += count
	}

	largestSize := sizes[0]
	totalItems := remainderSolution.totalItems + (biggestAvailablePacks[largestSize] * largestSize)
	totalPacks := remainderSolution.totalPacks + biggestAvailablePacks[largestSize]

	packs := make([]domain.PackResult, 0, len(finalCounts))

	for _, size := range sizes {
		if count := finalCounts[size]; count > 0 {
			packs = append(packs, domain.PackResult{
				PackSize: size,
				Quantity: count,
			})
		}
	}

	// transform to domain model
	return domain.CalculationResult{
		OrderQuantity: requestedQuantity,
		TotalItems:    totalItems,
		TotalPacks:    totalPacks,
		Packs:         packs,
	}
}

// AddSize validates and adds a new pack size to the system via the storage port
func (s *CalculatorService) AddSize(ctx context.Context, size int) error {
	if size <= 0 {
		return fmt.Errorf("pack size must be greater than zero")
	}

	// Delegate to the storage adapter
	return s.store.SaveSize(ctx, size)
}

// RemoveSize removes an existing pack size from the system via the storage port
func (s *CalculatorService) RemoveSize(ctx context.Context, size int) error {
	if size <= 0 {
		return fmt.Errorf("invalid pack size")
	}

	// Delegate to the storage adapter
	return s.store.DeleteSize(ctx, size)
}
