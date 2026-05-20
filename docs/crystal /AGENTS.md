# Crystal — Agent Map
> Read this first. It is a map, not a manual. Follow pointers to go deeper.

## What Crystal is
A dealer performance dashboard for BPCL Central Delhi Sales Area.
41 retail outlets (ROs). Daily data ingested from Google Sheets via cron.
Stack: Node/Express backend · React frontend · PostgreSQL · JWT auth.

## Hard rules (non-negotiable)
1. Never write raw SQL — always use the query builder (Knex).
2. Never trust Excel/Sheet data without running it through the Zod schema in `src/etl/schema.ts`.
3. Never store JWT secret in code — read from `process.env.JWT_SECRET`.
4. BPCL brand colors only — see `docs/design-docs/bpcl-tokens.md`.
5. Backend tests in `tests/backend/`, frontend tests in `tests/frontend/`, Playwright in `tests/e2e/`.
6. Every ETL run must POST to the health endpoint — see `docs/design-docs/etl.md`.
7. Session idle timeout = 20 minutes. This is a business requirement, not optional.

## Where to look
| What you need | Go here |
|---|---|
| Data schema (all 41 ROs, all fields) | `docs/crystal /db-schema.md` |
| Auth design (JWT + inactivity) | `docs/crystal /auth.md` |
| ETL design (cron + healthcheck) | `docs/crystal /etl.md` |
| Test strategy | `docs/crystal /test-strategy.md` |
| Active build plan | `docs/exec-plans/active/` |
| BPCL design tokens | `docs/design-docs/bpcl-tokens.md` |

## Dealer master (41 ROs)
See `docs/generated/db-schema.md` → `dealers` table. CC Code is the primary key. Never hardcode dealer names.

## Layer rules
```
types → schema (Zod) → db (Knex) → service → route → controller
```
Cross-cutting (auth, logger, healthcheck) enters only through middleware. No circular deps.
