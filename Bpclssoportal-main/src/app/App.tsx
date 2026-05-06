import { useState, useCallback } from 'react';
import { Navigation } from './components/Navigation';
import { InputSection } from './components/InputSection';
import { ROHeader } from './components/ROHeader';
import { KPIStrip } from './components/KPIStrip';
import { PerformanceTable } from './components/PerformanceTable';
import { AnalysisCharts } from './components/AnalysisCharts';
import { AnalysisKPIs } from './components/AnalysisKPIs';
import { ManagerProfile } from './components/ManagerProfile';
import { SetTargetsModal } from './components/SetTargetsModal';
import { DealerScorecard } from './components/DealerScorecard';
import { TerritoryView } from './components/TerritoryView';
import { AdminView } from './components/AdminView';
import { api } from '../api/client';

// ── Types ──────────────────────────────────────────────────────────────────────

interface PerformanceRow {
  product: string;
  target: number;
  achieved: number;
  ly: number;
  volume: number;
}

interface DashboardData {
  outlet: any;
  performance: any;
}

interface AnalysisData {
  analysis: any;
  performance: any;
}

interface LeaderboardData {
  competition: any;
}

interface MarketShareStatus {
  competition_id: string;
  total_dealers: number;
  dealers_with_ms_data: number;
  dealers_missing: number;
  message: string;
}

const COMPETITION_ID = 'c0000000-0001-0001-0001-000000000001';

// ── Helpers ────────────────────────────────────────────────────────────────────

function toPeriod(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  return `${y}-${m}`;
}

function periodMinus(period: string, months: number): string {
  const [y, m] = period.split('-').map(Number);
  const d = new Date(y, m - 1 - months, 1);
  return toPeriod(d);
}

function buildFuelRows(fuel: any[]): PerformanceRow[] {
  const rows = fuel.map((f) => ({
    product: f.product_code,
    target: f.target ?? 0,
    achieved: f.achieved ?? 0,
    ly: f.last_year ?? 0,
    volume: f.volume_kl ?? 0,
  }));
  const totals = rows.reduce(
    (acc, r) => ({
      product: 'TOTAL',
      target: acc.target + r.target,
      achieved: acc.achieved + r.achieved,
      ly: acc.ly + r.ly,
      volume: acc.volume + r.volume,
      isTotal: true,
    }),
    { product: 'TOTAL', target: 0, achieved: 0, ly: 0, volume: 0, isTotal: true } as any,
  );
  return [...rows, totals];
}

function buildNonFuelRows(nonFuel: any[]): PerformanceRow[] {
  const rows = nonFuel.map((f) => ({
    product: f.product_code,
    target: f.target ?? 0,
    achieved: f.achieved ?? 0,
    ly: f.last_year ?? 0,
    volume: 0,
  }));
  const totals = rows.reduce(
    (acc, r) => ({
      product: 'TOTAL',
      target: acc.target + r.target,
      achieved: acc.achieved + r.achieved,
      ly: acc.ly + r.ly,
      volume: 0,
      isTotal: true,
    }),
    { product: 'TOTAL', target: 0, achieved: 0, ly: 0, volume: 0, isTotal: true } as any,
  );
  return [...rows, totals];
}

function buildKPIData(perf: any) {
  const fuel = perf.fuel ?? [];
  const nf = perf.non_fuel ?? [];
  const fuelAvgAch =
    fuel.length > 0
      ? fuel.reduce((s: number, f: any) => s + (f.target_hit_pct ?? 0), 0) / fuel.length
      : 0;
  const nfAvgAch =
    nf.length > 0
      ? nf.reduce((s: number, f: any) => s + (f.target_hit_pct ?? 0), 0) / nf.length
      : 0;
  const fuelPct = Math.round(fuelAvgAch);
  const nfPct = Math.round(nfAvgAch);

  return {
    totalRevenue: ((perf.total_revenue_cr ?? 0) * 100).toFixed(1),
    targetAchievement: (perf.target_achievement_pct ?? 0).toFixed(1),
    yoyGrowth: `${(perf.yoy_growth_pct ?? 0) >= 0 ? '+' : ''}${(perf.yoy_growth_pct ?? 0).toFixed(1)}`,
    fuelVsNonFuel: `${fuelPct}:${nfPct}`,
  };
}

