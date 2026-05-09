import { TrendingUp, TrendingDown, Target, Award } from 'lucide-react'

export function KPIStrip({ data }: { data: any }) {
  const yoyPositive = !String(data.yoyGrowth ?? '').startsWith('-')
  const momPositive = !String(data.momGrowth ?? '').startsWith('-')

  return (
    <div className="grid grid-cols-4 gap-6 mb-6">
      {/* Card 1: Rank — highlighted */}
      <div
        data-testid="rank-card"
        className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 transition-shadow hover:shadow-md"
        style={{ borderLeft: '4px solid #FFE000' }}
      >
        <div className="flex items-start justify-between mb-3">
          <div
            data-testid="rank-icon"
            className="w-10 h-10 rounded-lg flex items-center justify-center"
            style={{ backgroundColor: '#FFE000' }}
          >
            <Award size={20} className="text-gray-900" />
          </div>
        </div>
        <div className="text-3xl text-gray-900 mb-1">#{data.rank ?? '–'}</div>
        <div className="text-sm text-gray-600">Rank</div>
      </div>

      {/* Card 2: Target Achievement */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 transition-shadow hover:shadow-md">
        <div className="flex items-start justify-between mb-3">
          <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: '#007BC9' }}>
            <Target size={20} className="text-white" />
          </div>
        </div>
        <div className="text-3xl text-gray-900 mb-1">{data.targetAchievement ?? '0.0'}%</div>
        <div className="text-sm text-gray-600">Target Achievement</div>
      </div>

      {/* Card 3: YoY Growth with trend arrow */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 transition-shadow hover:shadow-md">
        <div className="flex items-start justify-between mb-3">
          <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: '#007BC9' }}>
            <TrendingUp size={20} className="text-white" />
          </div>
          <span
            data-testid="yoy-trend-arrow"
            data-direction={yoyPositive ? 'up' : 'down'}
            className={yoyPositive ? 'text-green-500' : 'text-red-500'}
          >
            {yoyPositive ? '↑' : '↓'}
          </span>
        </div>
        <div className="text-3xl text-gray-900 mb-1">{data.yoyGrowth ?? '+0.0'}%</div>
        <div className="text-sm text-gray-600">YoY Growth</div>
      </div>

      {/* Card 4: MoM Growth — icon background is green/red based on direction */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 transition-shadow hover:shadow-md">
        <div className="flex items-start justify-between mb-3">
          <div
            className="w-10 h-10 rounded-lg flex items-center justify-center"
            style={{ backgroundColor: momPositive ? '#10b981' : '#ef4444' }}
          >
            {momPositive
              ? <TrendingUp size={20} className="text-white" />
              : <TrendingDown size={20} className="text-white" />
            }
          </div>
          <div
            data-testid="mom-trend-indicator"
            data-direction={momPositive ? 'up' : 'down'}
            className={`flex items-center gap-1 ${momPositive ? 'text-green-600' : 'text-red-600'}`}
          >
            {momPositive ? '↑' : '↓'}
          </div>
        </div>
        <div className="text-3xl text-gray-900 mb-1">{data.momGrowth ?? '+0.0'}%</div>
        <div className="text-sm text-gray-600">MoM Growth</div>
      </div>
    </div>
  )
}
