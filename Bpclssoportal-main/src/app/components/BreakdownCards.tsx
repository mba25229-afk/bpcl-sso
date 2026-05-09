export function BreakdownCards({ fuelData, nonFuelData, paymentData }: {
  fuelData: { achieved: number; target: number }
  nonFuelData: { achieved: number; target: number }
  paymentData: { achieved: number; target: number }
}) {
  const calculatePercentage = (achieved: number, target: number) =>
    Math.min((achieved / target) * 100, 100)

  const getProgressColor = (pct: number) => {
    if (pct >= 90) return '#10b981'
    if (pct >= 70) return '#FFE000'
    return '#ef4444'
  }

  const cards = [
    { title: 'Fuel',     achieved: fuelData.achieved,    target: fuelData.target },
    { title: 'Non-Fuel', achieved: nonFuelData.achieved, target: nonFuelData.target },
    { title: 'Payment',  achieved: paymentData.achieved, target: paymentData.target },
  ]

  return (
    <div className="grid grid-cols-3 gap-6 mb-6">
      {cards.map((card) => {
        const pct = calculatePercentage(card.achieved, card.target)
        const color = getProgressColor(pct)
        return (
          <div key={card.title} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="text-gray-900 mb-4">{card.title}</h3>
            <div className="mb-3">
              <div className="text-3xl text-gray-900 mb-1">₹{card.achieved.toLocaleString()}</div>
              <div className="text-sm text-gray-600">Target: ₹{card.target.toLocaleString()}</div>
            </div>
            <div className="space-y-2">
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  data-testid="breakdown-bar"
                  className="h-2 rounded-full transition-all"
                  style={{ width: `${pct}%`, backgroundColor: color }}
                />
              </div>
              <div className="text-sm" style={{ color }}>
                Target Achieved: {pct.toFixed(1)}%
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
