export function PerformanceTable({ title, data, showVolume = true }: { title: string; data: any[]; showVolume?: boolean }) {
  const calculateProgress = (achieved: number, target: number) => {
    return Math.min((achieved / target) * 100, 100);
  };

  const calculateVsLY = (current: number, ly: number) => {
    if (!ly) return 0;
    return ((current - ly) / ly) * 100;
  };

  const totalRow = data.find(row => row.isTotal);
  const regularRows = data.filter(row => !row.isTotal);

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200" style={{ backgroundColor: '#007BC9' }}>
        <h3 className="text-white">{title}</h3>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm text-gray-700">Product</th>
              <th className="px-6 py-3 text-right text-sm text-gray-700">Target</th>
              <th className="px-6 py-3 text-right text-sm text-gray-700">Achieved</th>
              <th className="px-6 py-3 text-right text-sm text-gray-700">LY</th>
              {showVolume && <th className="px-6 py-3 text-right text-sm text-gray-700">Volume (KL)</th>}
            </tr>
          </thead>
          <tbody>
            {regularRows.map((row, index) => {
              const progress = calculateProgress(row.achieved, row.target);
              const vsLY = calculateVsLY(row.achieved, row.ly);

              return (
                <tr key={index} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 text-gray-900">{row.product}</td>
                  <td className="px-6 py-4 text-right text-gray-700">{row.target.toLocaleString()}</td>
                  <td className="px-6 py-4 text-right">
                    <div className="space-y-1">
                      <div className="text-gray-900">{row.achieved.toLocaleString()}</div>
                      <div className="w-full bg-gray-200 rounded-full h-1.5">
                        <div
                          className="h-1.5 rounded-full transition-all"
                          style={{
                            width: `${progress}%`,
                            backgroundColor: progress >= 90 ? '#10b981' : progress >= 70 ? '#FFE000' : '#f59e0b',
                          }}
                        />
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="space-y-1">
                      <div className="text-gray-700">{row.ly.toLocaleString()}</div>
                      <div className={`text-xs ${vsLY >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {vsLY >= 0 ? '+' : ''}{vsLY.toFixed(1)}%
                      </div>
                    </div>
                  </td>
                  {showVolume && <td className="px-6 py-4 text-right text-gray-900">{row.volume.toLocaleString()}</td>}
                </tr>
              );
            })}
            {totalRow && (
              <tr className="bg-gray-50 border-t-2 border-gray-300">
                <td className="px-6 py-4 font-semibold text-gray-900">{totalRow.product}</td>
                <td className="px-6 py-4 text-right font-semibold text-gray-900">{totalRow.target.toLocaleString()}</td>
                <td className="px-6 py-4 text-right font-semibold text-gray-900">{totalRow.achieved.toLocaleString()}</td>
                <td className="px-6 py-4 text-right font-semibold text-gray-900">{totalRow.ly.toLocaleString()}</td>
                {showVolume && <td className="px-6 py-4 text-right font-semibold text-gray-900">{totalRow.volume.toLocaleString()}</td>}
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
