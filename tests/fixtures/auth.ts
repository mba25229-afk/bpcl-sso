import { Page } from '@playwright/test'
import { CREDS, BASE_URL } from './constants'

const tokenCache: Map<string, { token: string, expires: number }> = new Map()
const TOKEN_TTL = 5 * 60 * 1000 // 5 minutes

export async function loginAs(page: Page, role: 'admin'|'tm'|'ro') {
  await page.goto('/')
  await page.waitForLoadState('networkidle')
  const inputs = page.locator('input')
  await inputs.first().fill(CREDS[role].id)
  await inputs.nth(1).fill(CREDS[role].password)
  await page.locator('button[type="submit"], button:has-text("Sign"), button:has-text("Login")').first().click()
  await page.waitForTimeout(3000)
}

export async function getToken(role: 'admin'|'tm'|'ro'): Promise<string> {
  const now = Date.now()
  const cached = tokenCache.get(role)
  if (cached && cached.expires > now) {
    return cached.token
  }

  // Retry up to 3 times on rate limit
  for (let attempt = 0; attempt < 3; attempt++) {
    const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        employee_id: CREDS[role].id,
        password: CREDS[role].password
      })
    })
    const data = await res.json()
    if (data.token) {
      tokenCache.set(role, { token: data.token, expires: now + TOKEN_TTL })
      return data.token
    }
    if (data.code === 'RATE_LIMITED') {
      await new Promise(r => setTimeout(r, 2000 * (attempt + 1)))
      continue
    }
    throw new Error(`Login failed for ${role}: ${JSON.stringify(data)}`)
  }
  throw new Error(`Rate limited after 3 attempts for ${role}`)
}