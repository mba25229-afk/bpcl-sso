import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { BreakdownCards } from '../BreakdownCards'

const props = {
  fuelData:    { achieved: 5616, target: 6849 },
  nonFuelData: { achieved: 1584, target: 1932 },
  paymentData: { achieved: 1200, target: 1500 },
}

describe('BreakdownCards', () => {

  it('renders three cards: Fuel, Non-Fuel, Payment', () => {
    render(<BreakdownCards {...props} />)
    expect(screen.getByText('Fuel')).toBeInTheDocument()
    expect(screen.getByText('Non-Fuel')).toBeInTheDocument()
    expect(screen.getByText('Payment')).toBeInTheDocument()
  })

  it('shows achieved value for Fuel (₹5,616)', () => {
    render(<BreakdownCards {...props} />)
    expect(screen.getByText('₹5,616')).toBeInTheDocument()
  })

  it('shows target for Non-Fuel', () => {
    render(<BreakdownCards {...props} />)
    expect(screen.getByText('Target: ₹1,932')).toBeInTheDocument()
  })

  it('calculates and shows achievement % for Payment (80.0%)', () => {
    render(<BreakdownCards {...props} />)
    expect(screen.getByText('Target Achieved: 80.0%')).toBeInTheDocument()
  })

  it('renders three progress bars', () => {
    render(<BreakdownCards {...props} />)
    const bars = document.querySelectorAll('[data-testid="breakdown-bar"]')
    expect(bars).toHaveLength(3)
  })

  it('caps achievement at 100% when over-target', () => {
    const overProps = {
      fuelData: { achieved: 9999, target: 1000 },
      nonFuelData: props.nonFuelData,
      paymentData: props.paymentData,
    }
    render(<BreakdownCards {...overProps} />)
    expect(screen.getByText('Target Achieved: 100.0%')).toBeInTheDocument()
  })

})
