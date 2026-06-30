# Smart Shopper Agent – GitHub Copilot Instructions

## Project Overview
A multi-agent Go backend + Expo (React Native) cross-platform frontend for intelligent shopping route optimization. Backend exposes a REST API consumed by Web, iOS, and Android clients.

---

## Backend (Go)

### Language & Runtime
- **Go 1.26+** is required. Use only standard library features available in Go 1.26+.
- Module path: `smart-shopper-agent` (see `go.mod`).

### Architecture
- `cmd/app/main.go` – entrypoint, wires dependencies, starts HTTP server on port 8080.
- `internal/agents/` – AI pipeline: Parser → Pricer → Optimizer.
- `internal/mcp/` – MCP tools: PriceScraper (JSON DB), RoutePlanner (OSRM API).
- `internal/api/` – HTTP handlers, middleware (rate limiter, CORS, security headers).
- `internal/models/` – shared data structs (ShoppingList, ShoppingItem, RoutePlan).
- `internal/data/prices.json` – product price & shop coordinate database.
- `internal/automation/` – n8n workflow blueprints.

### Coding Standards
- **Structured logging only**: use `log/slog` with a JSON handler. Never use `fmt.Print*` or the `log` package for application logs.
- **Error handling**: every public function must return `error`. Use `fmt.Errorf("context: %w", err)` for wrapping. Never swallow errors silently.
- **Retry logic**: external HTTP calls (Gemini API, OSRM) must implement exponential-backoff retry (max 2 retries) using the pattern established in `internal/agents/parser.go`.
- **Timeouts**: all outgoing HTTP clients must set a timeout (default 10 s). Never use `http.DefaultClient` for external calls.
- **Clean architecture**: agents must not import MCP packages directly – dependency injection via constructor parameters only.
- **No hardcoded coordinates or secrets**: read all config from `.env` / environment variables via `godotenv`.

### Testing
- **Minimum 70% coverage** on `internal/agents/` and `internal/api/` packages.
- Every new handler must have a corresponding `*_test.go` table-driven test covering happy path, bad input, and auth failure.
- Run tests with: `go test -short ./... -v`
- Use `httptest.NewRecorder()` for handler tests; mock external HTTP calls with `httptest.NewServer()`.

### API Conventions
- Base path: `/api/v1/`
- All error responses: `{"error": "<message>"}` JSON, correct HTTP status code.
- Admin endpoints protected by `X-Admin-Token` header (value from `ADMIN_TOKEN` env var).
- Swagger annotations on every handler; regenerate docs with `swag init -g cmd/app/main.go`.

---

## Frontend (Expo / React Native)

### Platform Targets – MANDATORY
The app **must** work on **Web, iOS, and Android**. Every component, hook, and screen must be verified (or at minimum not break) on all three platforms.

- Web bundler: **Metro** (`"web": { "bundler": "metro" }` in `app.json`).
- Required web packages: `react-dom`, `react-native-web`, `@expo/metro-runtime`.
- Use `Platform.OS` guards only when a feature is genuinely unavailable on a platform; prefer cross-platform abstractions.
- Maps: `react-native-maps` renders on iOS/Android. For web, wrap with a `Platform.OS === 'web'` check and render a placeholder or a web-compatible map (e.g. Leaflet via `<WebView>`).

### Language & Tooling
- **TypeScript strict mode** (`"strict": true` in `tsconfig.json`). No `any` types without explicit justification comment.
- Expo SDK **56** (`expo: ~56.0.x`). Always check [https://docs.expo.dev/versions/v56.0.0/](https://docs.expo.dev/versions/v56.0.0/) before using any Expo API.
- Formatter: Prettier (default Expo config). Linter: ESLint with `expo` preset.

### Architecture
- `mobile/src/screens/` – full-page screen components (one per route).
- `mobile/src/components/` – reusable UI components.
- `mobile/src/hooks/` – custom hooks (business logic separated from UI). All API calls and GPS logic live here.
- `mobile/src/services/api.ts` – typed API client; all `fetch` calls go here, with full TypeScript interfaces for request/response/error shapes.

### Coding Standards
- **No inline styles** for repeated patterns; use `StyleSheet.create`.
- **Typed props**: every component must have an explicit `Props` interface.
- **Async storage**: use `@react-native-async-storage/async-storage` for offline caching; always handle `null` return (cold start / no cache).
- **Error boundaries**: wrap screen-level components; show user-friendly fallback UI on render error.
- **Loading states**: every async operation must update a `isLoading` boolean; disable interactive controls while loading.
- **Fallback GPS**: if `expo-location` permission is denied or unavailable, fall back to safe default coordinates (Budapest center).

### Testing
- Test framework: `jest` + `jest-expo` + `@testing-library/react-native`.
- Every screen must have a smoke-render test.
- Every custom hook must have unit tests covering success and error paths.
- Run tests with: `cd mobile && npm test`

### Monetization (Future – Phase 22+)
- Design components to support **Ad banners** (bottom of screen) and **Pro subscription** paywalls.
- Use `expo-ads-admob` (or equivalent) for ads; wrap in a `<AdBanner>` component that renders `null` for Pro users.
- In-App Purchases via `expo-in-app-purchases` (iOS) / `react-native-iap` (Android/Web); abstract behind a `usePurchase` hook.
- Feature flags: gate Pro features with a `useProStatus` hook reading from AsyncStorage + backend validation.

---

## General Rules

### Git
- Branch naming: `feat/<short-description>`, `fix/<short-description>`, `chore/<short-description>`.
- Commit messages: Conventional Commits format (`feat:`, `fix:`, `chore:`, `docs:`, `test:`).
- Always include Co-authored-by trailer for Copilot commits.
- Never commit secrets, `.env` files, or build artifacts.

### CI/CD (`.github/workflows/backend-ci.yml`)
- Pipeline runs on every push/PR to `main`.
- Steps: Go lint → Go test → Docker build.
- A failing pipeline blocks merges.

### Docker
- Multi-stage `Dockerfile`: `golang:1.26-alpine` builder → `alpine:latest` runtime.
- `docker-compose.yml` mounts `prices.json` as a volume for live updates via n8n.
- Build & start: `docker compose up --build -d`

### n8n Automation
- Price update workflow in `internal/automation/n8n_price_updater_workflow.json`.
- Runs daily at 02:00 via cron trigger.
- Pushes updated prices to `/api/v1/admin/prices` with `X-Admin-Token`.
