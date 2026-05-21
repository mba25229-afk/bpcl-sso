import { motion } from 'motion/react';
import { LucideIcon, Edit3 } from 'lucide-react';

interface SalesCardProps {
  title: string;
  volume?: number;
  revenue?: number;
  target: number;
  icon?: LucideIcon;
  color: string;
  delay?: number;
  manuallyUpdated?: boolean;
  unit?: 'kL' | 'Cr';
}

export function SalesCard({
  title,
  volume,
  revenue,
  target,
  icon: Icon,
  color,
  delay = 0,
  manuallyUpdated = false,
  unit = 'Cr'
}: SalesCardProps) {
  const displayValue = unit === 'kL' ? volume : revenue;
  const formattedValue = unit === 'kL'
    ? displayValue?.toLocaleString('en-IN')
    : (displayValue! / 10000000).toFixed(2);
  const formattedTarget = unit === 'kL'
    ? target.toLocaleString('en-IN')
    : (target / 10000000).toFixed(2);

  const achievementPercentage = Math.min((displayValue! / target) * 100, 100);

  let barColor = color;
  if (achievementPercentage < 70) {
    barColor = '#EF4444';
  } else if (achievementPercentage < 90) {
    barColor = '#FFE000';
  } else {
    barColor = '#10B981';
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      className="bg-white rounded-xl p-4 shadow-sm border border-gray-100 hover:shadow-md transition-shadow relative"
    >
      {manuallyUpdated && (
        <div className="absolute top-2 right-2 group">
          <div className="bg-[#FFE000] rounded-full p-1.5">
            <Edit3 size={12} className="text-[#007BC9]" />
          </div>
          <div className="absolute right-0 top-8 bg-gray-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
            Manually Updated
          </div>
        </div>
      )}

      <div className="flex items-start justify-between mb-3">
        <div>
          <div className="text-sm text-gray-600 mb-1">{title}</div>
          <div className="text-2xl font-bold text-gray-900">
            {unit === 'kL' ? `${formattedValue} kL` : `₹${formattedValue} Cr`}
          </div>
          <div className="text-xs text-gray-500 mt-0.5">
            Target: {unit === 'kL' ? `${formattedTarget} kL` : `₹${formattedTarget} Cr`}
          </div>
        </div>
        {Icon && (
          <div className={`p-2 rounded-lg`} style={{ backgroundColor: `${color}20` }}>
            <Icon size={20} style={{ color }} />
          </div>
        )}
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="text-gray-600">Target Achieved</span>
          <span className="font-bold" style={{ color: barColor }}>
            {achievementPercentage.toFixed(1)}%
          </span>
        </div>
        <div className="w-full bg-gray-100 rounded-full h-3 overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: `${Math.min(achievementPercentage, 100)}%` }}
            transition={{ duration: 1, delay: delay + 0.3 }}
            className="h-full rounded-full relative"
            style={{
              backgroundColor: barColor,
              boxShadow: achievementPercentage >= 100 ? `0 0 10px ${barColor}` : 'none'
            }}
          />
        </div>
      </div>
    </motion.div>
  );
}
