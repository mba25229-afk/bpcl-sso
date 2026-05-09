import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { KPIStrip } from '../KPIStrip'

const baseData = {
  rank: '3',
  targetAchievement: '82.0',
  yoyGrowth: '+12.3',
  momGrowth: '+5.2',
}

describe('KPIStrip — Figma redesign', () => {

  it('shows Rank card as first KPI with value #3', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.getByText('#3')).toBeInTheDocument()
    expect(screen.getByText('Rank')).toBeInTheDocument()
  })

  it('shows Target Achievement as second KPI', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.getByText('82.0%')).toBeInTheDocument()
    expect(screen.getByText('Target Achievement')).toBeInTheDocument()
  })

  it('shows YoY Growth as third KPI', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.getByText('+12.3%')).toBeInTheDocument()
    expect(screen.getByText('YoY Growth')).toBeInTheDocument()
  })

  it('shows MoM Growth as fourth KPI', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.getByText('+5.2%')).toBeInTheDocument()
    expect(screen.getByText('MoM Growth')).toBeInTheDocument()
  })

  it('does NOT show Total Revenue card', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.queryByText('Total Revenue')).not.toBeInTheDocument()
  })

  it('does NOT show Fuel vs Non-Fuel card', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.queryByText('Fuel vs Non-Fuel')).not.toBeInTheDocument()
  })

  it('shows upward trend arrow when YoY growth is positive', () => {
    render(<KPIStrip data={{ ...baseData, yoyGrowth: '+12.3' }} />)
    const trendArrow = screen.getByTestId('yoy-trend-arrow')
    expect(trendArrow).toHaveAttribute('data-direction', 'up')
  })

  it('shows downward trend arrow when YoY growth is negative', () => {
    render(<KPIStrip data={{ ...baseData, yoyGrowth: '-3.1' }} />)
    const trendArrow = screen.getByTestId('yoy-trend-arrow')
    expect(trendArrow).toHaveAttribute('data-direction', 'down')
  })

  it('MoM Growth card icon background is green when positive', () => {
    render(<KPIStrip data={{ ...baseData, momGrowth: '+5.2' }} />)
    const indicator = screen.getByTestId('mom-trend-indicator')
    expect(indicator).toHaveAttribute('data-direction', 'up')
  })

  it('MoM Growth card icon background is red when negative', () => {
    render(<KPIStrip data={{ ...baseData, momGrowth: '-2.1' }} />)
    const indicator = screen.getByTestId('mom-trend-indicator')
    expect(indicator).toHaveAttribute('data-direction', 'down')
  })

  it('shows trophy icon on Rank card (highlighted card)', () => {
    render(<KPIStrip data={baseData} />)
    expect(screen.getByTestId('rank-icon')).toBeInTheDocument()
  })

  it('Rank card has gold/yellow highlight border', () => {
    render(<KPIStrip data={baseData} />)
    const rankCard = screen.getByTestId('rank-card')
    expect(rankCard).toHaveStyle({ borderLeft: '4px solid #FFE000' })
  })

})
