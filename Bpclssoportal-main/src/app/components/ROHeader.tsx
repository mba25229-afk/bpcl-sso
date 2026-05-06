import { MapPin, Target } from 'lucide-react';

export function ROHeader({ data, onSetTargets, canSetTargets }: { data: any; onSetTargets?: () => void; canSetTargets?: boolean }) {
  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl text-gray-900 mb-1">{data.name}</h2>
          <div className="flex items-center gap-2 text-gray-600">
            <MapPin size={16} />
            <span>{data.location}</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="px-4 py-2 rounded-lg bg-gray-100">
            <span className="text-sm text-gray-600">CC: </span>
            <span className="font-semibold text-gray-900">{data.ccNumber}</span>
          </div>
          <div className="px-4 py-2 rounded-lg" style={{ backgroundColor: '#007BC9' }}>
            <span className="text-white text-sm">Rank: {data.rank}</span>
          </div>
          {canSetTargets && onSetTargets && (
            <button
              onClick={onSetTargets}
              className="px-4 py-2 rounded-lg border-2 hover:bg-gray-50 transition-colors flex items-center gap-2"
              style={{ borderColor: '#FFE000' }}
            >
              <Target size={16} className="text-gray-700" />
              <span className="text-sm text-gray-700">Set Targets</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