function buildROData(outlet: any, cc: string) {
  return {
    name: outlet.name ?? cc,
    location: outlet.location ?? '',
    ccNumber: outlet.cc_number ?? cc,
    rank: outlet.rank ?? '–',
  };
}

function buildAnalysisKPIs(analysis: any, perf: any) {
  const fuelGrowth = analysis.category_growth?.find((c: any) => c.category === 'fuel')?.growth_pct ?? 0;
  const nfGrowth = analysis.category_growth?.find((c: any) => c.category === 'non_fuel')?.growth_pct ?? 0;
  return {
    fuelYoY: parseFloat(fuelGrowth.toFixed(1)),
    nonFuelYoY: parseFloat(nfGrowth.toFixed(1)),
    volumeGrowth: parseFloat((perf?.yoy_growth_pct ?? 0).toFixed(1)),
    revenueGrowth: parseFloat((perf?.yoy_growth_pct ?? 0).toFixed(1)),
  };
}

function buildChartData(analysis: any) {
  const fuelData = (analysis.fuel_mix ?? []).map((p: any) => ({
    name: p.product_code,
    value: parseFloat(p.value.toFixed(2)),
  }));
  const nonFuelData = (analysis.non_fuel_mix ?? []).map((p: any) => ({
    name: p.product_code,
    value: parseFloat(p.value.toFixed(2)),
  }));
  const targetData = (analysis.target_vs_achieved ?? []).map((t: any) => ({
    category: t.period,
    target: parseFloat(t.target.toFixed(0)),
    achieved: parseFloat(t.achieved.toFixed(0)),
  }));
  const trendData = (analysis.weekly_trend ?? []).map((w: any) => ({
    date: w.week,
    fuel: parseFloat((w.volume * 0.7).toFixed(0)),
    nonFuel: parseFloat((w.volume * 0.3).toFixed(0)),
  }));
  const growthData = (analysis.monthly_growth ?? []).map((m: any) => ({
    month: m.period,
    currentYear: parseFloat((m.value / 10_000_000).toFixed(4)),
    lastYear: parseFloat(((m.value / 10_000_000) * 0.9).toFixed(4)),
  }));
  const categoryGrowthData = (analysis.category_growth ?? []).map((c: any) => ({
    category: c.category === 'non_fuel' ? 'Non-Fuel' : 'Fuel',
    growth: parseFloat(c.growth_pct.toFixed(2)),
  }));
  return { fuelData, nonFuelData, targetData, trendData, growthData, categoryGrowthData };
}

// ── Leaderboard component ──────────────────────────────────────────────────────

interface MarketShareStatus {
  competition_id: string;
  total_dealers: number;
  dealers_with_ms_data: number;
  dealers_missing: number;
  message: string;
}

