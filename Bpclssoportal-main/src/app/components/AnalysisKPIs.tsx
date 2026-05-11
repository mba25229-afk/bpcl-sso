import { TrendingUp, TrendingDown } from 'lucide-react';

export function AnalysisKPIs({ data }: { data: any }) {
  const kpis = [
    { label: 'Fuel YoY Growth', value: data.fuelYoY },
    { label: 'Non-Fuel YoY Growth', subtitle: 'Lubricants + QOC', value: data.nonFuelYoY },
    { label: 'UFill Numbers Growth', value: data.ufillGrowth },
  ];

  return (
    <div className="grid grid-cols-3 gap-6 mb-6">
      {kpis.map((kpi, index) => {
        const Icon = kpi.icon;
        const isPositive = (kpi.value ?? 0) >= 0;
        const displayVal = kpi.value != null ? kpi.value : 0;

        return (
          <div key={index} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-start justify-between mb-3">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: isPositive ? '#10b981' : '#ef4444' }}
              >
                {isPositive
                  ? <TrendingUp size={20} className="text-white" />
                  : <TrendingDown size={20} className="text-white" />}
              </div>
              <div className={`px-2 py-1 rounded text-xs ${isPositive ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                {isPositive ? '↑' : '↓'} {Math.abs(displayVal)}%
              </div>
            </div>
            <div className="text-3xl text-gray-900 mb-1">
              {isPositive ? '+' : ''}{displayVal}%
            </div>
            <div className="text-sm text-gray-600">{kpi.label}</div>
            {kpi.subtitle && <div className="text-xs text-gray-400 mt-0.5">{kpi.subtitle}</div>}
          </div>
        );
      })}
    </div>
  );
}
