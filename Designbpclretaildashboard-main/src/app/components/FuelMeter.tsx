import { motion } from 'motion/react';
import { useEffect, useState } from 'react';

interface FuelMeterProps {
  current: number;
  target: number;
  label: string;
}

export function FuelMeter({ current, target, label }: FuelMeterProps) {
  const [displayValue, setDisplayValue] = useState(0);
  const percentage = Math.min((current / target) * 100, 100);

  useEffect(() => {
    let start = 0;
    const duration = 2000;
    const increment = current / (duration / 16);

    const timer = setInterval(() => {
      start += increment;
      if (start >= current) {
        setDisplayValue(current);
        clearInterval(timer);
      } else {
        setDisplayValue(Math.floor(start));
      }
    }, 16);

    return () => clearInterval(timer);
  }, [current]);

  return (
    <div className="flex flex-col items-center py-8">
      <div className="flex items-center justify-center gap-12 mb-6">
        <div className="relative w-[450px] h-[250px]">
          <svg viewBox="0 0 300 160" className="w-full h-full">
            <defs>
              <linearGradient id="meterGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#007BC9" />
                <stop offset="100%" stopColor="#FFE000" />
              </linearGradient>
            </defs>

            <path
              d="M 30 140 A 120 120 0 0 1 270 140"
              fill="none"
              stroke="#E5E7EB"
              strokeWidth="24"
              strokeLinecap="round"
            />

            <motion.path
              d="M 30 140 A 120 120 0 0 1 270 140"
              fill="none"
              stroke="url(#meterGradient)"
              strokeWidth="24"
              strokeLinecap="round"
              strokeDasharray="377"
              initial={{ strokeDashoffset: 377 }}
              animate={{ strokeDashoffset: 377 - (377 * percentage) / 100 }}
              transition={{ duration: 2, ease: "easeOut" }}
            />

            <text x="30" y="158" className="text-xs fill-gray-600" textAnchor="middle">
              0 kL
            </text>
          </svg>
        </div>

        <div className="flex flex-col items-center justify-center bg-white rounded-xl p-6 shadow-sm border-2 border-[#FFE000]">
          <div className="text-sm text-gray-600 mb-2">Volume Target</div>
          <div className="text-3xl font-bold text-[#007BC9]">{target.toLocaleString('en-IN')} kL</div>
        </div>
      </div>

      <div className="text-center">
        <div className="text-sm text-gray-600 mb-3">{label}</div>
        <div className="bg-gradient-to-r from-[#007BC9] to-[#007BC9]/90 rounded-xl px-8 py-4 inline-block shadow-lg">
          <div className="text-4xl font-mono tracking-wider text-white">
            {displayValue.toLocaleString('en-IN')} kL
          </div>
        </div>
        <div className="mt-3 text-sm font-medium" style={{ color: '#007BC9' }}>
          {percentage.toFixed(1)}% of target achieved
        </div>
      </div>
    </div>
  );
}
