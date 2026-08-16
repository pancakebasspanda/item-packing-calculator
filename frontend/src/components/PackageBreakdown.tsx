import { CalculationResult } from '../types/types';

interface PackageBreakdownProps {
    result: CalculationResult | null;
}

export default function PackageBreakdown({ result }: PackageBreakdownProps) {
    if (!result) return null;

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