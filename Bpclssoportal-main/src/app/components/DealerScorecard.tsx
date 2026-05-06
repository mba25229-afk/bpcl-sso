import { Award, X, Download } from 'lucide-react';

interface ScorecardParameter {
  name: string;
  max_marks: number;
  score: number | null;
}

interface ScorecardData {
  cc_number: string;
  outlet_name: string;
  competition_id: string;
  total_score: number | null;
  rank: number | null;
  parameters: ScorecardParameter[];
}

interface DealerScorecardProps {
  data: ScorecardData;
  competitionName: string;
  competitionPeriod: string;
  onClose: () => void;
}

function ScoreBar({ score, maxMarks }: { score: number | null; maxMarks: number }) {
  const scoreValue = score ?? 0;
  const percentage = maxMarks > 0 ? (Math.abs(scoreValue) / maxMarks) * 100 : 0;
  const isNegative = scoreValue < 0;
  const barColor = isNegative ? '#DC2626' : scoreValue === 0 ? '#9CA3AF' : '#007BC9';
  const barWidth = Math.min(percentage, 100);

  return (
    <div className="flex items-center gap-3">
      <div className="w-24 h-3 rounded-full overflow-hidden bg-gray-200">
        <div
          className="h-full rounded-full transition-all duration-500"
          style={{
            width: `${barWidth}%`,
            backgroundColor: barColor,
          }}
        />
      </div>
      <span
        className="text-sm font-medium w-10"
        style={{ color: isNegative ? '#DC2626' : '#1F2937' }}
      >
        {scoreValue.toFixed(1)}
      </span>
    </div>
  );
}

export function DealerScorecard({ data, competitionName, competitionPeriod, onClose }: DealerScorecardProps) {
  const totalScore = data.total_score ?? 0;
  const rank = data.rank ?? 0;

  const fuelParams = data.parameters.filter(
    (p) =>
      p.name.includes('MS') ||
      p.name.includes('HSD') ||
      p.name.includes('SPEED')
  );
  const nonFuelParams = data.parameters.filter(
    (p) =>
      p.name.includes('QOC') ||
      p.name.includes('Lubricants') ||
      p.name.includes('UFill') ||
      p.name.includes('Speed')
  );
  const complianceParams = data.parameters.filter(
    (p) =>
      p.name.includes('Cleanliness') ||
      p.name.includes('Sangam') ||
      p.name.includes('Google') ||
      p.name.includes('IPS')
  );
  const bonusParams = data.parameters.filter((p) => p.name.includes('Bonus'));

  const medals = ['', '', ''];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <div>
            <h2 className="text-xl text-gray-900">{data.outlet_name}</h2>
            <p className="text-gray-600">CC: {data.cc_number}</p>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X size={20} className="text-gray-600" />
          </button>
        </div>

        <div className="px-6 py-4 bg-gray-50 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold"
                style={{
                  backgroundColor:
                    rank === 1
                      ? '#FFE000'
                      : rank === 2
                        ? '#E5E7EB'
                        : rank === 3
                          ? '#FCD9A0'
                          : '#F3F4F6',
                }}
              >
                {rank <= 3 ? (
                  <span>{medals[rank - 1]}</span>
                ) : (
                  rank
                )}
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">
                  {totalScore.toFixed(1)} / 100
                </p>
                <p className="text-gray-600">{competitionName}</p>
              </div>
            </div>
            <p className="text-gray-600">{competitionPeriod}</p>
          </div>
        </div>

        <div className="p-6 space-y-6">
          {fuelParams.length > 0 && (
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              <div
                className="px-6 py-3"
                style={{ backgroundColor: '#007BC9' }}
              >
                <h3 className="text-white">FUEL PERFORMANCE</h3>
              </div>
              <div className="p-6 space-y-4">
                {fuelParams.map((param) => (
                  <div key={param.name} className="flex items-center justify-between">
                    <span className="text-gray-700 w-40">{param.name}</span>
                    <ScoreBar score={param.score} maxMarks={param.max_marks} />
                    <span className="text-gray-500 text-sm w-12">
                      /{param.max_marks}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {nonFuelParams.length > 0 && (
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              <div
                className="px-6 py-3"
                style={{ backgroundColor: '#007BC9' }}
              >
                <h3 className="text-white">NON-FUEL PERFORMANCE</h3>
              </div>
              <div className="p-6 space-y-4">
                {nonFuelParams.map((param) => (
                  <div key={param.name} className="flex items-center justify-between">
                    <span className="text-gray-700 w-40">{param.name}</span>
                    <ScoreBar score={param.score} maxMarks={param.max_marks} />
                    <span className="text-gray-500 text-sm w-12">
                      /{param.max_marks}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {complianceParams.length > 0 && (
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              <div
                className="px-6 py-3"
                style={{ backgroundColor: '#007BC9' }}
              >
                <h3 className="text-white">COMPLIANCE & CUSTOMER</h3>
              </div>
              <div className="p-6 space-y-4">
                {complianceParams.map((param) => (
                  <div key={param.name} className="flex items-center justify-between">
                    <span className="text-gray-700 w-40">{param.name}</span>
                    <ScoreBar score={param.score} maxMarks={param.max_marks} />
                    <span className="text-gray-500 text-sm w-12">
                      /{param.max_marks}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {bonusParams.length > 0 && (
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              <div
                className="px-6 py-3"
                style={{ backgroundColor: '#FFE000' }}
              >
                <h3 className="text-gray-900">BONUS</h3>
              </div>
              <div className="p-6 space-y-4">
                {bonusParams.map((param) => (
                  <div key={param.name} className="flex items-center justify-between">
                    <span className="text-gray-700 w-40">{param.name}</span>
                    <ScoreBar score={param.score} maxMarks={param.max_marks} />
                    <span className="text-gray-500 text-sm w-12">
                      /{param.max_marks}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="sticky bottom-0 bg-white border-t border-gray-200 px-6 py-4 flex items-center justify-center">
          <button
            onClick={() => window.print()}
            className="px-6 py-2 rounded-lg flex items-center gap-2 text-white transition-all hover:shadow-md"
            style={{ backgroundColor: '#007BC9' }}
          >
            <Download size={16} />
            Download PDF
          </button>
        </div>
      </div>
    </div>
  );
}