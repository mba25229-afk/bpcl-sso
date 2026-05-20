import { useState } from 'react';
import { FuelMeter } from './FuelMeter';
import { SalesCard } from './SalesCard';
import { DataEntryDrawer } from './DataEntryDrawer';
import { Fuel, Droplet, Wind, Zap, Coffee, Wrench, CreditCard, Smartphone, Wallet, Car, Plus } from 'lucide-react';
import { Bell, Calendar } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';

const fuelData = [
  { title: 'Petrol', volume: 1850, target: 2000, icon: Fuel, color: '#007BC9', unit: 'kL' as const },
  { title: 'Diesel', volume: 1620, target: 1800, icon: Droplet, color: '#0056A3', unit: 'kL' as const },
  { title: 'CNG', volume: 680, target: 700, icon: Wind, color: '#00A86B', unit: 'kL' as const },
  { title: 'Speed', volume: 260, target: 300, icon: Zap, color: '#FFE000', unit: 'kL' as const },
];

const initialNonFuelData = [
  { title: 'In & Out', revenue: 3850000, target: 4000000, icon: Car, color: '#9333EA', manuallyUpdated: false, unit: 'Cr' as const },
  { title: 'BeCafe', revenue: 3300000, target: 4000000, icon: Coffee, color: '#EA580C', manuallyUpdated: false, unit: 'Cr' as const },
  { title: 'NO2', revenue: 2200000, target: 3000000, icon: Wind, color: '#0EA5E9', manuallyUpdated: false, unit: 'Cr' as const },
  { title: 'MAK Lubricants', revenue: 1650000, target: 2000000, icon: Wrench, color: '#EAB308', manuallyUpdated: false, unit: 'Cr' as const },
];

const paymentData = [
  { title: 'Cash', amount: 15400000, target: 16000000, color: '#10B981', icon: Wallet },
  { title: 'UPI', amount: 24795000, target: 27000000, color: '#6366F1', icon: Smartphone },
  { title: 'Credit Cards', amount: 9918000, target: 10800000, color: '#F59E0B', icon: CreditCard },
  { title: 'UFill', amount: 4959000, target: 6000000, color: '#EC4899', icon: Car },
];

const ufillBreakdown = [
  { type: '4 Wheelers', share: 65, count: 1240, color: '#EC4899' },
  { type: '2 Wheelers', share: 35, count: 680, color: '#F472B6' },
];

