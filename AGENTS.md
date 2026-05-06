# AGENTS.md — BPCL SSO Portal
## Read this before doing anything. This is your map, not your manual.

---

## What This System Is

A two-part platform for BPCL Delhi retail operations:

1. **SSO Dashboard** — Sales Officers view per-outlet KPIs (fuel + non-fuel performance vs. targets)
2. **Boost and Win** — Monthly dealer competition with a 100-point scoring engine and leaderboard

Backend: Go 1.22 + Chi + pgx/v5 + PostgreSQL (no ORM)
Frontend: React + Vite (pre-built at `../Bpclssoportal-main/`)

---

## Execution Order — Follow This. Do Not Skip Phases.

| Phase | Doc | What gets built |
|---|---|---|
| 1 | `docs/exec-plans/chunk-1-database-schema.md` | All 10 migration files, schema only |
| 2 | `docs/exec-plans/chunk-2-seed-data.md` | Realistic seed SQL for all tables |
| 3 | `docs/exec-plans/chunk-3-backend-foundation.md` | go.mod, config, db, models |
| 4 | `docs/exec-plans/chunk-4-repository-layer.md` | All repository SQL functions |
| 5 | `docs/exec-plans/chunk-5-service-handler-layer.md` | Business logic + HTTP handlers |
| 6 | `docs/exec-plans/chunk-6-router-main.md` | Router, middleware, main.go |
| 7 | `docs/exec-plans/chunk-7-frontend-connection.md` | CORS, env, frontend API wiring |

---

## Critical Design Rules (enforced in every phase)

- Handler → Service → Repository. Never skip a layer. Never call DB from handler.
- No ORM. Raw pgx SQL only. No gorm, no ent.
- No fmt.Sprintf in SQL. Parameterized queries only ($1, $2...).
- Blank ≠ Zero in scoring. See `docs/design-docs/scoring-engine.md` Bug #2.
- Access check is the FIRST call in every outlet-scoped service function.
- Period is always stored as DATE = first day of month (2025-04 → 2025-04-01).

---

## Domain Knowledge

| Topic | Where to read |
|---|---|
| 100-point scoring system + 5 bugs | `docs/design-docs/scoring-engine.md` |
| Delhi Master market share logic | `docs/design-docs/data-pipeline.md` |
| Full system architecture | `docs/ARCHITECTURE.md` |
| API contracts (request/response shapes) | `architecture/CONTRACTS.md` |
| DB schema + data models | `architecture/SYSTEM_DESIGN.md` |

---

## Access Control (memorize this)

| Role | Own outlet | Territory outlets | All outlets | Set targets |
|---|---|---|---|---|
| ro_manager | ✅ | ❌ | ❌ | ❌ |
| territory_manager | ✅ | ✅ | ❌ | ✅ |
| admin | ✅ | ✅ | ✅ | ✅ |

Check: `outlet.territory_code == user.territory_code OR user.role = 'admin'`

---

## Products Reference

| Code | Category | Unit |
|---|---|---|
| MS | fuel | kL |
| HSD | fuel | kL |
| SPEED | fuel | kL |
| QOC | non_fuel | Nos |
| Lubricants | non_fuel | ₹ |
| UFill | non_fuel | Nos |
| SBI | non_fuel | Nos |
| BeCafe | non_fuel | ₹ |

---

## When Something Breaks

1. Check the relevant exec-plan doc for the verification step
2. Check `docs/design-docs/scoring-engine.md` for scoring logic errors
3. Check `architecture/CONTRACTS.md` for shape mismatches
4. Never "try harder" — identify what capability is missing and add it to docs first

---

## What Agents Cannot See (encode it or it doesn't exist)

- Slack discussions about architecture decisions → encode in docs/design-docs/
- Excel column format of actual upload files → document in parser/excel.go header comment
- BPCL business rules not in these docs → ask the human, then write it down before building
