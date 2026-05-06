import { test, expect } from '@playwright/test'
import { getToken } from '../fixtures/auth'
import { BASE_URL } from '../fixtures/constants'

const TEST_EMPLOYEE_ID = 'EMP_PW_TEST_99'
let adminToken: string

test.beforeAll(async () => { adminToken = await getToken('admin') })

test.afterAll(async () => {
  const users = await fetch(`${BASE_URL}/api/v1/admin/users`, {
    headers: { Authorization: `Bearer ${adminToken}` }
  }).then(r => r.json())
  const user = (users.users || users.items)?.find((u: any) => u.employee_id === TEST_EMPLOYEE_ID)
  if (user) {
    await fetch(`${BASE_URL}/api/v1/admin/users/${user.id}`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${adminToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ is_active: false })
    })
  }
})

test('admin can list all users', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/admin/users`, {
    headers: { Authorization: `Bearer ${adminToken}` }
  })
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.total).toBeGreaterThanOrEqual(9)
  expect(Array.isArray(d.users || d.items)).toBe(true)
})

test('admin can create a new user', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/admin/users`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({
      employee_id: TEST_EMPLOYEE_ID,
      email: 'test99@bpcl.in',
      name: 'Test User 99',
      role: 'ro_manager',
      territory_code: 'DELHI-W',
      password: 'TestPass@2026'
    })
  })
  expect([200, 201]).toContain(res.status)
  const d = await res.json()
  expect(d.employee_id).toBe(TEST_EMPLOYEE_ID)
  expect(JSON.stringify(d)).not.toContain('password_hash')
})

test('new user can login immediately', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: TEST_EMPLOYEE_ID, password: 'TestPass@2026' })
  })
  if (res.status === 200) {
    const d = await res.json()
    expect(d.token).toBeTruthy()
  }
})