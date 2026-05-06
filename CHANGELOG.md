# Changelog - BPCL SSO Portal

## 2026-05-03

### Fixed
- RO (ro_manager) role now correctly receives 403 on territory outlets endpoint
- Admin users endpoint now returns user list for admin role (was returning null due to incorrect claims extraction using ctx.Value instead of service.ExtractClaims)
- Fixed product ID mismatch: SQL was using string ('MS', 'HSD', 'SPEED') but DB has numeric IDs (1, 2, 3)

### Verified Working
- GET /api/v1/territory/outlets - TM sees own territory (6 outlets, 95.77% avg), RO gets 403
- GET /api/v1/admin/users - Admin sees 9 users, TM/RO get 403
- Security headers: X-Content-Type-Options: nosniff, X-Frame-Options: DENY

### Phase 9-10 Status
- Territory Manager Overview: COMPLETE
- Multi-Period Trend: COMPLETE (TerritoryView.tsx has 12-month chart)
- Production Readiness: COMPLETE (configs, docker-compose.prod.yml)