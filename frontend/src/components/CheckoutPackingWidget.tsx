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

        if (isNaN(quantity)) {
            setResult(null);
            setError('');
            return;
        }

        if (quantity <= 0) {
            setResult(null);
            setError('Order quantity must be greater than zero');
            return;
        }

        const MAX_ITEMS = Number.MAX_SAFE_INTEGER;
        if (quantity > MAX_ITEMS) {
            setResult(null);
            setError(`Apologies, we can only process orders up to ${MAX_ITEMS.toLocaleString()} items at a time.`);
            return;
        }

        let wakeupTimer: ReturnType<typeof setTimeout>;

        const fetchPacks = async () => {
            setIsLoading(true);
            setIsWakingUp(false);
            setError('');

            // If the server takes longer than 3 seconds, assume a cold start
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
                setIsLoading(false);
                setIsWakingUp(false);
            }
        };

        const debounceTimer = setTimeout(fetchPacks, 300);

        return () => {
            clearTimeout(debounceTimer);
            clearTimeout(wakeupTimer);
        };
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

            {isLoading && (
                <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-100 flex items-center gap-3 transition-all">
                    <svg className="animate-spin h-5 w-5 text-blue-600 flex-shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <div>
                        <p className="text-sm font-medium text-blue-900">
                            Calculating optimal packaging...
                        </p>
                        {isWakingUp && (
                            <p className="text-xs text-blue-700 mt-0.5 animate-pulse">
                                Waking up server instance on Render. This may take up to 30 seconds...
                            </p>
                        )}
                    </div>
                </div>
            )}

            {error && (
                <div className="mt-4 p-3 bg-red-50 text-red-700 text-sm rounded-lg border border-red-200">
                    {error}
                </div>
            )}

            {!isLoading && !error && <PackageBreakdown result={result} />}
        </div>
    );
}