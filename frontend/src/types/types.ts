export interface PackResult {
    packSize: number;
    quantity: number;
}

export interface CalculationResult {
    orderQuantity: number;
    totalItems: number;
    totalPacks: number;
    packs: PackResult[];
}

export interface ErrorResponse {
    error: string;
}