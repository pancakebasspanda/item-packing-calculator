import { CalculationResult } from '../types/types';

interface PackageBreakdownProps {
    result: CalculationResult | null;
}

export default function PackageBreakdown({ result }: PackageBreakdownProps) {
    if (!result) return null;

    return (
        <div className="mt-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Pack Breakdown</h3>
            {/* Add your pack mapping logic here */}
        </div>
    );
}