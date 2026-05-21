import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp, AlertCircle, Zap } from 'lucide-react';
import { motion } from 'motion/react';

const volumeTrendData = [
  { date: 'Apr 1', fuel: 145, nonFuel: 0.35 },
  { date: 'Apr 5', fuel: 152, nonFuel: 0.38 },
  { date: 'Apr 10', fuel: 158, nonFuel: 0.36 },
  { date: 'Apr 15', fuel: 165, nonFuel: 0.42 },
  { date: 'Apr 20', fuel: 172, nonFuel: 0.45 },
  { date: 'Apr 25', fuel: 180, nonFuel: 0.48 },
  { date: 'Apr 29', fuel: 185, nonFuel: 0.52 },
];

const productPerformanceData = [
  { product: 'Petrol', current: 1850, previous: 1720 },
  { product: 'Diesel', current: 1620, previous: 1580 },
  { product: 'CNG', current: 680, previous: 620 },
  { product: 'In & Out', current: 38.5, previous: 34 },
  { product: 'BeCafe', current: 33, previous: 31 },
];

const paymentTrendData = [
  { month: 'Jan', cash: 35, upi: 38, cards: 20, ufill: 7 },
  { month: 'Feb', cash: 32, upi: 40, cards: 20, ufill: 8 },
  { month: 'Mar', cash: 30, upi: 42, cards: 19, ufill: 9 },
  { month: 'Apr', cash: 28, upi: 45, cards: 18, ufill: 9 },
];

const insights = [
  {
    icon: TrendingUp,
    title: 'Weekend Peak Demand',
    description: 'Expected 25% increase in sales this weekend',
    color: '#10B981',
  },
  {
    icon: Zap,
    title: 'UPI Growth',
    description: 'Digital payments up 18% month-over-month',
    color: '#6366F1',
  },
  {
    icon: AlertCircle,
    title: 'Diesel Trend',
    description: 'Slight decline observed, monitor commercial fleet activity',
    color: '#F59E0B',
  },
];

export function Trends() {
  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Trends & Forecasting</h1>
        <p className="text-gray-600">Historical patterns and predictive insights</p>
      </div>

      <div className="grid grid-cols-3 gap-4">
        {insights.map((insight, index) => {
          const Icon = insight.icon;
          return (
            <motion.div
              key={insight.title}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="bg-white rounded-xl p-4 shadow-sm border border-gray-100"
            >
              <div className="flex items-start gap-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: `${insight.color}20` }}>
                  <Icon size={20} style={{ color: insight.color }} />
                </div>
                <div>
                  <h4 className="font-bold text-gray-900 mb-1">{insight.title}</h4>
                  <p className="text-sm text-gray-600">{insight.description}</p>
                </div>
              </div>
            </motion.div>
          );
        })}
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h3 className="text-lg font-bold text-gray-900 mb-4">Volume Trends (April 2026)</h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={volumeTrendData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
            <XAxis dataKey="date" stroke="#6B7280" />
            <YAxis stroke="#6B7280" />
            <Tooltip formatter={(value: number) => `${value.toLocaleString('en-IN')} kL`} />
            <Legend />
            <Line
              type="monotone"
              dataKey="fuel"
              stroke="#007BC9"
              strokeWidth={3}
              name="Fuel Volume (kL)"
              dot={{ fill: '#007BC9', r: 4 }}
            />
            <Line
              type="monotone"
              dataKey="nonFuel"
              stroke="#FFE000"
              strokeWidth={3}
              name="Non-Fuel Revenue (Cr)"
              dot={{ fill: '#FFE000', r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <h3 className="text-lg font-bold text-gray-900 mb-4">Volume Performance (Month-over-Month)</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={productPerformanceData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
              <XAxis dataKey="product" stroke="#6B7280" />
              <YAxis stroke="#6B7280" />
              <Tooltip formatter={(value: number) => `${value.toLocaleString('en-IN')} kL`} />
              <Legend />
              <Bar dataKey="previous" fill="#9CA3AF" name="Previous Month" />
              <Bar dataKey="current" fill="#007BC9" name="Current Month" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <h3 className="text-lg font-bold text-gray-900 mb-4">Payment Mode Trends</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={paymentTrendData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
              <XAxis dataKey="month" stroke="#6B7280" />
              <YAxis stroke="#6B7280" />
              <Tooltip formatter={(value: number) => `${value}%`} />
              <Legend />
              <Line type="monotone" dataKey="cash" stroke="#10B981" strokeWidth={2} name="Cash" />
              <Line type="monotone" dataKey="upi" stroke="#6366F1" strokeWidth={2} name="UPI" />
              <Line type="monotone" dataKey="cards" stroke="#F59E0B" strokeWidth={2} name="Cards" />
              <Line type="monotone" dataKey="ufill" stroke="#EC4899" strokeWidth={2} name="UFill" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-gradient-to-br from-blue-50 to-yellow-50 rounded-xl p-6 border border-gray-200">
        <h3 className="text-lg font-bold text-gray-900 mb-4">Forecast: Next Week</h3>
        <div className="grid grid-cols-3 gap-6">
          <div className="bg-white rounded-lg p-4">
            <div className="text-sm text-gray-600 mb-1">Expected Volume</div>
            <div className="text-2xl font-bold text-gray-900">1,175 kL</div>
            <div className="text-sm text-green-600 mt-1">+8.5% vs last week</div>
          </div>
          <div className="bg-white rounded-lg p-4">
            <div className="text-sm text-gray-600 mb-1">Peak Day</div>
            <div className="text-2xl font-bold text-gray-900">Saturday</div>
            <div className="text-sm text-gray-600 mt-1">Expected: 195 kL</div>
          </div>
          <div className="bg-white rounded-lg p-4">
            <div className="text-sm text-gray-600 mb-1">Top Product</div>
            <div className="text-2xl font-bold text-gray-900">Petrol</div>
            <div className="text-sm text-gray-600 mt-1">Demand surge expected</div>
          </div>
        </div>
      </div>
    </div>
  );
}
