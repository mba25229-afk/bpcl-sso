import { TrendingUp, TrendingDown, Activity, Percent } from 'lucide-react';

export function AnalysisKPIs({ data }: { data: any }) {
  const kpis = [
    {
      label: 'Fuel YoY Growth',
      value: `+${data.fuelYoY}%`,
      change: data.fuelYoY,
      icon: TrendingUp,
    },
    {
      label: 'Non-Fuel YoY Growth',
      value: `+${data.nonFuelYoY}%`,
      change: data.nonFuelYoY,
      icon: TrendingUp,
    },
    {
      label: 'Volume Growth',
      value: `+${data.volumeGrowth}%`,
      change: data.volumeGrowth,
      icon: Activity,
    },
    {
      label: 'Revenue Growth Rate',
      value: `+${data.revenueGrowth}%`,
      change: data.revenueGrowth,
      icon: Percent,
    },
  ];

  return (
    <div className="grid grid-cols-4 gap-6 mb-6">
      {kpis.map((kpi, index) => {
        const Icon = kpi.icon;
        const isPositive = kpi.change >= 0;

        return (
          <div
            key={index}
            className="bg-white rounded-xl shadow-sm border border-gray-200 p-6"
          >
            <div className="flex items-start justify-between mb-3">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: isPositive ? '#10b981' : '#ef4444' }}
              >
                <Icon size={20} className="text-white" />
              </div>
              <div className={`px-2 py-1 rounded text-xs ${isPositive ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                {isPositive ? '↑' : '↓'} {Math.abs(kpi.change)}%
              </div>
            </div>
            <div className="text-3xl text-gray-900 mb-1">{kpi.value}</div>
            <div className="text-sm text-gray-600">{kpi.label}</div>
          </div>
        );
      })}
    </div>
  );
}
