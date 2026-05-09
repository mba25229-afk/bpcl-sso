import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { ROHeader } from '../ROHeader'

const baseData = {
  name: 'BPCL Syall Service Station - Kirti Nagar',
  location: 'KIRTI NAGAR DELHI, WEST DELHI, Delhi 110015',
  ccNumber: '12345',
  rating: 'A+',
}

describe('ROHeader — Figma redesign', () => {

  it('shows the outlet name', () => {
    render(<ROHeader data={baseData} />)
    expect(screen.getByText('BPCL Syall Service Station - Kirti Nagar')).toBeInTheDocument()
  })

  it('shows the CC number', () => {
    render(<ROHeader data={baseData} />)
    expect(screen.getByText('12345')).toBeInTheDocument()
  })

  it('shows Rating badge with value A+', () => {
    render(<ROHeader data={baseData} />)
    expect(screen.getByTestId('rating-badge')).toBeInTheDocument()
    expect(screen.getByText('Rating: A+')).toBeInTheDocument()
  })

  it('Rating badge is blue (#007BC9)', () => {
    render(<ROHeader data={baseData} />)
    const badge = screen.getByTestId('rating-badge')
    expect(badge).toHaveStyle({ backgroundColor: '#007BC9' })
  })

  it('does NOT show a Rank badge', () => {
    render(<ROHeader data={baseData} />)
    expect(screen.queryByText(/^Rank:/)).not.toBeInTheDocument()
  })

  it('shows Set Targets button when canSetTargets is true', () => {
    const onSetTargets = () => {}
    render(<ROHeader data={baseData} onSetTargets={onSetTargets} canSetTargets={true} />)
    expect(screen.getByText('Set Targets')).toBeInTheDocument()
  })

  it('Set Targets button has solid gold background (#FFE000), not a border', () => {
    const onSetTargets = () => {}
    render(<ROHeader data={baseData} onSetTargets={onSetTargets} canSetTargets={true} />)
    const btn = screen.getByText('Set Targets').closest('button')!
    expect(btn).toHaveStyle({ backgroundColor: '#FFE000' })
    expect(btn).not.toHaveClass('border-2')
  })

  it('hides Set Targets button when canSetTargets is false', () => {
    render(<ROHeader data={baseData} canSetTargets={false} />)
    expect(screen.queryByText('Set Targets')).not.toBeInTheDocument()
  })

  it('shows location with pin icon area', () => {
    render(<ROHeader data={baseData} />)
    expect(screen.getByText('KIRTI NAGAR DELHI, WEST DELHI, Delhi 110015')).toBeInTheDocument()
  })

})
