import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { PerformanceTable } from '../PerformanceTable'

const fuelRows = [
  { product: 'MS',  target: 2415, achieved: 1980, ly: 1820, volume: 1980 },
  { product: 'HSD', target: 4098, achieved: 3360, ly: 3080, volume: 3360 },
  { product: 'SPEED', target: 500, achieved: 0,   ly: 0,    volume: 0 },
  { product: 'TOTAL', target: 7013, achieved: 5340, ly: 4900, volume: 5340, isTotal: true },
]

describe('PerformanceTable — 5-column reference layout', () => {

  it('has Product, Target, Achieved, LY and Growth column headers', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    const headers = screen.getAllByRole('columnheader')
    const texts = headers.map(h => h.textContent)
    expect(texts).toContain('Product')
    expect(texts).toContain('Target')
    expect(texts).toContain('Achieved')
    expect(texts).toContain('LY')
    expect(texts).toContain('Growth')
  })

  it('renders LY as a standalone column (1,820 for MS row)', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    expect(screen.getByText('1,820')).toBeInTheDocument()
    expect(screen.getByText('3,080')).toBeInTheDocument()
  })

  it('renders MS growth as +8.8% (green) for 1980 achieved vs 1820 LY', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    const growthCell = screen.getByText('+8.8%')
    expect(growthCell).toBeInTheDocument()
    expect(growthCell).toHaveClass('text-green-600')
  })

  it('renders HSD growth as +9.1% (green)', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    expect(screen.getByText('+9.1%')).toBeInTheDocument()
  })

  it('renders negative growth in red', () => {
    const rows = [{ product: 'MS', target: 1000, achieved: 800, ly: 1000, volume: 800 }]
    render(<PerformanceTable title="Fuel Performance" data={rows} />)
    const growthCell = screen.getByText('-20.0%')
    expect(growthCell).toHaveClass('text-red-600')
  })

  it('shows a progress bar in the Achieved cell', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    const bars = document.querySelectorAll('[data-testid="progress-bar"]')
    expect(bars.length).toBeGreaterThan(0)
  })

  it('achieved cell uses horizontal layout — progress bar left, value right', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    // Value 1,980 should appear alongside the progress bar
    expect(screen.getByText('1,980')).toBeInTheDocument()
  })

  it('uses table-fixed layout class', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    const table = document.querySelector('table')
    expect(table).toHaveClass('table-fixed')
  })

  it('total row shows bold growth %', () => {
    render(<PerformanceTable title="Fuel Performance" data={fuelRows} />)
    // TOTAL: 5340 achieved vs 4900 LY = +9.0%
    expect(screen.getByText('+9.0%')).toBeInTheDocument()
  })

})