function LeaderboardView({
  competition,
  marketShareStatus,
  onViewScorecard,
}: {
  competition: any;
  marketShareStatus?: MarketShareStatus | null;
  onViewScorecard: (cc: string) => void;
}) {
  const entries = competition.entries ?? [];
  const missing = marketShareStatus?.dealers_missing ?? 0;
  const total = marketShareStatus?.total_dealers ?? 39;

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200" style={{ backgroundColor: '#007BC9' }}>
        <div className="flex items-center justify-between">
          <h3 className="text-white">{competition.name}</h3>
          <span className="text-blue-100 text-sm">Period: {competition.period}</span>
        </div>
      </div>

      {/* Market Share Status Banner */}
      {missing > 0 ? (
        <div className="px-6 py-3" style={{ backgroundColor: '#FFE000' }}>
          <p className="text-gray-900 text-sm font-medium">
            ⚠️ Market share data incomplete — {missing}/{total} dealers missing 20-mark MS/HSD gain scores. Upload Delhi Master to unlock full scoring.
          </p>
        </div>
      ) : (
        <div className="px-6 py-3" style={{ backgroundColor: '#007BC9' }}>
          <p className="text-white text-sm font-medium">
            ✅ All dealers fully scored across all 100 marks.
          </p>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm text-gray-700">Rank</th>
              <th className="px-6 py-3 text-left text-sm text-gray-700">Outlet Name</th>
              <th className="px-6 py-3 text-left text-sm text-gray-700">CC Number</th>
              <th className="px-6 py-3 text-right text-sm text-gray-700">Score</th>
              <th className="px-6 py-3 text-left text-sm text-gray-700"></th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e: any) => (
              <tr
                key={e.cc_number}
                className={`border-b border-gray-100 hover:bg-gray-50 transition-colors ${e.rank <= 3 ? 'font-semibold' : ''}`}
              >
                <td className="px-6 py-4">
                  <div
                    className="w-8 h-8 rounded-full flex items-center justify-center text-sm"
                    style={{
                      backgroundColor:
                        e.rank === 1
                          ? '#FFE000'
                          : e.rank === 2
                            ? '#e5e7eb'
                            : e.rank === 3
                              ? '#fcd9a0'
                              : 'transparent',
                      color: e.rank <= 3 ? '#1f2937' : '#6b7280',
                    }}
                  >
                    {e.rank === 1 ? '🥇' : e.rank === 2 ? '🥈' : e.rank === 3 ? '🥉' : e.rank}
                  </div>
                </td>
                <td className="px-6 py-4 text-gray-900">{e.outlet_name}</td>
                <td className="px-6 py-4 text-gray-500 text-sm">{e.cc_number}</td>
                <td className="px-6 py-4 text-right">
                  <span
                    className="px-3 py-1 rounded-full text-sm"
                    style={{
                      backgroundColor: e.is_highlighted ? '#eff6ff' : '#f9fafb',
                      color: e.is_highlighted ? '#1d4ed8' : '#374151',
                    }}
                  >
                    {e.total_score.toFixed(2)}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <button
                    onClick={() => onViewScorecard(e.cc_number)}
                    className="text-sm px-3 py-1 rounded border border-gray-300 hover:bg-gray-50 transition-colors"
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
  );
}

// ── Error / loading states ─────────────────────────────────────────────────────

function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="mt-8 bg-white rounded-xl border border-red-200 p-8 text-center">
      <div className="w-12 h-12 rounded-full bg-red-50 mx-auto mb-3 flex items-center justify-center">
        <span className="text-red-500 text-xl">!</span>
      </div>
      <p className="text-red-700 mb-3">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 rounded-lg border border-gray-300 text-sm text-gray-700 hover:bg-gray-50"
        >
          Retry
        </button>
      )}
    </div>
  );
}

// ── App ────────────────────────────────────────────────────────────────────────

