import { PieChart, Pie, Cell, BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Area, AreaChart, ComposedChart } from 'recharts';

const COLORS = {
  primary: '#007BC9',
  accent: '#FFE000',
  success: '#10b981',
  warning: '#f59e0b',
  purple: '#8b5cf6',
  pink: '#ec4899',
};

export function AnalysisCharts({ fuelData, nonFuelData, targetData, trendData, growthData, categoryGrowthData }: any) {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="mb-4 text-gray-900">Fuel Share Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={fuelData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {fuelData.map((entry: any, index: number) => (
                  <Cell key={`cell-${index}`} fill={Object.values(COLORS)[index % Object.values(COLORS).length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="mb-4 text-gray-900">Non-Fuel Breakdown</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={nonFuelData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {nonFuelData.map((entry: any, index: number) => (
                  <Cell key={`cell-${index}`} fill={Object.values(COLORS)[index % Object.values(COLORS).length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="mb-4 text-gray-900">Target vs Achieved</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={targetData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="category" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="target" fill="#94a3b8" name="Target" />
            <Bar dataKey="achieved" fill={COLORS.primary} name="Achieved" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="mb-4 text-gray-900">YoY Revenue Growth Trend</h3>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={growthData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Area type="monotone" dataKey="currentYear" stroke={COLORS.primary} fill={COLORS.primary} fillOpacity={0.3} name="Current Year (Cr)" />
              <Area type="monotone" dataKey="lastYear" stroke="#94a3b8" fill="#94a3b8" fillOpacity={0.3} name="Last Year (Cr)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="mb-4 text-gray-900">Category-wise YoY Growth</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={categoryGrowthData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="category" />
              <YAxis label={{ value: 'Growth %', angle: -90, position: 'insideLeft' }} />
              <Tooltip />
              <Legend />
              <Bar dataKey="growth" fill={COLORS.success} name="YoY Growth %" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="mb-4 text-gray-900">Sales Trend</h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={trendData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="fuel" stroke={COLORS.primary} strokeWidth={2} name="Fuel" />
            <Line type="monotone" dataKey="nonFuel" stroke={COLORS.accent} strokeWidth={2} name="Non-Fuel" />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
