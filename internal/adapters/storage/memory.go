package storage

import (
	"cmp"
	"context"
	"fmt"
	"item-packing-calculator/internal/core/ports"
	"slices"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("storage-adapter")

// check MemoryStorage implements ports.PackStorage
var _ ports.PackStorage = (*MemoryStorage)(nil)

// MemoryStorage in an in-memory implementation of ports.PackStorage
type MemoryStorage struct {
	mu    sync.RWMutex
	sizes []int
}

// NewMemoryStorage creates a new instance of MemoryStorage
func NewMemoryStorage(sizes []int) *MemoryStorage {
	// create a copy to prevent mutating the caller's original slice
	sorted := make([]int, len(sizes))
	copy(sorted, sizes)

	// sort the slice in descending order to help with selecting the largest packs first
	slices.SortFunc(sorted, func(a, b int) int {
		return cmp.Compare(b, a)
	})

	return &MemoryStorage{
		sizes: sorted,
	}
}

// GetSizes returns the stored pack sizes
func (m *MemoryStorage) GetSizes(ctx context.Context) []int {
	ctx, span := tracer.Start(ctx, "MemoryStorage.GetSizes")
	defer span.End()

	// RLock ensures multiple goroutines can read the sizes at the same time
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]int, len(m.sizes))
	copy(result, m.sizes)

	return result
}

// AddSize adds a size to the active list in memory
func (m *MemoryStorage) SaveSize(ctx context.Context, size int) error {
	ctx, span := tracer.Start(ctx, "MemoryStorage.SaveSize")
	defer span.End()

	span.SetAttributes(attribute.Int("pack.size_attempted", size))

	m.mu.Lock()
	defer m.mu.Unlock()

	if slices.Contains(m.sizes, size) {
		err := fmt.Errorf("pack size %d already exists", size)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	m.sizes = append(m.sizes, size)

	// Re-sort descending after insertion
	slices.SortFunc(m.sizes, func(a, b int) int {
		return cmp.Compare(b, a)
	})

	span.SetStatus(codes.Ok, "size added")

	return nil
}

// RemoveSize deletes a pack size from the active list in memory
func (m *MemoryStorage) DeleteSize(ctx context.Context, size int) error {
	ctx, span := tracer.Start(ctx, "MemoryStorage.DeleteSize")
	defer span.End()

	span.SetAttributes(attribute.Int("pack.size_attempted", size))

	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.Index(m.sizes, size)
	if idx == -1 {
		err := fmt.Errorf("pack size %d not found", size)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Fast slice deletion as slices.Delete does the same shifting under the hood
	m.sizes = append(m.sizes[:idx], m.sizes[idx+1:]...)

	span.SetStatus(codes.Ok, "size removed")
	return nil
}
