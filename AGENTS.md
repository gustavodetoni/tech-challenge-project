# Tech Challenge Project — Agent Notes

## Objective
This API manages an auto shop workflow:
- Administrative management (clients, vehicles, services, parts/stock).
- Service Orders (OS) creation and tracking.
- Budget generation, approval/rejection, and lifecycle status transitions.

## Core Principles
- **Layered boundaries**: HTTP (Gin) is an adapter; business orchestration lives in application services; DB is an adapter behind repository interfaces.
- **Dependency direction**: controllers/routes depend on application services; application services depend on repository interfaces; infra/db implements those interfaces.
- **Prefer small, explicit types** over “fat” models; keep mapping at the edge (controllers + repositories).
- **Use Postgres schema as source of truth**: tables/types live in `internal/infra/db/migrations`.

## High-Level Architecture
**Composition root**
- `cmd/api/main.go`: builds dependencies (DB, repos, services, controllers) and registers routes.

**HTTP Adapter (inbound)**
- `internal/interfaces/http/routes`: router wiring + route groups.
- `internal/interfaces/http/controllers`: request/response DTOs, validation at HTTP boundary, calls application services.
- `internal/interfaces/http/middlewares`: auth middleware (JWT).

**Application (use cases)**
- `internal/application/auth`: register/login use cases (bcrypt hashing + token issuance).
- `internal/application/admin`: admin use cases for CRUD and OS metrics.
- `internal/application/serviceorder`: OS flow (create draft, send budget, approve/reject, status transitions).

**Domain (entities/value objects)**
- `internal/domain/*`: `client`, `vehicle`, `service`, `part`, `order` (OS + budget).

**Outbound Ports (interfaces)**
- `internal/interfaces/repository/*`: repository interfaces and shared errors (`ErrNotFound`, `ErrConflict`).

**Infra / DB Adapter (outbound)**
- `internal/infra/db/db.go`: connect + `InitAndCheckMigration()` using `goose.Up`.
- `internal/infra/db/repositories/*`: Gorm/Postgres implementations of repository ports.

## Database + Migrations
- Migrations directory: `internal/infra/db/migrations`
- Startup calls `db.InitAndCheckMigration(...)` which runs `goose.Up`.
- Key tables: `users`, `clients`, `vehicles`, `services`, `parts`, `service_orders`, `budgets`, `budget_*`, `service_order_*`, `stock_movements`, `service_order_status_history`.

## Auth (JWT)
- JWT HS256 is implemented in `internal/infra/auth/jwt.go`.
- Middleware:
  - `RequireAuth()` validates Authorization.
  - `RequireRoles("ADMIN","MANAGER")` protects admin routes.
- Header format supported:
  - `Authorization: Bearer <token>` (preferred)
  - raw token (fallback) is accepted.

### Roles
Database enum `user_role`: `ADMIN`, `MANAGER`, `MECHANIC`, `ATTENDANT`, `VIEWER`.
New registered users default to `VIEWER` (not admin).

## Validation
Brazilian document/plate validation uses `github.com/brazilian-utils/go`:
- CPF/CNPJ: `cpf.IsValid`, `cnpj.IsValid`
- Plate: `licenseplate.IsValid` (input is normalized before validating).

## API Route Conventions
### Public
- `GET /health`
- `GET /swagger/index.html`
- `POST /auth/register`
- `POST /auth/login`
- Client OS progress/approval (no JWT):
  - `GET /client/service-orders/{code}?document_number=...`
  - `POST /client/service-orders/{code}/budget/approve?document_number=...`
  - `POST /client/service-orders/{code}/budget/reject?document_number=...`

### Protected (any logged in user)
- `GET /me`

### Admin (JWT + role ADMIN/MANAGER)
CRUD:
- Clients: `/admin/clients`
- Vehicles: `/admin/clients/{id}/vehicles` and `/admin/vehicles/{id}`
- Services: `/admin/services`
- Parts + stock movements: `/admin/parts` and `/admin/parts/{id}/stock-movements`

Service Orders:
- Listing/detail: `GET /admin/service-orders`, `GET /admin/service-orders/{id}`
- Flow:
  - `POST /admin/service-orders` (create OS + draft budget)
  - `POST /admin/service-orders/{id}/diagnosis/start`
  - `POST /admin/service-orders/{id}/budget/send`
  - `POST /admin/service-orders/{id}/finish`
  - `POST /admin/service-orders/{id}/deliver`

Metrics:
- `GET /admin/metrics/avg-execution-time`

## Service Order (OS) Flow
Implemented against the existing schema (no schema changes):
1. **Create Draft OS**: finds/creates client by CPF/CNPJ; finds/creates vehicle by plate; creates `service_orders` (status `RECEIVED`) + `budgets` (status `DRAFT`) + budget lines.
2. **Start Diagnosis**: status `IN_DIAGNOSIS`.
3. **Send Budget**: budget `SENT`, OS `WAITING_APPROVAL`.
4. **Client Approves**: budget `APPROVED`, OS `IN_PROGRESS`; parts stock is decremented and `stock_movements` are created (reference `SERVICE_ORDER`).
5. **Client Rejects**: budget `REJECTED`, OS returns to `IN_DIAGNOSIS`.
6. **Finish / Deliver**: `FINISHED` → `DELIVERED`.

Status transitions are validated in the DB flow repository to avoid invalid jumps.

## Swagger
- UI: `/swagger/index.html`
- Generated docs live in `docs/` and are regenerated via `swag init`.
- Comment annotations live mainly in controller methods.

## Local Development
Environment variables (see `.env.example`):
- `PORT`
- `DATABASE_URL`
- `JWT_SECRET` (required)
- `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_EXPIRY_MINUTES`

Run:
- `make run` (or `go run ./cmd/api`)
- If using Air: `air -c air.toml`

## Adding a New Feature (pattern)
1. Create/extend domain types in `internal/domain/...` (keep simple).
2. Add/extend repository port in `internal/interfaces/repository/...`.
3. Implement in `internal/infra/db/repositories/...` (Gorm mapping + soft-delete rules).
4. Add an application service method in `internal/application/...` (validation + orchestration).
5. Add controller + route wiring in `internal/interfaces/http/...` (HTTP DTOs + Swagger annotations).
6. Add unit tests at application layer when possible (prefer fake repos).

