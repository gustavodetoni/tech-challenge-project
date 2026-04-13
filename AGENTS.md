# Agent Guidelines (Repo)

## `pkg/` Policy
Use `pkg/` for **stable, reusable, non-domain utilities** that may be shared across multiple layers/binaries without importing anything from `internal/`.

Good candidates:
- Input normalization helpers (e.g., BR documents/plates).
- Generic helpers that don’t depend on Gin/Gorm/app-specific services.

Currently extracted:
- `pkg/br/document`: digits-only normalization + CPF/CNPJ validation wrappers.
- `pkg/br/plate`: plate normalization (uppercase, remove spaces/hyphens).

Avoid putting in `pkg/`:
- Domain entities (`internal/domain/...`).
- Use cases / application services (`internal/application/...`).
- Repository ports (`internal/interfaces/repository/...`).
- HTTP controllers/middlewares (Gin-specific) and DB adapters (Gorm-specific).

Rule of thumb:
- If it imports from `internal/`, it **does not belong** in `pkg/`.
- If it would lock us to Gin/Gorm, prefer keeping it under `internal/` (adapter layer).

This file is the canonical agent guide for the repo root.
