import { CalculationResult, ErrorResponse } from '../types/types';

const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

export const calculatePacking = async (quantity: number): Promise<CalculationResult> => {
    const response = await fetch(`${API_BASE}/packing/calculate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orderQuantity: quantity }),
    });

    if (!response.ok) {
        const errorData: ErrorResponse = await response.json().catch(() => ({ error: 'Unknown API error' }));
        throw new Error(errorData.error || 'Failed to calculate packing');
    }

    return response.json();
};