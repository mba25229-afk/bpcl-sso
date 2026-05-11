import { useEffect, useState, useCallback } from 'react';
import { crystalApi } from '../../lib/api';

// ── Types ──────────────────────────────────────────────────────────────────────

interface MTDRow {
  cc_code: string;
  ro_name: string;
  ufill: number | null;
  qoc: number | null;
  speed_kl: number | null;
  ms_kl: number | null;
  hsd_kl: number | null;
  ufill_target: number | null;
  qoc_target: number | null;
  speed_target: number | null;
  ms_target: number | null;
  hsd_target: number | null;
}

interface DashboardPayload {
  month: string;
  as_of: string;
  dealers: MTDRow[];
}

interface CronStatus {
  last_run_at: string | null;
  status: string | null;
  detail: string | null;
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function pct(actual: number | null, target: number | null): number | null {
  if (actual == null || target == null || target === 0) return null;
  return (actual / target) * 100;
}

type TrafficLight = 'green' | 'amber' | 'red' | 'none';

function trafficLight(actual: number | null, target: number | null): TrafficLight {
  const p = pct(actual, target);
  if (p == null) return 'none';
  if (p >= 100) return 'green';
  if (p >= 80) return 'amber';
  return 'red';
}

const LIGHT_CLASSES: Record<TrafficLight, string> = {
  green: 'bg-green-100 text-green-800',
  amber: 'bg-yellow-100 text-yellow-800',
  red:   'bg-red-100 text-red-800',
  none:  'bg-gray-100 text-gray-500',
};

function MetricCell({
  actual,
  target,
  fmt,
}: {
  actual: number | null;
  target: number | null;
  fmt?: (v: number) => string;
}) {
  const light = trafficLight(actual, target);
  const display = actual == null ? '–' : (fmt ? fmt(actual) : String(actual));
  return (
    <td className="px-3 py-3 text-center">
      <span className={`px-2 py-0.5 rounded text-sm font-medium ${LIGHT_CLASSES[light]}`}>
        {display}
      </span>
    </td>
  );
}

function fmtKL(v: number) {
  return v.toFixed(1);
}

function currentMonthStr() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

// ── Drill-down modal ───────────────────────────────────────────────────────────

function ReportCard({ row, onClose }: { row: MTDRow; onClose: () => void }) {
  const metrics: { key: keyof MTDRow; label: string; target: keyof MTDRow; fmt?: (v: number) => string }[] = [
    { key: 'ufill',    label: 'UFILL',     target: 'ufill_target' },
    { key: 'qoc',      label: 'QOC',       target: 'qoc_target' },
    { key: 'speed_kl', label: 'SPEED (KL)', target: 'speed_target', fmt: fmtKL },
    { key: 'ms_kl',    label: 'MS (KL)',    target: 'ms_target',    fmt: fmtKL },
    { key: 'hsd_kl',   label: 'HSD (KL)',   target: 'hsd_target',   fmt: fmtKL },
  ];

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onClick={onClose}>
      <div
        className="bg-white rounded-2xl shadow-xl w-full max-w-lg"
        onClick={e => e.stopPropagation()}
      >
        <div className="px-6 py-4 border-b flex items-center justify-between" style={{ backgroundColor: '#007BC9' }}>
          <div>
            <p className="text-white font-semibold text-lg">{row.ro_name}</p>
            <p className="text-blue-100 text-sm">{row.cc_code}</p>
          </div>
          <button onClick={onClose} className="text-white hover:text-blue-200 text-2xl leading-none">×</button>
        </div>

        <div className="p-6">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left pb-2 text-gray-500 font-medium">Metric</th>
                <th className="text-right pb-2 text-gray-500 font-medium">Target</th>
                <th className="text-right pb-2 text-gray-500 font-medium">MTD</th>
                <th className="text-right pb-2 text-gray-500 font-medium">Gap</th>
              </tr>
            </thead>
            <tbody>
              {metrics.map(({ key, label, target, fmt }) => {
                const act = row[key] as number | null;
                const tgt = row[target] as number | null;
                const gap = act != null && tgt != null ? act - tgt : null;
                const light = trafficLight(act, tgt);
                return (
                  <tr key={key} className="border-b last:border-0">
                    <td className="py-3 font-medium text-gray-700">{label}</td>
                    <td className="py-3 text-right text-gray-500">
                      {tgt == null ? '–' : (fmt ? fmt(tgt) : tgt)}
                    </td>
                    <td className="py-3 text-right">
                      <span className={`px-2 py-0.5 rounded text-xs font-semibold ${LIGHT_CLASSES[light]}`}>
                        {act == null ? '–' : (fmt ? fmt(act) : act)}
                      </span>
                    </td>
                    <td className={`py-3 text-right text-xs ${gap == null ? 'text-gray-400' : gap >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {gap == null ? '–' : `${gap >= 0 ? '+' : ''}${fmt ? fmt(gap) : gap}`}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

// ── Cron status badge ──────────────────────────────────────────────────────────

function CronBadge({ status }: { status: CronStatus | null }) {
  if (!status?.last_run_at) {
    return <span className="px-2 py-1 rounded text-xs bg-gray-100 text-gray-500">ETL: no data yet</span>;
  }
  const ago = Math.round((Date.now() - new Date(status.last_run_at).getTime()) / 3600_000);
  const ok = status.status === 'ok';
  return (
    <span className={`px-2 py-1 rounded text-xs font-medium ${ok ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
      ETL: {ok ? '✓' : '✗'} {ago}h ago
    </span>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────

export function CrystalDashboard() {
  const [month, setMonth] = useState(currentMonthStr);
  const [data, setData] = useState<DashboardPayload | null>(null);
  const [cronStatus, setCronStatus] = useState<CronStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<MTDRow | null>(null);

  const load = useCallback(async (m: string) => {
    setLoading(true);
    setError(null);
    try {
      const [dash, cron] = await Promise.all([
        crystalApi.getMTDDashboard(m),
        crystalApi.getCronStatus().catch(() => null),
      ]);
      setData(dash);
      setCronStatus(cron);
    } catch (e: any) {
      setError(e.message || 'Failed to load dashboard');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(month); }, [load, month]);

  const dealers = data?.dealers ?? [];

  return (
    <div>
      {/* Header bar */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Crystal MTD Dashboard</h2>
          {data && (
            <p className="text-sm text-gray-500 mt-0.5">
              {dealers.length} dealers · as of {data.as_of}
            </p>
          )}
        </div>
        <div className="flex items-center gap-4">
          <CronBadge status={cronStatus} />
          <input
            type="month"
            value={month}
            onChange={e => setMonth(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2"
            style={{ '--tw-ring-color': '#007BC9' } as any}
          />
        </div>
      </div>

      {/* Legend */}
      <div className="flex gap-3 mb-4 text-xs">
        <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-green-100 inline-block border border-green-300" /> ≥ 100%</span>
        <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-yellow-100 inline-block border border-yellow-300" /> 80–99%</span>
        <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-red-100 inline-block border border-red-300" /> &lt; 80%</span>
        <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-gray-100 inline-block border border-gray-300" /> No target</span>
      </div>

      {loading && (
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2" style={{ borderColor: '#007BC9' }} />
        </div>
      )}

      {error && !loading && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700 text-sm">
          {error}
        </div>
      )}

      {!loading && !error && dealers.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-gray-200 shadow-sm">
          <table className="w-full text-sm" data-testid="crystal-dashboard-table">
            <thead>
              <tr style={{ backgroundColor: '#007BC9' }}>
                <th className="px-4 py-3 text-left text-white font-medium sticky left-0" style={{ backgroundColor: '#007BC9' }}>Outlet</th>
                <th className="px-3 py-3 text-center text-white font-medium">UFILL</th>
                <th className="px-3 py-3 text-center text-white font-medium">QOC</th>
                <th className="px-3 py-3 text-center text-white font-medium">SPEED (KL)</th>
                <th className="px-3 py-3 text-center text-white font-medium">MS (KL)</th>
                <th className="px-3 py-3 text-center text-white font-medium">HSD (KL)</th>
                <th className="px-3 py-3 text-center text-white font-medium"></th>
              </tr>
            </thead>
            <tbody>
              {dealers.map((row, i) => (
                <tr
                  key={row.cc_code}
                  className={`border-b border-gray-100 hover:bg-blue-50 transition-colors cursor-pointer ${i % 2 === 0 ? 'bg-white' : 'bg-gray-50'}`}
                  onClick={() => setSelected(row)}
                >
                  <td className="px-4 py-3 sticky left-0 bg-inherit">
                    <p className="font-medium text-gray-900 text-sm">{row.ro_name}</p>
                    <p className="text-xs text-gray-400">{row.cc_code}</p>
                  </td>
                  <MetricCell actual={row.ufill}    target={row.ufill_target} />
                  <MetricCell actual={row.qoc}      target={row.qoc_target} />
                  <MetricCell actual={row.speed_kl} target={row.speed_target} fmt={fmtKL} />
                  <MetricCell actual={row.ms_kl}    target={row.ms_target}  fmt={fmtKL} />
                  <MetricCell actual={row.hsd_kl}   target={row.hsd_target} fmt={fmtKL} />
                  <td className="px-3 py-3 text-center">
                    <button
                      className="text-xs text-blue-600 hover:underline"
                      onClick={e => { e.stopPropagation(); setSelected(row); }}
                    >
                      View
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!loading && !error && dealers.length === 0 && data && (
        <div className="text-center py-16 text-gray-500">
          No dealer data for {month}. Run the ETL to ingest data.
        </div>
      )}

      {selected && <ReportCard row={selected} onClose={() => setSelected(null)} />}
    </div>
  );
}
