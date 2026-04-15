<INSTRUCTIONS>
# Agent Guidelines (Tests)

This document defines the canonical testing conventions for this repository.

Scope: everything under `internal/tests/`.

## Goals

- Keep tests fast, deterministic, and easy to run locally and in CI.
- Use the same layered architecture as production code: application → interfaces/http → infra.
- Prefer testing behavior (inputs/outputs, side effects, error mapping) over implementation details.
- Make `make tests` the only required command to run the whole test suite.

## Directory Layout (Mirror `internal/`)

The `internal/tests/` tree must mirror the `internal/` tree so tests are easy to discover:

- Application unit tests:
  - `internal/tests/application/<module>/*_test.go`
- HTTP unit tests (controllers, middlewares, shared helpers):
  - `internal/tests/interfaces/http/controllers/*_test.go`
  - `internal/tests/interfaces/http/middlewares/*_test.go`
  - `internal/tests/interfaces/http/shared/*_test.go`
- Integration tests (HTTP flows using the real Gin router):
  - `internal/tests/interfaces/http/routes/*_test.go`
- Test-only infra (in-memory repositories, fakes):
  - `internal/tests/infra/...`
- Repository interface mocks:
  - `internal/tests/interfaces/repository/mocks/*.go`

Do not place tests next to production files under `internal/`. Keep them under `internal/tests/`.

## Test Types

### Unit Tests

Unit tests validate a single unit with its dependencies mocked:

- Application services: mock repository interfaces.
- Controllers: use `httptest` + Gin router and mock the service layer.
- Middlewares: use `httptest` + Gin and issue real JWTs with the real JWT manager.

Unit tests should not talk to Postgres or the filesystem.

### Integration Tests (Main Flows)

Integration tests validate the main flows through HTTP endpoints using:

- The real Gin router (`routes.Register`)
- Real controllers and middlewares
- In-memory repositories (test-only infra) instead of DB

Integration tests should cover end-to-end flows such as:

- service-order lifecycle (create → diagnosis → revise budget → send → approve → finish → deliver)
- stock consumption on budget approval
- metrics endpoints

Integration tests must remain deterministic and not depend on wall-clock timing beyond reasonable tolerances.

## Libraries

- Standard library `testing`
- `testify`:
  - `require` for “must succeed” checks
  - `assert` for non-fatal comparisons
  - `mock` for mocks

Avoid additional libraries unless explicitly requested.

## Naming Conventions

- Test function names: `Test<Thing>_<Method>_<Scenario>` (CamelCase).
- Use table tests when a function has many small variants, otherwise keep tests focused.
- Use `t.Parallel()` by default, unless the test uses shared global state.

## Arrange / Act / Assert

Write tests in a clear AAA structure:

1) Arrange: create service/controller/router and mocks
2) Act: call method/handler
3) Assert: verify outputs and mock expectations

## Mocks (Mockery-style, committed)

We keep repository mocks committed so developers only need `make tests`.

- Mocks live under: `internal/tests/interfaces/repository/mocks/`
- They follow mockery conventions (a `mock.Mock` struct with methods calling `m.Called(...)`).
- Do not add a `make mocks` requirement.

When adding a new repository interface method used by tests, update the corresponding mock file.

### Mock Expectations

- Prefer `mock.MatchedBy(...)` for domain objects to assert key fields without tying tests to full structs.
- Always call `repo.AssertExpectations(t)` when you set expectations.
- Use `mock.Anything` for context arguments.

## HTTP Testing Guidelines

### Controllers

- Use `gin.SetMode(gin.TestMode)`.
- Build a minimal router and register only the handler(s) under test.
- Use `httptest.NewRequest` / `httptest.NewRecorder`.
- Assert HTTP status codes and response JSON.

### Middlewares

- Issue a real JWT using `internal/infra/auth.Manager` in tests.
- Set `Authorization: Bearer <token>`.
- Validate status codes for unauthorized/forbidden/success cases.

### Helpers

In integration tests, prefer a small helper to perform JSON requests:

- marshal `body` with `encoding/json`
- set `Content-Type: application/json` when needed
- set `Authorization` when needed
- unmarshal response JSON into `map[string]any` or a typed struct

## In-Memory Repositories (Test Infra)

For integration tests, prefer in-memory implementations of repository interfaces.

- Keep them under `internal/tests/infra/memory/...`
- Keep behavior consistent with the “happy path” and essential constraints:
  - uniqueness (document, plate, sku)
  - not found/conflict errors (`repository.ErrNotFound`, `repository.ErrConflict`)
  - stock consumption on budget approval
  - status transition rules consistent with production flow

Do not attempt to fully re-implement DB semantics; implement only what the flow needs.

## Running Tests

- Use `make tests` to run everything.
- The test runner uses `/tmp` for `GOCACHE`, `GOTMPDIR`, and `GOMODCACHE` to avoid sandbox/permission issues.

## Style

- Keep tests readable and minimal.
- Avoid large JSON blobs when a small body is enough to assert behavior.
- Prefer stable test data (fixed CPF, plate, etc.).
- Do not add inline comments unless they clarify a non-obvious edge case.
</INSTRUCTIONS>

