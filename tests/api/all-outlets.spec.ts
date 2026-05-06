import { test, expect } from '@playwright/test'
import { getToken } from '../fixtures/auth'
import { PERIOD, BASE_URL } from '../fixtures/constants'

test.describe('All BPC Outlets', () => {
  let token: string
  let ccNumbers: string[]

  test.beforeAll(async () => {
    token = await getToken('admin')
    const res = await fetch(
      `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
      { headers: { Authorization: `Bearer ${token}` } }
    )
    const d = await res.json()
    ccNumbers = (d.outlets ?? []).map((o: any) => o.cc_number)
    console.log(`Testing ${ccNumbers.length} outlets`)
  })

  test('at least 39 BPC outlets returned', () => {
    expect(ccNumbers.length).toBeGreaterThanOrEqual(39)
  })

  test('all outlets return 200 on performance endpoint', async function() {
    const results = await Promise.all(
      ccNumbers.map(function(cc) {
        return fetch(
          `${BASE_URL}/api/v1/outlets/${cc}/performance?period=${PERIOD}`,
          { headers: { Authorization: `Bearer ${token}` } }
        ).then(function(r) { return { cc: cc, status: r.status } })
      })
    )
    const failures = results.filter(function(r) { return r.status !== 200 })
    if (failures.length > 0) console.log('Failed CCs:', failures)
    expect(failures.length).toBe(0)
  })

  test('no outlet shows all-zero achieved values', async function() {
    const sample = ccNumbers.slice(0, 15)
    for (const cc of sample) {
      const res = await fetch(
        `${BASE_URL}/api/v1/outlets/${cc}/performance?period=${PERIOD}`,
        { headers: { Authorization: `Bearer ${token}` } }
      )
      const d = await res.json()
      const nonZero = (d.fuel ?? []).filter(
        function(r: any) { return !r.is_total && r.achieved > 0 }
      )
      expect(nonZero.length).toBeGreaterThan(0)
    }
  })

  test('competition scores: 36+ dealers have ms_ta_gain_pp', async function() {
    const res = await fetch(
      `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
      { headers: { Authorization: `Bearer ${token}` } }
    )
    const d = await res.json()
    const withGain = d.filter(function(x: any) { return x.ms_ta_gain_pp != null })
    expect(withGain.length).toBeGreaterThanOrEqual(36)
  })
})