import { useState, useEffect } from 'react';
import { calculatePacking } from '../api/client';
import { CalculationResult } from '../types/types';
import PackageBreakdown from './PackageBreakdown';

export default function CheckoutPackingWidget() {
    const [quantityInput, setQuantityInput] = useState<string>('1');
    const [result, setResult] = useState<CalculationResult | null>(null);
    const [error, setError] = useState<string>('');
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [isWakingUp, setIsWakingUp] = useState<boolean>(false);

    useEffect(() => {
        const quantity = parseInt(quantityInput, 10);

        // If the input is completely empty (NaN from backspacing), clear the UI quietly
        if (isNaN(quantity)) {
            setResult(null);
            setError('');
            return;
        }

        // If they explicitly typed 0 or a negative number, show a helpful error immediately
        if (quantity <= 0) {
            setResult(null);
            setError('Order quantity must be greater than zero');
            return;
        }

        // max limit check
        const MAX_ITEMS = Number.MAX_SAFE_INTEGER;
        if (quantity > MAX_ITEMS) {
            setResult(null);
            setError(`Apologies, we can only process orders up to ${MAX_ITEMS.toLocaleString()} items at a time.`);
            return;
        }

        // timer variable incase the backend is in cold start uo
        let wakeupTimer: ReturnType<typeof setTimeout>;

        const fetchPacks = async () => {
            setIsLoading(true);
            setIsWakingUp(false);
            setError('');

            // if the API takes > 3 seconds
            wakeupTimer = setTimeout(() => {
                setIsWakingUp(true);
            }, 3000);

            try {
                const data = await calculatePacking(quantity);
                setResult(data);
            } catch (err) {
                if (err instanceof Error) {
                    setError(err.message);
                } else {
                    setError('An unexpected error occurred');
                }
                setResult(null);
            } finally {
                clearTimeout(wakeupTimer);
                setIsWakingUp(false);
                setIsLoading(false);
            }
        };

        // Debounce the API call by 300ms
        const debounceTimer = setTimeout(fetchPacks, 300);
        return () => {
            clearTimeout(debounceTimer);
            clearTimeout(wakeupTimer);
        }
    }, [quantityInput]);

    return (
        <div className="w-full max-w-3xl mx-auto p-6 md:p-10 bg-white rounded-xl shadow-lg border border-gray-100 transition-all">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Order Details</h2>

            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                    Quantity Required
                </label>
                <input
                    type="number"
                    min="1"
                    value={quantityInput}
                    onChange={(e) => setQuantityInput(e.target.value)}
                    onKeyDown={(e) => {
                        if (['e', 'E', '+', '-', '.'].includes(e.key)) {
                            e.preventDefault();
                        }
                    }}
                    onPaste={(e) => {
                        const pasteData = e.clipboardData.getData('text');
                        if (/[eE+\-\.]/.test(pasteData)) {
                            e.preventDefault();
                        }
                    }}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
                    placeholder="Enter amount..."
                />
            </div>

            {error && (
                <div className="mt-4 p-3 bg-red-50 text-red-700 text-sm rounded-lg border border-red-200">
                    {error}
                </div>
            )}

            {!error && (
                <PackageBreakdown
                    result={result}
                    isLoading={isLoading}
                    isWakingUp={isWakingUp}
                />
            )}
        </div>
    );
}