export function Dashboard() {
  const totalVolume = 4410; // Total fuel volume in kL
  const targetVolume = 4800; // Target volume in kL
  const [showUfillBreakdown, setShowUfillBreakdown] = useState(false);
  const [showDataEntry, setShowDataEntry] = useState(false);
  const [nonFuelData, setNonFuelData] = useState(initialNonFuelData);

  const handleDataSave = (data: any) => {
    toast.success('Non-fuel data updated successfully!');
    const updatedData = nonFuelData.map(item => ({
      ...item,
      manuallyUpdated: true
    }));
    setNonFuelData(updatedData);
  };

  return (
    <>
      <DataEntryDrawer
        isOpen={showDataEntry}
        onClose={() => setShowDataEntry(false)}
        onSave={handleDataSave}
      />
      <div className="p-8 space-y-8">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Syall Service Station</h1>
          <p className="text-gray-600 mt-1">KIRTI NAGAR DELHI, WEST DELHI, Delhi 110015
</p>
          <div className="mt-3 inline-flex items-center gap-2 bg-[#FFE000] text-[#007BC9] px-5 py-2 rounded-full text-lg font-bold shadow-md">
            <span>Rank #5</span>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button className="p-2 hover:bg-gray-100 rounded-lg transition-colors relative">
            <Bell size={20} className="text-gray-600" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-[#FF6B35] rounded-full"></span>
          </button>
          <button className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            <Calendar size={18} className="text-gray-600" />
            <span className="text-sm text-gray-700">April 2026</span>
          </button>
          <button className="px-6 py-2 bg-[#007BC9] text-white rounded-lg hover:bg-[#0056A3] transition-colors">
            Check Performance
          </button>
        </div>
      </div>

      <div className="bg-gradient-to-br from-blue-50 to-yellow-50 rounded-2xl p-8">
        <FuelMeter current={totalVolume} target={targetVolume} label="Total Volume Sold" />
      </div>

      <div>
        <h2 className="text-xl font-bold text-gray-900 mb-4">Fuel Breakdown (Volume)</h2>
        <div className="grid grid-cols-4 gap-4">
          {fuelData.map((item, index) => (
            <SalesCard key={item.title} {...item} delay={index * 0.1} />
          ))}
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-900">Non-Fuel Breakdown (Revenue)</h2>
          <button
            onClick={() => setShowDataEntry(true)}
            className="flex items-center gap-2 px-4 py-2 bg-[#FFE000] text-[#007BC9] rounded-lg font-bold hover:bg-[#FFD000] transition-colors shadow-sm"
          >
            <Plus size={18} />
            Add / Edit Non-Fuel Data
          </button>
        </div>
        <div className="grid grid-cols-4 gap-4">
          {nonFuelData.map((item, index) => (
            <SalesCard key={item.title} {...item} delay={index * 0.1} />
          ))}
        </div>
      </div>

      <div>
        <h2 className="text-xl font-bold text-gray-900 mb-4">Payment Breakdown</h2>
        <div className="grid grid-cols-4 gap-4">
          {paymentData.map((payment) => {
            const Icon = payment.icon;
            const isUfill = payment.title === 'UFill';
            const achievementPercentage = Math.min((payment.amount / payment.target) * 100, 100);

            let barColor = payment.color;
            if (achievementPercentage < 70) {
              barColor = '#EF4444';
            } else if (achievementPercentage < 90) {
              barColor = '#FFE000';
            } else {
              barColor = '#10B981';
            }

            return (
              <motion.div
                key={payment.title}
                onClick={isUfill ? () => setShowUfillBreakdown(!showUfillBreakdown) : undefined}
                className={`bg-white rounded-xl p-6 shadow-sm border-2 hover:shadow-md transition-all ${
                  isUfill ? 'cursor-pointer hover:border-[#EC4899]' : 'border-gray-100'
                } ${isUfill && showUfillBreakdown ? 'border-[#EC4899]' : ''}`}
                whileHover={isUfill ? { scale: 1.02 } : {}}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="p-2 rounded-lg" style={{ backgroundColor: `${payment.color}20` }}>
                    <Icon size={24} style={{ color: payment.color }} />
                  </div>
                </div>
                <div className="text-sm text-gray-600 mb-1">{payment.title}</div>
                <div className="text-2xl font-bold text-gray-900 mb-1">
                  ₹{(payment.amount / 10000000).toFixed(2)} Cr
                </div>
                <div className="text-xs text-gray-500 mb-3">
                  Target: ₹{(payment.target / 10000000).toFixed(2)} Cr
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
                      transition={{ duration: 1 }}
                      className="h-full rounded-full"
                      style={{
                        backgroundColor: barColor,
                        boxShadow: achievementPercentage >= 100 ? `0 0 10px ${barColor}` : 'none'
                      }}
                    />
                  </div>
                </div>
                {isUfill && (
                  <div className="mt-2 text-xs text-[#EC4899] font-medium">
                    {showUfillBreakdown ? 'Hide breakdown ▲' : 'Click for breakdown ▼'}
                  </div>
                )}
              </motion.div>
            );
          })}
        </div>

        <AnimatePresence>
          {showUfillBreakdown && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="mt-4 bg-white rounded-xl p-6 shadow-sm border border-[#EC4899]"
            >
              <h3 className="text-lg font-bold text-gray-900 mb-4">UFill Vehicle Breakdown</h3>
              <div className="grid grid-cols-2 gap-6">
                {ufillBreakdown.map((item) => (
                  <div key={item.type} className="flex items-center gap-4">
                    <div className="flex-1">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-gray-600">{item.type}</span>
                        <span className="text-lg font-bold" style={{ color: item.color }}>
                          {item.share}%
                        </span>
                      </div>
                      <div className="w-full bg-gray-100 rounded-full h-3 overflow-hidden">
                        <motion.div
                          initial={{ width: 0 }}
                          animate={{ width: `${item.share}%` }}
                          transition={{ duration: 1 }}
                          className="h-full rounded-full"
                          style={{ backgroundColor: item.color }}
                        />
                      </div>
                      <div className="mt-2 text-xs text-gray-500">{item.count} transactions</div>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
    </>
  );
}
