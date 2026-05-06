import { TrendingUp, AlertCircle, CheckCircle } from 'lucide-react';

export function SmartInsights({ insights }: { insights: string[] }) {
  const icons = [TrendingUp, CheckCircle, AlertCircle];

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
      <h3 className="mb-4 text-gray-900">Smart Insights</h3>
      <div className="space-y-3">
        {insights.map((insight, index) => {
          const Icon = icons[index % icons.length];
          return (
            <div key={index} className="flex items-start gap-3 p-3 rounded-lg bg-gray-50">
              <div
                className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
                style={{ backgroundColor: '#007BC9' }}
              >
                <Icon size={16} className="text-white" />
              </div>
              <p className="text-gray-700 text-sm pt-1">{insight}</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}
