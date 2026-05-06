import { useState, useEffect } from 'react';
import { Minus, Loader2 } from 'lucide-react';
import { api } from '../../api/client';
import { LineChart, Line, ResponsiveContainer } from 'recharts';

interface TerritoryOutlet {
  cc_number: string;
  name: string;
  district: string | null;
  rank: string | null;
  ms_achieved: number;
  hsd_achieved: number;
  speed_achieved: number;
  ms_achievement_pct: number;
  hsd_achievement_pct: number;
  total_achievement_pct: number;
}

interface TerritorySummary {
  period: string;
  territory_code: string;
  summary: {
    total_outlets: number;
    above_target: number;
    below_target: number;
    avg_achievement_pct: number;
  };
  outlets: TerritoryOutlet[];
}

interface TrendRow {
  period: string;
  total_achieved: number | null;
  fuel_achieved: number | null;
  non_fuel_achieved: number | null;
}

interface TrendResponse {
  cc_number: string;
  name: string;
  periods: TrendRow[];
}

interface TerritoryViewProps {
  onSelectDealer: (cc: string) => void;
}

export function TerritoryView({ onSelectDealer }: TerritoryViewProps) {
  const [period, setPeriod] = useState(() => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  });
  const [data, setData] = useState<TerritorySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<'name' | 'achievement' | 'score'>('achievement');
  const [districtFilter, setDistrictFilter] = useState<string>('all');
  const [trendData, setTrendData] = useState<Map<string, TrendResponse>>(new Map());
  const [selectedCCForTrend, setSelectedCCForTrend] = useState<string | null>(null);
  const [trendLoading, setTrendLoading] = useState(false);

  const loadOutletTrend = async (cc: string) => {
    if (trendData.has(cc)) {
      setSelectedCCForTrend(cc);
      return;
    }
    setTrendLoading(true);
    try {
      const tr = await api.getTrend(cc, 6);
      setTrendData(prev => {
        const next = new Map(prev);
        next.set(cc, tr);
        return next;
      });
      setSelectedCCForTrend(cc);
    } catch { /* ignore */ }
    finally { setTrendLoading(false); }
  };

  useEffect(() => {
    loadTerritoryData();
  }, [period]);

  const loadTerritoryData = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.getTerritorySummary(period);
      setData(result);
    } catch (err: any) {
      setError(err.message || 'Failed to load territory data');
    } finally {
      setLoading(false);
    }
  };

  const getAchievementColor = (pct: number) => {
    if (pct >= 90) return 'text-green-600';
    if (pct >= 75) return 'text-orange-500';
    return 'text-red-600';
  };

  const getTrendIcon = () => {
    return <Minus size={16} className="text-gray-400" />;
  };

  const filteredOutlets = data?.outlets.filter(o => 
    districtFilter === 'all' || o.district === districtFilter
  ) || [];

  const sortedOutlets = [...filteredOutlets].sort((a, b) => {
    if (sortBy === 'achievement') return b.total_achievement_pct - a.total_achievement_pct;
    if (sortBy === 'name') return a.name.localeCompare(b.name);
    return 0;
  });

  const districts = [...new Set(data?.outlets.map(o => o.district).filter(Boolean))];

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="animate-spin text-blue-600" size={40} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <p className="text-red-600">{error}</p>
        <button
          onClick={loadTerritoryData}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">
          Territory Overview
        </h2>
        <div className="flex items-center gap-4">
          <input
            type="month"
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
            style={{ '--tw-ring-color': '#007BC9' } as any}
          />
          <select
            value={districtFilter}
            onChange={(e) => setDistrictFilter(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
            style={{ '--tw-ring-color': '#007BC9' } as any}
          >
            <option value="all">All Districts</option>
            {districts.map(d => (
              <option key={d} value={d!}>{d}</option>
            ))}
          </select>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
            style={{ '--tw-ring-color': '#007BC9' } as any}
          >
            <option value="achievement">Sort by Achievement</option>
            <option value="name">Sort by Name</option>
          </select>
        </div>
      </div>

      {data && (
        <>
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <p className="text-gray-500 text-sm">Total Outlets</p>
              <p className="text-3xl font-bold text-gray-900">{data.summary.total_outlets}</p>
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <p className="text-gray-500 text-sm">Above Target</p>
              <p className="text-3xl font-bold text-green-600">{data.summary.above_target}</p>
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <p className="text-gray-500 text-sm">Below Target</p>
              <p className="text-3xl font-bold text-red-600">{data.summary.below_target}</p>
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <p className="text-gray-500 text-sm">Avg Achievement</p>
              <p className="text-3xl font-bold" style={{ color: '#007BC9' }}>
                {data.summary.avg_achievement_pct.toFixed(1)}%
              </p>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-b border-gray-200">
                    <th className="px-6 py-3 text-left text-sm text-gray-700">Rank</th>
                    <th className="px-6 py-3 text-left text-sm text-gray-700">Dealer Name</th>
                    <th className="px-6 py-3 text-left text-sm text-gray-700">CC</th>
                    <th className="px-6 py-3 text-left text-sm text-gray-700">District</th>
                    <th className="px-6 py-3 text-right text-sm text-gray-700">MS Ach %</th>
                    <th className="px-6 py-3 text-right text-sm text-gray-700">HSD Ach %</th>
                    <th className="px-6 py-3 text-right text-sm text-gray-700">Total</th>
                    <th className="px-6 py-3 text-center text-sm text-gray-700">Trend</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedOutlets.map((outlet, idx) => (
                    <tr
                      key={outlet.cc_number}
                      onClick={() => onSelectDealer(outlet.cc_number)}
                      className="border-b border-gray-100 hover:bg-gray-50 transition-colors cursor-pointer"
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          {idx === 0 && <span>🥇</span>}
                          {idx === 1 && <span>🥈</span>}
                          {idx === 2 && <span>🥉</span>}
                          {idx > 2 && <span className="text-gray-500">{idx + 1}</span>}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-gray-900 font-medium">{outlet.name}</td>
                      <td className="px-6 py-4 text-gray-500">{outlet.cc_number}</td>
                      <td className="px-6 py-4 text-gray-500">{outlet.district || '-'}</td>
                      <td className={`px-6 py-4 text-right font-medium ${getAchievementColor(outlet.ms_achievement_pct)}`}>
                        {outlet.ms_achievement_pct.toFixed(1)}%
                      </td>
                      <td className={`px-6 py-4 text-right font-medium ${getAchievementColor(outlet.hsd_achievement_pct)}`}>
                        {outlet.hsd_achievement_pct.toFixed(1)}%
                      </td>
                      <td className={`px-6 py-4 text-right font-bold ${getAchievementColor(outlet.total_achievement_pct)}`}>
                        {outlet.total_achievement_pct.toFixed(1)}%
                      </td>
                      <td className="px-6 py-4 text-center">
                          <button
                            onClick={(e) => { e.stopPropagation(); loadOutletTrend(outlet.cc_number); }}
                            className="text-blue-600 hover:underline text-sm"
                          >
                            View
                          </button>
                        </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {selectedCCForTrend && (
            <div className="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">
                  Trend — {trendData.get(selectedCCForTrend)?.name || selectedCCForTrend}
                </h3>
                <button
                  onClick={() => setSelectedCCForTrend(null)}
                  className="text-gray-500 hover:text-gray-700"
                >
                  Close
                </button>
              </div>
              {trendLoading ? (
                <div className="h-48 flex items-center justify-center">
                  <Loader2 className="animate-spin text-blue-600" size={32} />
                </div>
              ) : (
                <ResponsiveContainer width="100%" height={200}>
                  <LineChart data={[... (trendData.get(selectedCCForTrend)?.periods || [])].reverse()}>
                    <Line
                      type="monotone"
                      dataKey="total_achieved"
                      stroke="#007BC9"
                      strokeWidth={2}
                      dot={false}
                      name="Total"
                    />
                  </LineChart>
                </ResponsiveContainer>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}