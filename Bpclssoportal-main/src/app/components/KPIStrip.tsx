import { TrendingUp, Target, DollarSign, PieChart } from 'lucide-react';

export function KPIStrip({ data }: { data: any }) {
  const kpis = [
    {
      label: 'Total Revenue',
      value: `₹${data.totalRevenue} Cr`,
      icon: DollarSign,
      highlight: true,
    },
    {
      label: 'Target Achievement',
      value: `${data.targetAchievement}%`,
      icon: Target,
      highlight: false,
    },
    {
      label: 'YoY Growth',
      value: `${data.yoyGrowth}%`,
      icon: TrendingUp,
      highlight: false,
    },
    {
      label: 'Fuel vs Non-Fuel',
      value: `${data.fuelVsNonFuel}`,
      icon: PieChart,
      highlight: false,
    },
  ];

  return (
    <div className="grid grid-cols-4 gap-6 mb-6">
      {kpis.map((kpi, index) => {
        const Icon = kpi.icon;
        return (
          <div
            key={index}
            className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 transition-shadow hover:shadow-md"
            style={kpi.highlight ? { borderLeft: '4px solid #FFE000' } : {}}
          >
            <div className="flex items-start justify-between mb-3">
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: kpi.highlight ? '#FFE000' : '#007BC9' }}
              >
                <Icon size={20} className={kpi.highlight ? 'text-gray-900' : 'text-white'} />
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
