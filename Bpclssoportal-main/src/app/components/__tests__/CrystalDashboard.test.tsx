import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router';
import { CrystalDashboard } from '../CrystalDashboard';
import * as apiModule from '../../../lib/api';

const DEALER_ROWS = Array.from({ length: 40 }, (_, i) => ({
  cc_code: `1000${i + 1}`,
  ro_name: `Dealer ${i + 1}`,
  ufill: 50 + i,
  qoc: 3 + i,
  speed_kl: 5.0 + i * 0.1,
  ms_kl: 80.0 + i,
  hsd_kl: 10.0 + i * 0.2,
  ufill_target: 60,
  qoc_target: 5,
  speed_target: 6.0,
  ms_target: 90.0,
  hsd_target: 12.0,
}));

const MOCK_DASHBOARD = {
  month: '2026-05',
  as_of: '2026-05-11',
  dealers: DEALER_ROWS,
};

function renderDashboard() {
  return render(
    <MemoryRouter>
      <CrystalDashboard />
    </MemoryRouter>
  );
}

describe('CrystalDashboard', () => {
  beforeEach(() => {
    vi.spyOn(apiModule.crystalApi, 'getMTDDashboard').mockResolvedValue(MOCK_DASHBOARD);
    vi.spyOn(apiModule.crystalApi, 'getCronStatus').mockResolvedValue(null);
  });

  it('renders 40 dealer rows', async () => {
    renderDashboard();
    await waitFor(() => {
      expect(screen.getByTestId('crystal-dashboard-table')).toBeInTheDocument();
    });
    const rows = screen.getAllByText(/^Dealer \d+$/);
    expect(rows).toHaveLength(40);
  });

  it('shows red badge when achievement < 80% of target', async () => {
    // UFILL: actual=50, target=60 → 83% → amber; let's override with a low value
    vi.spyOn(apiModule.crystalApi, 'getMTDDashboard').mockResolvedValue({
      ...MOCK_DASHBOARD,
      dealers: [{
        ...DEALER_ROWS[0],
        ufill: 10,        // 10/60 = 16% → red
        ufill_target: 60,
      }],
    });
    renderDashboard();
    await waitFor(() => screen.getByText('Dealer 1'));
    // find the UFILL cell badge (value = 10) — it should have red class
    const badge = screen.getByText('10');
    expect(badge.className).toMatch(/red/);
  });

  it('shows green badge when achievement >= 100% of target', async () => {
    vi.spyOn(apiModule.crystalApi, 'getMTDDashboard').mockResolvedValue({
      ...MOCK_DASHBOARD,
      dealers: [{
        ...DEALER_ROWS[0],
        ufill: 60,        // 60/60 = 100% → green
        ufill_target: 60,
      }],
    });
    renderDashboard();
    await waitFor(() => screen.getByText('Dealer 1'));
    const badge = screen.getByText('60');
    expect(badge.className).toMatch(/green/);
  });

  it('displays "–" for null actuals, not 0', async () => {
    vi.spyOn(apiModule.crystalApi, 'getMTDDashboard').mockResolvedValue({
      ...MOCK_DASHBOARD,
      dealers: [{ ...DEALER_ROWS[0], ufill: null, ufill_target: 60 }],
    });
    renderDashboard();
    await waitFor(() => screen.getByText('Dealer 1'));
    expect(screen.getAllByText('–').length).toBeGreaterThan(0);
  });

  it('opens report card modal on row click', async () => {
    renderDashboard();
    await waitFor(() => screen.getByText('Dealer 1'));
    await userEvent.click(screen.getByText('Dealer 1'));
    expect(screen.getByText('Target')).toBeInTheDocument();
    expect(screen.getByText('MTD')).toBeInTheDocument();
  });
});
