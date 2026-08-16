import { CalculationResult } from '../types/types';

interface PackageBreakdownProps {
    result: CalculationResult | null;
    isLoading?: boolean;
    isWakingUp?: boolean;
}

export default function PackageBreakdown({ result, isLoading, isWakingUp }: PackageBreakdownProps) {
    // show the loading spinner and cold start message if waiting for the API
    if (isLoading) {
        return (
            <div className="mt-6 p-6 bg-[#F2F5F3] rounded-xl border border-[#D4DDD6] border-l-4 border-l-[#5F7A65] flex flex-col items-center justify-center space-y-3 transition-all min-h-[150px]">
                {/* Tailwind SVG Spinner */}
                <svg className="animate-spin h-8 w-8 text-[#5F7A65]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-80" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>

                <div className="text-center">
                    <p className="text-sm font-medium text-gray-800">
                        Calculating optimal packaging...
                    </p>
                    {/* Render Cold Start Warning */}
                    {isWakingUp && (
                        <p className="text-xs text-gray-600 mt-1 animate-pulse">
                            Waking up backend server on Render. This may take up to 50 seconds...
                        </p>
                    )}
                </div>
            </div>
        );
    }

    // return nothing if not loading and there are no results
    if (!result) return null;

    // render the breakdown once results arrive
    // create a copy of the packs array and sort them from largest to smallest pack size
    const sortedPacks = [...result.packs].sort((a, b) => b.packSize - a.packSize);

    return (
        <div className="mt-6 p-4 bg-[#F2F5F3] rounded-xl border border-[#D4DDD6] border-l-4 border-l-[#5F7A65] transition-all">
            <h3 className="text-lg font-semibold text-gray-800 mb-2">Shipping Breakdown</h3>

            <p className="text-sm text-gray-600 mb-4">
                Your {result.orderQuantity} items will be shipped in {result.totalPacks} boxes.
            </p>

            {/* 1 column on mobile, 2 columns on small screens, 3 columns on large screens */}
            <ul className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
                {sortedPacks.map((pack) => {
                    // Skip if there's no quantity
                    if (pack.quantity <= 0) return null;

                    return (
                        <li key={pack.packSize} className="flex items-center justify-between p-4 bg-white rounded-lg shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
                            <div className="flex items-center space-x-3">
                                <span className="text-2xl" role="img" aria-label="box">
                                    📦
                                </span>
                                <span className="font-medium text-gray-700">{pack.packSize}-pack</span>
                            </div>
                            <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full font-bold">
                                x {pack.quantity}
                            </span>
                        </li>
                    );
                })}
            </ul>
        </div>
    );
}