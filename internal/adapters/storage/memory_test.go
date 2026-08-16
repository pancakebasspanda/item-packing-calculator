package storage

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryStorage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                   string
		input                  []int
		expectedStorage        []int
		expectedInputUnchanged []int
	}{
		{
			name:                   "unsorted input slice is sorted descending",
			input:                  []int{250, 5000, 1000, 500, 2000},
			expectedStorage:        []int{5000, 2000, 1000, 500, 250},
			expectedInputUnchanged: []int{250, 5000, 1000, 500, 2000},
		},
		{
			name:                   "already sorted descending input slice remains descending",
			input:                  []int{5000, 2000, 1000, 500, 250},
			expectedStorage:        []int{5000, 2000, 1000, 500, 250},
			expectedInputUnchanged: []int{5000, 2000, 1000, 500, 250},
		},
		{
			name:                   "ascending sorted input slice is converted to descending",
			input:                  []int{250, 500, 1000, 2000, 5000},
			expectedStorage:        []int{5000, 2000, 1000, 500, 250},
			expectedInputUnchanged: []int{250, 500, 1000, 2000, 5000},
		},
		{
			name:                   "input slice with duplicates preserves elements sorted descending",
			input:                  []int{500, 1000, 500, 2000},
			expectedStorage:        []int{2000, 1000, 500, 500},
			expectedInputUnchanged: []int{500, 1000, 500, 2000},
		},
		{
			name:                   "single element slice",
			input:                  []int{500},
			expectedStorage:        []int{500},
			expectedInputUnchanged: []int{500},
		},
		{
			name:                   "empty input slice",
			input:                  []int{},
			expectedStorage:        []int{},
			expectedInputUnchanged: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStorage(tt.input)

			// sorted descending check
			assert.Equal(t, tt.expectedStorage, store.GetSizes(ctx))

			// check input slice was not mutated
			assert.Equal(t, tt.expectedInputUnchanged, tt.input)
		})
	}
}

func TestMemoryStorage_GetSizes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		initialSizes  []int
		expectedSizes []int
	}{
		{
			name:          "returns pack sizes sorted descending",
			initialSizes:  []int{1000, 5000, 500},
			expectedSizes: []int{5000, 1000, 500},
		},
		{
			name:          "returns empty slice when initialized empty",
			initialSizes:  []int{},
			expectedSizes: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStorage(tt.initialSizes)
			result := store.GetSizes(ctx)
			assert.Equal(t, tt.expectedSizes, result)
		})
	}

	// check that the returned slice is a copy of the original slice and doest affect the input
	t.Run("ensures returned slice is an independent copy preventing external mutation", func(t *testing.T) {
		store := NewMemoryStorage([]int{1000, 500})
		result := store.GetSizes(ctx)

		// mutate returned slice
		result[0] = 99999

		// verify internal storage remains untouched
		assert.Equal(t, []int{1000, 500}, store.GetSizes(ctx))
	})
}

func TestMemoryStorage_AddSize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		initialSizes  []int
		sizeToAdd     int
		expectedSizes []int
		expectedErr   string
	}{
		{
			name:          "successfully adds new size and maintains descending order",
			initialSizes:  []int{5000, 1000, 250},
			sizeToAdd:     2000,
			expectedSizes: []int{5000, 2000, 1000, 250},
			expectedErr:   "",
		},
		{
			name:          "successfully adds size to empty storage",
			initialSizes:  []int{},
			sizeToAdd:     500,
			expectedSizes: []int{500},
			expectedErr:   "",
		},
		{
			name:          "returns error when adding a duplicate pack size",
			initialSizes:  []int{5000, 1000, 500},
			sizeToAdd:     1000,
			expectedSizes: []int{5000, 1000, 500},
			expectedErr:   "pack size 1000 already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStorage(tt.initialSizes)
			err := store.SaveSize(ctx, tt.sizeToAdd)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSizes, store.GetSizes(ctx))
		})
	}
}

func TestMemoryStorage_RemoveSize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		initialSizes  []int
		sizeToRemove  int
		expectedSizes []int
		expectedErr   string
	}{
		{
			name:          "successfully removes existing pack size",
			initialSizes:  []int{5000, 2000, 1000, 500},
			sizeToRemove:  2000,
			expectedSizes: []int{5000, 1000, 500},
			expectedErr:   "",
		},
		{
			name:          "returns error when removing non-existent pack size",
			initialSizes:  []int{5000, 1000},
			sizeToRemove:  250,
			expectedSizes: []int{5000, 1000},
			expectedErr:   "pack size 250 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStorage(tt.initialSizes)
			err := store.DeleteSize(ctx, tt.sizeToRemove)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSizes, store.GetSizes(ctx))
		})
	}
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	initialSizes := []int{5000, 2000, 1000, 500, 250}
	store := NewMemoryStorage(initialSizes)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // read, write(add), write(remove)

	// concurrent readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			sizes := store.GetSizes(ctx)
			assert.NotEmpty(t, sizes)
		}()
	}

	// concurrent writers add size
	for i := 0; i < goroutines; i++ {
		sizeToAdd := 10000 + i
		go func(s int) {
			defer wg.Done()
			err := store.SaveSize(ctx, s)
			assert.NoError(t, err)
		}(sizeToAdd)
	}

	// concurrent writers remove size
	// 1 goroutine wins, 99 fail to find it as it is already removed
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			err := store.DeleteSize(ctx, 250)
			if err != nil {
				assert.EqualError(t, err, "pack size 250 not found")
			}
		}()
	}

	wg.Wait()

	// verify slice state remains sorted descending after all operations complete
	finalSizes := store.GetSizes(ctx)
	assert.NotEmpty(t, finalSizes)
	for i := 0; i < len(finalSizes)-1; i++ {
		assert.GreaterOrEqual(t, finalSizes[i], finalSizes[i+1], "storage must remain sorted descending")
	}
}
