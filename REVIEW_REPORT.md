# GreenLedger Verification Report

## Status: ✅ Workable (with fixes)

The repository is now confirmed to be workable in a local development environment using SQLite.

### Summary of Changes

To enable the repository to run locally as described in the documentation, the following changes were made:

1.  **SQLite Support Implementation**:
    *   Updated `shared/database/postgres.go` to support the SQLite driver (`github.com/glebarez/sqlite`) when `DB_TYPE=sqlite` is configured.
    *   Updated `shared/config/config.go` to load `DB_TYPE` and `DB_PATH` from environment variables.

2.  **Environment Configuration**:
    *   Updated `scripts/dev-setup.sh` to fix broken tool installation (`air`) and generate correct run scripts.
    *   Run scripts (`scripts/run/*.sh`) now correctly export `DB_TYPE=sqlite` and `DB_PATH`.

3.  **Cross-Database Model Compatibility**:
    *   Refactored GORM models across all services (`calculator`, `tracker`, `wallet`, `user-auth`, `certifier`, `reporting`) to remove PostgreSQL-specific SQL syntax.
    *   Removed `type:uuid` and `default:gen_random_uuid()` (relying on existing Go-side ID generation).
    *   Replaced `type:jsonb` with `type:text` (with `serializer:json` for maps) to support JSON storage in SQLite.
    *   Replaced `default:now()` with `default:CURRENT_TIMESTAMP`.
    *   Replaced `type:decimal(x,y)` with `type:numeric` for broader compatibility.

4.  **Feature Completion**:
    *   Implemented the missing `GET /api/v1/calculator/emission-factors` endpoint in the Calculator service to verify data access.

### Verification Results

*   **Build & Test:** All services build successfully. Unit tests pass.
*   **Startup:** All 4 core services (User Auth, Calculator, Tracker, Wallet) start successfully on ports 8081-8084.
*   **Database:** Services successfully connect to local SQLite databases in `data/*.db` and run auto-migrations.
*   **Functional Flow:**
    *   ✅ **User Registration:** Created new user successfully.
    *   ✅ **Authentication:** Login works and returns valid JWT.
    *   ✅ **Wallet:** User balance retrieves correctly (initialized to 0).
    *   ✅ **Calculator:** Emission factors seeded and retrievable via API.

### Recommendations

*   **Documentation:** The `README.md` and `LOCAL_DEVELOPMENT.md` are now accurate reflections of the codebase's capabilities.
*   **Testing:** Future additions should ensure models remain cross-compatible if SQLite support is to be maintained.
*   **Production:** When deploying to production with PostgreSQL, the current `type:text` changes for JSON fields are compatible, though `jsonb` would offer better performance. Conditional build tags could be considered if performance becomes critical.
