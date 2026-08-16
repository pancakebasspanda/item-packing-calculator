package service

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"item-packing-calculator/internal/core/domain"
	mockPorts "item-packing-calculator/internal/core/ports/mocks"
)

func TestCalculatorService_Calculate(t *testing.T) {
	ctx := context.Background()

	standardSizes := []int{5000, 2000, 1000, 500, 250}

	tests := []struct {
		name                string
		orderQuantity       int
		storageMockOutcomes func(storageMock *mockPorts.MockPackStorage)
		expectedResult      domain.CalculationResult
		expectedErr         string
	}{
		{
			name:          "returns error when order quantity is zero",
			orderQuantity: 0,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedResult: domain.CalculationResult{},
			expectedErr:    "order quantity must be greater than zero",
		},
		{
			name:          "returns error when order quantity is negative",
			orderQuantity: -15,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedResult: domain.CalculationResult{},
			expectedErr:    "order quantity must be greater than zero",
		},
		{
			name:          "returns error when no pack sizes exist in storage",
			orderQuantity: 500,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					GetSizes(gomock.Any()).
					Times(1).
					Return([]int{})
			},
			expectedResult: domain.CalculationResult{},
			expectedErr:    "no pack sizes available in the system",
		},
		{
			name:          "successfully calculates exact match with a single pack",
			orderQuantity: 250,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					GetSizes(gomock.Any()).
					Times(1).
					Return(standardSizes)
			},
			expectedResult: domain.CalculationResult{
				OrderQuantity: 250,
				TotalItems:    250,
				TotalPacks:    1,
				Packs: []domain.PackResult{
					{PackSize: 250, Quantity: 1},
				},
			},
			expectedErr: "",
		},
		{
			name:          "successfully calculates optimal combination minimizing total items first",
			orderQuantity: 251,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					GetSizes(gomock.Any()).
					Times(1).
					Return(standardSizes)
			},
			expectedResult: domain.CalculationResult{
				OrderQuantity: 251,
				TotalItems:    500,
				TotalPacks:    1,
				Packs: []domain.PackResult{
					{PackSize: 500, Quantity: 1},
				},
			},
			expectedErr: "",
		},
		{
			name:          "successfully calculates multi-pack breakdown for complex quantity",
			orderQuantity: 501,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					GetSizes(gomock.Any()).
					Times(1).
					Return(standardSizes)
			},
			expectedResult: domain.CalculationResult{
				OrderQuantity: 501,
				TotalItems:    750,
				TotalPacks:    2,
				Packs: []domain.PackResult{
					{PackSize: 500, Quantity: 1},
					{PackSize: 250, Quantity: 1},
				},
			},
			expectedErr: "",
		},
		{
			name:          "successfully optimizes massive bulk order without stack overflow",
			orderQuantity: 50001,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					GetSizes(gomock.Any()).
					Times(1).
					Return(standardSizes)
			},
			expectedResult: domain.CalculationResult{
				OrderQuantity: 50001,
				TotalItems:    50250,
				TotalPacks:    11,
				Packs: []domain.PackResult{
					{PackSize: 5000, Quantity: 10},
					{PackSize: 250, Quantity: 1},
				},
			},
			expectedErr: "",
		},
		{
			name:          "returns error when order quantity exceeds safe integer limit",
			orderQuantity: math.MaxInt,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				// No storage calls expected because validation fails early
			},
			expectedResult: domain.CalculationResult{},
			expectedErr:    "order quantity exceeds maximum supported limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			storageMock := mockPorts.NewMockPackStorage(mockController)
			tt.storageMockOutcomes(storageMock)

			svc := NewCalculatorService(storageMock)
			result, err := svc.Calculate(ctx, tt.orderQuantity)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
				assert.Equal(t, domain.CalculationResult{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCalculatorService_SaveSize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		size                int
		storageMockOutcomes func(storageMock *mockPorts.MockPackStorage)
		expectedErr         string
	}{
		{
			name: "returns error when pack size is zero",
			size: 0,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedErr: "pack size must be greater than zero",
		},
		{
			name: "returns error when pack size is negative",
			size: -100,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedErr: "pack size must be greater than zero",
		},
		{
			name: "successfully adds new pack size when storage already contains sizes",
			size: 3000,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					SaveSize(gomock.Any(), 3000).
					Times(1).
					Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "returns error when adding a size that already exists in storage",
			size: 1000,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					SaveSize(gomock.Any(), 1000).
					Times(1).
					Return(errors.New("pack size 1000 already exists"))
			},
			expectedErr: "pack size 1000 already exists",
		},
		{
			name: "returns error when storage adapter fails to insert size",
			size: 3000,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					SaveSize(gomock.Any(), 3000).
					Times(1).
					Return(errors.New("storage writing failure"))
			},
			expectedErr: "storage writing failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			storageMock := mockPorts.NewMockPackStorage(mockController)
			tt.storageMockOutcomes(storageMock)

			svc := NewCalculatorService(storageMock)
			err := svc.AddSize(ctx, tt.size)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculatorService_DeleteSize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		size                int
		storageMockOutcomes func(storageMock *mockPorts.MockPackStorage)
		expectedErr         string
	}{
		{
			name: "returns error when pack size is zero",
			size: 0,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedErr: "invalid pack size",
		},
		{
			name: "returns error when pack size is negative",
			size: -500,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
			},
			expectedErr: "invalid pack size",
		},
		{
			name: "successfully removes an existing pack size from storage",
			size: 500,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					DeleteSize(gomock.Any(), 500).
					Times(1).
					Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "returns error when attempting to remove size that does not exist in storage",
			size: 9999,
			storageMockOutcomes: func(storageMock *mockPorts.MockPackStorage) {
				storageMock.EXPECT().
					DeleteSize(gomock.Any(), 9999).
					Times(1).
					Return(errors.New("pack size 9999 not found in storage"))
			},
			expectedErr: "pack size 9999 not found in storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			storageMock := mockPorts.NewMockPackStorage(mockController)
			tt.storageMockOutcomes(storageMock)

			svc := NewCalculatorService(storageMock)
			err := svc.RemoveSize(ctx, tt.size)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
