import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { TrendingUp, TrendingDown } from 'lucide-react';
import { motion } from 'motion/react';

const volumeShareData = [
  { name: 'Petrol', value: 1850, color: '#007BC9' },
  { name: 'Diesel', value: 1620, color: '#0056A3' },
  { name: 'CNG', value: 680, color: '#00A86B' },
  { name: 'Speed', value: 260, color: '#FFE000' },
];

const paymentDistributionData = [
  { name: 'UPI', value: 45, color: '#6366F1' },
  { name: 'Cash', value: 28, color: '#10B981' },
  { name: 'Credit Cards', value: 18, color: '#F59E0B' },
  { name: 'UFill', value: 9, color: '#EC4899' },
];

const vehicleTypeData = [
  { name: '4 Wheeler', value: 42, color: '#007BC9' },
  { name: '2 Wheeler', value: 38, color: '#FFE000' },
  { name: 'Commercial', value: 20, color: '#00A86B' },
];

const productShareData = [
  { name: 'Petrol', value: 33.5, color: '#003D7A' },
  { name: 'Diesel', value: 29.4, color: '#0066CC' },
  { name: 'In & Out', value: 7.0, color: '#9333EA' },
  { name: 'CNG', value: 12.3, color: '#00A86B' },
  { name: 'BeCafe', value: 6.0, color: '#EA580C' },
  { name: 'Speed', value: 4.7, color: '#FF6B35' },
  { name: 'NO2', value: 4.0, color: '#0EA5E9' },
  { name: 'MAK Lubricants', value: 3.0, color: '#EAB308' },
];

const kpiData = [
  { label: 'Volume Growth', value: '+12.5%', trend: 'up', color: '#10B981' },
  { label: 'Avg Volume/Day', value: '147 kL', trend: 'up', color: '#10B981' },
  { label: 'Daily Volume Sold', value: '147 kL', trend: 'up', color: '#10B981' },
  { label: 'Peak Hours', value: '8-10 AM', trend: 'neutral', color: '#6B7280' },
  { label: 'Target Achievement', value: '91.9%', trend: 'up', color: '#10B981' },
  { label: 'Non-Fuel Revenue', value: '₹1.10 Cr', trend: 'up', color: '#10B981' },
];

export function Analysis() {
  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Analysis Dashboard</h1>
        <p className="text-gray-600">Comprehensive performance metrics and insights</p>
      </div>

      <div className="grid grid-cols-3 gap-6">
        {kpiData.map((kpi, index) => (
          <motion.div
            key={kpi.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
            className="bg-white rounded-xl p-6 shadow-sm border border-gray-100"
          >
            <div className="flex items-start justify-between">
              <div>
                <div className="text-sm text-gray-600 mb-2">{kpi.label}</div>
                <div className="text-3xl font-bold text-gray-900">{kpi.value}</div>
              </div>
              {kpi.trend === 'up' && <TrendingUp size={24} style={{ color: kpi.color }} />}
              {kpi.trend === 'down' && <TrendingDown size={24} style={{ color: kpi.color }} />}
            </div>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <h3 className="text-lg font-bold text-gray-900 mb-4">Fuel Volume Share</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={volumeShareData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {volumeShareData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip formatter={(value: number) => `${value.toLocaleString('en-IN')} kL`} />
            </PieChart>
          </ResponsiveContainer>
          <div className="mt-4 grid grid-cols-2 gap-2">
            {volumeShareData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-600">{item.name}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <h3 className="text-lg font-bold text-gray-900 mb-4">Payment Distribution</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={paymentDistributionData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name} ${value}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {paymentDistributionData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip formatter={(value: number) => `${value}%`} />
            </PieChart>
          </ResponsiveContainer>
          <div className="mt-4 grid grid-cols-2 gap-3">
            {paymentDistributionData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-600">{item.name}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <h3 className="text-lg font-bold text-gray-900 mb-4">Vehicle Type Share</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={vehicleTypeData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name} ${value}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {vehicleTypeData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip formatter={(value: number) => `${value}%`} />
            </PieChart>
          </ResponsiveContainer>
          <div className="mt-4 flex flex-col gap-2">
            {vehicleTypeData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-600">{item.name}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h3 className="text-lg font-bold text-gray-900 mb-4">Product Category Share (Pyramid)</h3>
        <div className="flex flex-col items-center gap-2">
          {productShareData.map((product, index) => {
            const width = product.value * 10;
            return (
              <motion.div
                key={product.name}
                initial={{ width: 0, opacity: 0 }}
                animate={{ width: `${width}%`, opacity: 1 }}
                transition={{ delay: index * 0.1, duration: 0.5 }}
                className="flex items-center justify-between px-4 py-3 rounded-lg text-white"
                style={{ backgroundColor: product.color, maxWidth: '600px' }}
              >
                <span className="font-medium">{product.name}</span>
                <span className="font-bold">{product.value}%</span>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
