import { MapPin, Target } from 'lucide-react'

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
          <div
            data-testid="rating-badge"
            className="px-4 py-2 rounded-lg"
            style={{ backgroundColor: '#007BC9' }}
          >
            <span className="text-white text-sm">Rating: {data.rating ?? '–'}</span>
          </div>
          {canSetTargets && onSetTargets && (
            <button
              onClick={onSetTargets}
              className="px-4 py-2 rounded-lg transition-all hover:shadow-lg shadow-md flex items-center gap-2"
              style={{ backgroundColor: '#FFE000' }}
            >
              <Target size={16} className="text-gray-900" />
              <span className="text-sm text-gray-900 font-medium">Set Targets</span>
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
