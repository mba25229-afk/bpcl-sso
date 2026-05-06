# Graph Report - bpcl-portal-api  (2026-05-03)

## Corpus Check
- 77 files · ~55,235 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 480 nodes · 985 edges · 29 communities detected
- Extraction: 50% EXTRACTED · 50% INFERRED · 0% AMBIGUOUS · INFERRED: 495 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]

## God Nodes (most connected - your core abstractions)
1. `New()` - 40 edges
2. `Close()` - 30 edges
3. `writeJSON()` - 25 edges
4. `Handler` - 23 edges
5. `main()` - 22 edges
6. `newHandler()` - 22 edges
7. `writeError()` - 21 edges
8. `mapServiceError()` - 21 edges
9. `newRequest()` - 21 edges
10. `MarketShareRepo` - 20 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewMarketShareRepo()`  [INFERRED]
  cmd/api/main.go → internal/repository/market_share.go
- `main()` --calls--> `NewMarketShareService()`  [INFERRED]
  cmd/api/main.go → internal/service/market_share.go
- `main()` --calls--> `NewUserService()`  [INFERRED]
  cmd/api/main.go → internal/service/user.go
- `main()` --calls--> `Load()`  [INFERRED]
  cmd/api/main.go → internal/config/config.go
- `main()` --calls--> `Connect()`  [INFERRED]
  cmd/api/main.go → internal/repository/db.go

## Communities

### Community 0 - "Community 0"
Cohesion: 0.11
Nodes (15): ExtractClaims(), errorBody, Handler, User, UserRole, mapServiceError(), userIDFromCtx(), parsePeriod() (+7 more)

### Community 1 - "Community 1"
Cohesion: 0.07
Nodes (32): Close(), Connect(), cellValue(), isMonthHeader(), openFile(), ParseDelhiMaster(), ParsePerformance(), makeDelhiMasterXLSX() (+24 more)

### Community 2 - "Community 2"
Cohesion: 0.08
Nodes (36): Auth(), writeUnauth(), NewCompetitionService(), TestCompetitionService_GetDealerScorecard_Forbidden(), TestCompetitionService_GetDealerScorecard_Success(), TestCompetitionService_GetLeaderboard_FilterAdhoc(), TestCompetitionService_GetLeaderboard_NoActivePeriod(), Config (+28 more)

### Community 3 - "Community 3"
Cohesion: 0.07
Nodes (28): NewAuditRepo(), TestAuditRepo_Log(), TestAuditRepo_Log_EmptyCC(), limiterEntry, RateLimiter, canAccess(), NewPerformanceRepo(), TestPerformanceRepo_GetByPeriod() (+20 more)

### Community 4 - "Community 4"
Cohesion: 0.12
Nodes (29): TestLoginHandler_MissingFields(), TestLoginHandler_Success(), TestLoginHandler_WrongCredentials(), scoreValue(), TestGetDealerScorecardHandler_Forbidden(), TestGetDealerScorecardHandler_InvalidCompetitionID(), TestGetDealerScorecardHandler_Success(), TestGetLeaderboardHandler_MissingTerritory() (+21 more)

### Community 5 - "Community 5"
Cohesion: 0.1
Nodes (10): computeMSGainScore(), NewMarketShareService(), normalizeOMC(), normalizeString(), parsePeriodString(), parseVolume(), parseVolumePair(), splitMonthYear() (+2 more)

### Community 6 - "Community 6"
Cohesion: 0.12
Nodes (17): NewCompetitionRepo(), scanAuditScore(), scanBonus(), scanPeriod(), scanScore(), TestCompetitionRepo_GetActivePeriod(), TestCompetitionRepo_GetActivePeriod_NotFound(), TestCompetitionRepo_GetAuditScore() (+9 more)

### Community 7 - "Community 7"
Cohesion: 0.11
Nodes (14): UploadRepo, mockAuditRepo, mockUploadRepo, mockUserRepo, NewUploadRepo(), scanUpload(), TestUploadRepo_CreateGetUpdateDelete(), TestUploadRepo_GetByID_NotFound() (+6 more)

### Community 8 - "Community 8"
Cohesion: 0.09
Nodes (7): mockAuthSvc, mockCompSvc, mockMarketShareSvc, mockOutletSvc, mockPerfSvc, mockTargetSvc, mockUploadSvc

### Community 9 - "Community 9"
Cohesion: 0.24
Nodes (13): NewAuthService(), makeUser(), newAuthSvc(), TestAuthService_Login_InactiveUser(), TestAuthService_Login_Success(), TestAuthService_Login_UserNotFound(), TestAuthService_Login_WrongPassword(), TestAuthService_ValidateToken_Invalid() (+5 more)

### Community 10 - "Community 10"
Cohesion: 0.24
Nodes (8): UploadListResponse, UploadResponse, UploadService, NewUploadService(), parseFloat(), TestUploadService_GetHistory(), TestUploadService_GetHistory_WithCC(), TestUploadService_ProcessAsync_BadFile()

### Community 11 - "Community 11"
Cohesion: 0.33
Nodes (6): NewOutletRepo(), TestOutletRepo_GetByCC(), TestOutletRepo_GetByCC_NotFound(), TestOutletRepo_ListAll(), TestOutletRepo_ListByTerritory(), mockOutletRepo

### Community 12 - "Community 12"
Cohesion: 0.28
Nodes (2): UserRepo, scanUser()

### Community 13 - "Community 13"
Cohesion: 0.25
Nodes (7): AuthServiceI, CompetitionServiceI, OutletServiceI, PerformanceServiceI, TargetServiceI, UploadServiceI, UserServiceI

### Community 14 - "Community 14"
Cohesion: 0.25
Nodes (7): AuditRepository, CompetitionRepository, OutletRepository, PerformanceRepository, TargetRepository, UploadRepository, UserRepository

### Community 15 - "Community 15"
Cohesion: 0.29
Nodes (6): CreateUserInput, ListUsersParams, ListUsersResponse, ResetPasswordInput, UpdateUserInput, NewUserService()

### Community 16 - "Community 16"
Cohesion: 0.29
Nodes (6): MarketShareData, MarketShareStatusResponse, RecomputeResponse, TopScore, TradingArea, TradingAreaTotal

### Community 17 - "Community 17"
Cohesion: 0.33
Nodes (5): DistrictSummary, RetailOutlet, TerritoryOutlet, TerritoryStats, TerritorySummary

### Community 18 - "Community 18"
Cohesion: 0.33
Nodes (5): KPIData, PerformanceRecord, PerformanceSummary, TrendResponse, TrendRow

### Community 19 - "Community 19"
Cohesion: 0.4
Nodes (4): CompetitionBonus, CompetitionPeriod, CompetitionScore, DealerAuditScore

### Community 20 - "Community 20"
Cohesion: 0.67
Nodes (2): Product, ProductCategory

### Community 21 - "Community 21"
Cohesion: 0.67
Nodes (2): Target, TargetsResponse

### Community 22 - "Community 22"
Cohesion: 1.0
Nodes (1): loginRequest

### Community 23 - "Community 23"
Cohesion: 1.0
Nodes (1): MarketShareServiceI

### Community 24 - "Community 24"
Cohesion: 1.0
Nodes (1): UploadedFile

### Community 25 - "Community 25"
Cohesion: 1.0
Nodes (0): 

### Community 26 - "Community 26"
Cohesion: 1.0
Nodes (0): 

### Community 27 - "Community 27"
Cohesion: 1.0
Nodes (0): 

### Community 28 - "Community 28"
Cohesion: 1.0
Nodes (0): 

## Knowledge Gaps
- **69 isolated node(s):** `loginRequest`, `AuthServiceI`, `OutletServiceI`, `PerformanceServiceI`, `TargetServiceI` (+64 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 22`** (2 nodes): `loginRequest`, `auth.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 23`** (2 nodes): `MarketShareServiceI`, `market_share.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 24`** (2 nodes): `upload.go`, `UploadedFile`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 25`** (1 nodes): `me.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 26`** (1 nodes): `competition.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 27`** (1 nodes): `target.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 28`** (1 nodes): `errors.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Community 2` to `Community 1`, `Community 3`, `Community 5`, `Community 6`, `Community 7`, `Community 9`, `Community 10`, `Community 11`, `Community 15`?**
  _High betweenness centrality (0.222) - this node is a cross-community bridge._
- **Why does `New()` connect `Community 2` to `Community 0`, `Community 3`, `Community 4`, `Community 7`, `Community 10`?**
  _High betweenness centrality (0.197) - this node is a cross-community bridge._
- **Why does `Close()` connect `Community 1` to `Community 0`, `Community 2`, `Community 5`, `Community 6`, `Community 7`, `Community 10`?**
  _High betweenness centrality (0.197) - this node is a cross-community bridge._
- **Are the 38 inferred relationships involving `New()` (e.g. with `main()` and `TestGetLeaderboardHandler_Success()`) actually correct?**
  _`New()` has 38 INFERRED edges - model-reasoned connections that need verification._
- **Are the 28 inferred relationships involving `Close()` (e.g. with `main()` and `.HandleUpload()`) actually correct?**
  _`Close()` has 28 INFERRED edges - model-reasoned connections that need verification._
- **Are the 23 inferred relationships involving `writeJSON()` (e.g. with `.HandleUpload()` and `.GetUploadHistory()`) actually correct?**
  _`writeJSON()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **Are the 21 inferred relationships involving `main()` (e.g. with `Load()` and `.Validate()`) actually correct?**
  _`main()` has 21 INFERRED edges - model-reasoned connections that need verification._