export default function App({ onLogout }: { onLogout: () => void }) {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [showProfile, setShowProfile] = useState(false);
  const [showTargetsModal, setShowTargetsModal] = useState(false);

  // User role for access control
  const user = (() => {
    try {
      return JSON.parse(localStorage.getItem('bpcl_user') || 'null');
    } catch {
      return null;
    }
  })();
  const canSetTargets = user?.role === 'territory_manager' || user?.role === 'admin';

  // Fetch state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [analysisData, setAnalysisData] = useState<AnalysisData | null>(null);
  const [leaderboardData, setLeaderboardData] = useState<LeaderboardData | null>(null);
  const [marketShareStatus, setMarketShareStatus] = useState<MarketShareStatus | null>(null);
  const [currentCC, setCurrentCC] = useState('');
  const [currentPeriod, setCurrentPeriod] = useState('');

  // Targets (pulled from API on fetch, editable locally)
  const [targets, setTargets] = useState<any>(null);

  // Scorecard modal state
  const [selectedDealer, setSelectedDealer] = useState<any>(null);
  const [scorecardData, setScorecardData] = useState<any>(null);

  const handleFetch = useCallback(async (cc: string, selectedDate: Date) => {
    const cleanCC = cc.replace(/^CC\s*/i, '').trim();
    const period = toPeriod(selectedDate);
    const fromPeriod = periodMinus(period, 2); // 3-month lookback for analysis

    setLoading(true);
    setError(null);
    setCurrentCC(cleanCC);
    setCurrentPeriod(period);

    try {
      const [outletRes, perfRes] = await Promise.all([
        api.getOutlet(cleanCC),
        api.getPerformance(cleanCC, period),
      ]);

      console.log('RAW PERFORMANCE RESPONSE:', JSON.stringify(perfRes, null, 2));

      setDashboardData({ outlet: outletRes, performance: perfRes });

      // Build default targets from API targets or from performance data
      let apiTargets: any = null;
      try {
        apiTargets = await api.getTargets(cleanCC, period);
      } catch {
        // targets may not exist yet — use values from performance
      }

      if (apiTargets?.targets) {
        const fuel: Record<string, number> = {};
        const nonFuel: Record<string, number> = {};
        for (const t of apiTargets.targets) {
          if (['MS', 'HSD', 'SPEED'].includes(t.product_code)) {
            fuel[t.product_code] = t.target_value;
          } else {
            nonFuel[t.product_code] = t.target_value;
          }
        }
        setTargets({ fuel, nonFuel });
      } else {
        // Build from performance response
        const fuel: Record<string, number> = {};
        const nonFuel: Record<string, number> = {};
        for (const f of perfRes.fuel ?? []) {
          fuel[f.product_code] = f.target ?? 0;
        }
        for (const f of perfRes.non_fuel ?? []) {
          nonFuel[f.product_code] = f.target ?? 0;
        }
        setTargets({ fuel, nonFuel });
      }

      // Fetch analysis and leaderboard in background (don't block dashboard)
      const territory = outletRes.territory_code ?? 'DELHI-W';
      Promise.all([
        api.getAnalysis(cleanCC, fromPeriod, period).catch(() => null),
        api.getLeaderboard(territory).catch(() => null),
        api.getMarketShareStatus(COMPETITION_ID).catch(() => null),
      ]).then(([analysisRes, leaderboardRes, msStatusRes]) => {
        if (analysisRes) setAnalysisData({ analysis: analysisRes, performance: perfRes });
        if (leaderboardRes) setLeaderboardData({ competition: leaderboardRes });
        if (msStatusRes) setMarketShareStatus(msStatusRes);
      });
    } catch (err: any) {
      if (err.status === 403) {
        setError("You don't have access to this outlet.");
      } else if (err.status === 404) {
        setError('Outlet not found. Check the CC number and try again.');
      } else {
        setError(err.message || 'Failed to fetch data. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('bpcl_token');
    localStorage.removeItem('bpcl_user');
    onLogout();
  };

  const handleSaveTargets = async (newTargets: any) => {
    setTargets(newTargets);
    // Persist to API
    if (currentCC && currentPeriod) {
      try {
        const body: any[] = [];
        for (const [code, val] of Object.entries(newTargets.fuel)) {
          body.push({ product_code: code, target_value: val });
        }
        for (const [code, val] of Object.entries(newTargets.nonFuel)) {
          body.push({ product_code: code, target_value: val });
        }
        await api.setTargets(currentCC, { period: currentPeriod, targets: body });
      } catch {
        // non-fatal: local state already updated
      }
    }
  };

  const handleViewScorecard = async (cc: string, competitionId: string) => {
    try {
      const data = await api.getDealerScorecard(cc, competitionId);
      setScorecardData(data);
      setSelectedDealer({ cc, competitionId });
    } catch (err) {
      console.error('Failed to load scorecard:', err);
    }
  };

  const handleCloseScorecard = () => {
    setSelectedDealer(null);
    setScorecardData(null);
  };

  if (showProfile) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navigation
          activeTab={activeTab}
          onTabChange={setActiveTab}
          onProfileClick={() => setShowProfile(true)}
          onLogout={handleLogout}
        />
        <div className="max-w-[1440px] mx-auto px-8 py-8">
          <ManagerProfile onBack={() => setShowProfile(false)} />
        </div>
      </div>
    );
  }

  const hasDashboard = !!dashboardData;
  console.log('DASHBOARD DATA STATE:', JSON.stringify(dashboardData, null, 2));
  const roData = hasDashboard ? buildROData(dashboardData.outlet, currentCC) : null;
  const kpiData = hasDashboard ? buildKPIData(dashboardData.performance) : null;
  const fuelRows = hasDashboard ? buildFuelRows(dashboardData.performance.fuel ?? []) : [];
  const nonFuelRows = hasDashboard ? buildNonFuelRows(dashboardData.performance.non_fuel ?? []) : [];

  console.log('FUEL ROWS:', JSON.stringify(fuelRows));
  console.log('NON_FUEL ROWS:', JSON.stringify(nonFuelRows));

  const hasAnalysis = !!analysisData;
  const analysisKPIs = hasAnalysis ? buildAnalysisKPIs(analysisData.analysis, analysisData.performance) : null;
  const chartData = hasAnalysis ? buildChartData(analysisData.analysis) : null;

  const hasLeaderboard = !!leaderboardData;

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation
        activeTab={activeTab}
        onTabChange={setActiveTab}
        onProfileClick={() => setShowProfile(true)}
        onLogout={handleLogout}
        showLeaderboard
        showTerritory={user?.role === 'admin' || user?.role === 'territory_manager'}
        showAdmin={user?.role === 'admin'}
      />

      <div className="max-w-[1440px] mx-auto px-8 py-8">
        <InputSection onFetch={handleFetch} />

        {loading && (
          <div className="mt-8 flex items-center justify-center">
            <div
              className="animate-spin rounded-full h-12 w-12 border-b-2"
              style={{ borderColor: '#007BC9' }}
            />
          </div>
        )}

        {error && !loading && <ErrorState message={error} onRetry={() => handleFetch(currentCC, new Date())} />}

        {hasDashboard && !loading && !error && (
          <div className="mt-8">
            {activeTab === 'dashboard' && (
              <>
                <ROHeader
                  data={roData}
                  onSetTargets={() => setShowTargetsModal(true)}
                  canSetTargets={canSetTargets}
                />
                <KPIStrip data={kpiData} />
                <div className="space-y-6">
                  <PerformanceTable title="Fuel Performance" data={fuelRows} showVolume={false} />
                  <PerformanceTable title="Non-Fuel Performance" data={nonFuelRows} showVolume={false} />
                </div>
              </>
            )}

            {activeTab === 'analysis' && (
              <>
                <ROHeader data={roData} />
                {hasAnalysis ? (
                  <>
                    <AnalysisKPIs data={analysisKPIs} />
                    <AnalysisCharts {...chartData} />
                  </>
                ) : (
                  <div className="mt-8 text-center text-gray-500 py-12">
                    Loading analysis data…
                  </div>
                )}
              </>
            )}

            {activeTab === 'leaderboard' && (
              <>
                {hasLeaderboard ? (
                  <LeaderboardView
                    competition={leaderboardData!.competition}
                    marketShareStatus={marketShareStatus}
                    onViewScorecard={(cc: string) =>
                      handleViewScorecard(cc, leaderboardData!.competition.competition_id)
                    }
                  />
                ) : (
                  <div className="mt-8 text-center text-gray-500 py-12">
                    Loading leaderboard…
                  </div>
                )}
              </>
            )}

            {activeTab === 'territory' && (
              <TerritoryView
                onSelectDealer={(cc: string) => {
                  setCurrentCC(cc);
                  setActiveTab('dashboard');
                }}
              />
            )}

            {activeTab === 'admin' && <AdminView />}
          </div>
        )}

        {!hasDashboard && !loading && !error && (
          <div className="mt-16 text-center">
            <div
              className="w-16 h-16 rounded-full mx-auto mb-4 flex items-center justify-center"
              style={{ backgroundColor: '#007BC9' }}
            >
              <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
            <h3 className="text-xl text-gray-900 mb-2">Welcome to BPCL Insight</h3>
            <p className="text-gray-600">Enter a CC Number and select a month to view performance data</p>
          </div>
        )}

        {showTargetsModal && targets && (
          <SetTargetsModal
            onClose={() => setShowTargetsModal(false)}
            onSave={handleSaveTargets}
            initialTargets={targets}
          />
        )}

        {selectedDealer && scorecardData && (
          <DealerScorecard
            data={scorecardData}
            competitionName={leaderboardData?.competition?.name || 'Boost and Win'}
            competitionPeriod={leaderboardData?.competition?.period || ''}
            onClose={handleCloseScorecard}
          />
        )}
      </div>
    </div>
  );
}
