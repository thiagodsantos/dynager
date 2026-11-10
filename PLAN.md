# DynamoDB Manager — Implementation Plan

## 1. Goal

Build a modern, responsive web application for configuring DynamoDB connections, browsing tables and items, filtering records, and editing or deleting records. The frontend and backend live in one monorepo and are developed, tested, and operated through a root `Makefile`.

## 2. Scope and key decisions

### In scope

- Create, test, list, and manage connection configurations.
- Browse a connection's tables and inspect table key/schema details.
- Browse records with pagination and filtering.
- Edit and delete records while respecting each table's partition and optional sort key.
- Light and dark themes, responsive layouts, and clear loading, empty, and error states.
- Unit and integration tests, including a Docker Compose-based integration environment.
- Project documentation, application icon/image, and MIT license.

### Decisions to confirm before implementation

- **Credential and connection persistence:** decide whether connections are local-only and stored on the user's machine, or shared/persisted by a deployed service. Do not store AWS secrets in browser storage or expose them to the frontend. Prefer the AWS SDK's credential chain where possible; if credentials must be persisted, define encryption and access controls first.
- **Filtering behavior:** DynamoDB `Scan` filters are not equivalent to indexed queries and can be expensive. Start with explicit, bounded filtering and pagination; decide whether the UI supports only scans/filters or also key-based queries using table indexes.
- **Deployment and access control:** the initial requirements do not specify hosting, multi-user access, or authentication. Treat the first milestone as a trusted, local/single-user tool unless these needs are clarified; do not expose an unauthenticated service to an untrusted network.
- **Template usage:** review the TailAdmin template's license, required attribution, and compatibility with the chosen React/Tailwind versions before adopting components.

## 3. Proposed monorepo layout

```text
.
├── frontend/                # Vite + React + TypeScript application
├── backend/                 # Go REST API
├── docker/                  # Dev/test Docker Compose environments and fixtures
├── agents/                  # Project agent/skill configuration and usage notes
├── assets/                  # Application icon and related source assets
├── Makefile                 # Common install, run, test, build, and lint commands
├── README.md
├── LICENSE                  # MIT
└── PLAN.md
```

Keep the layout lean; add subpackages only when they represent a real boundary. Use frontend path aliases consistently rather than relative imports.

## 4. Architecture and implementation guidelines

### Frontend

- Use the latest stable Node.js release supported by the selected Vite/tooling versions, PNPM, Vite, React, TypeScript, and Tailwind CSS.
- Use the TailAdmin free dashboard as a visual/component reference, while keeping application-specific screens and state management maintainable.
- Organize UI with Atomic Design conventions (atoms, molecules, organisms, templates, and routed pages) without forcing trivial code into extra layers.
- Add routing for setup/configuration, connection list, and table/record browsing pages. Keep connection/table selection reflected in route parameters where useful.
- Implement reusable form controls, data table, filter controls, dialogs, notifications, and theme switching. Persist the theme preference locally.
- Keep AWS credentials and DynamoDB access in the backend; the browser communicates only with the REST API.

### Backend

- Use Go and `chi` for a versioned REST API. Require `X-API-Version: v1` for v1 routes and return a clear client error for unsupported or missing versions. Keep version selection in middleware so it is not duplicated across handlers.
- Separate HTTP transport, application/use-case logic, and DynamoDB access behind small interfaces. Apply SOLID, DRY, and YAGNI pragmatically; avoid speculative abstractions.
- Use the official AWS SDK for Go and its credential/configuration mechanisms. Return sanitized errors and never log secrets or raw credential values.
- Use DynamoDB expression builders/placeholders for keys and filters; never interpolate user-provided filter text into expressions.
- Handle pagination explicitly. Preserve key values and DynamoDB attribute types when displaying and editing items, including composite keys and nested attributes.
- Validate inputs and map expected failures to consistent HTTP status codes and error response bodies.

### API outline (v1)

All routes use `/api` and require `X-API-Version: v1`.

- `GET /api/connections` — list configured connections (never return secrets).
- `POST /api/connections` — add/update a connection configuration.
- `POST /api/connections/test` — validate credentials and connectivity without saving unless explicitly requested.
- `DELETE /api/connections/{connectionID}` — remove a configuration.
- `GET /api/connections/{connectionID}/tables` — list tables.
- `GET /api/connections/{connectionID}/tables/{tableName}/items` — retrieve a page of items with bounded filter/query parameters and a continuation token.
- `PUT /api/connections/{connectionID}/tables/{tableName}/items` — update an item, identifying it by its complete primary key.
- `DELETE /api/connections/{connectionID}/tables/{tableName}/items` — delete an item by its complete primary key.

Finalize request/response schemas and the connection identifier/storage model before implementing the handlers. Do not accept arbitrary AWS operations through a generic proxy endpoint.

## 5. Delivery phases

### Phase 0 — Decisions and foundation

- Resolve connection persistence/credential handling, supported connection modes (for example AWS region/profile or explicit credentials), and intended deployment boundary.
- Check TailAdmin licensing and select compatible dependency versions.
- Create the monorepo directories, root Makefile, formatting/lint conventions, and environment-variable examples containing no real secrets.
- Document local prerequisites and the supported Node.js version.

**Exit criteria:** repository structure and security boundaries are documented; `make help` (or equivalent) lists the agreed commands.

### Phase 1 — Backend API foundation

- Initialize the Go module and `chi` server with configuration, graceful shutdown, health check, CORS policy, request logging, and consistent JSON errors.
- Add API-version middleware for `X-API-Version: v1`.
- Define connection configuration and DynamoDB client interfaces; implement connection test and table listing.
- Add unit tests for middleware, validation, and error mapping; use mocked interfaces for handler/use-case tests.

**Exit criteria:** API starts locally, rejects unsupported API versions, and connection test/table-list behavior is covered by tests.

### Phase 2 — Frontend shell and connection management

- Scaffold the Vite/React/TypeScript/Tailwind app and configure path aliases, routing, formatting, and linting.
- Apply the TailAdmin-inspired layout with responsive navigation, theme toggle, and light/dark design tokens.
- Build setup/configuration and connection-list pages, including validation, test-connection feedback, and safe secret-field handling.
- Connect the UI to the backend through a typed API client that consistently sends the version header.

**Exit criteria:** users can add/configure a connection, test it, view configured connections, and see actionable loading/error/success states.

### Phase 3 — Tables and record operations

- Add table selection and table metadata/key-schema display.
- Implement record pagination and supported filters, with clear scan/query behavior and warnings for potentially expensive scans.
- Build a detail/edit form that preserves DynamoDB values and requires the full primary key; add a delete confirmation flow.
- Handle empty tables, invalid filters, permission errors, expired/invalid continuation tokens, and concurrent update failures.

**Exit criteria:** users can safely browse, filter, edit, and delete records across tables with partition-only and composite keys.

### Phase 4 — Integration tests and local environment

- Add Docker Compose configuration for DynamoDB Local and any required test services.
- Provide deterministic test fixtures for table creation, key schemas, and representative item types.
- Write backend integration tests against DynamoDB Local; keep unit tests independent of Docker.
- Add a Makefile target to start dependencies, run integration tests, and clean up reliably.

**Exit criteria:** a documented command runs the integration suite from a clean checkout and verifies connection, table, pagination/filter, update, and delete flows.

### Phase 5 — Polish, docs, and release readiness

- Create an original application icon/image and wire it into the frontend metadata.
- Complete `README.md` with purpose, features, tech stack, architecture, prerequisites, configuration, local use, tests, build, and security notes.
- Add the MIT `LICENSE`; include any required attribution for third-party template assets.
- Verify responsive behavior, accessibility basics (keyboard use, labels, focus states, contrast), and light/dark themes.
- Run formatting, lint, unit tests, integration tests, and production builds through Makefile targets.

**Exit criteria:** a new contributor can install, run, test, and build the monorepo by following the README and Makefile.

## 6. Testing strategy

- **Frontend:** unit/component tests for forms, API error handling, filter controls, and edit/delete interactions; end-to-end tests for the main connection-to-record workflow if an E2E framework is selected.
- **Backend unit tests:** API version middleware, validation, use cases, pagination token handling, and HTTP response mapping.
- **Backend integration tests:** exercise DynamoDB Local for connection/configuration flow where applicable, table listing, key handling, pagination, filtering, item update, and deletion.
- **Build checks:** frontend typecheck/lint/build and Go formatting/vet/test. Integration tests should be explicitly runnable separately from fast unit tests.
- **Makefile:** provide consistent targets such as `install`, `run`, `build`, `test`, `test-unit`, `test-integration`, `lint`, `format`, and `clean`; document whether each target needs Docker or external AWS access.

## 7. Agent and skill setup

Use technology-specialized agents with clear ownership. The coordinating agent assigns work, resolves cross-stack decisions, and ensures the frontend/backend contract remains consistent; specialized agents should not independently change another agent's owned area without coordination.

- **Frontend agent:** owns `frontend/` and the React, TypeScript, Vite, Tailwind, routing, accessibility, and browser-side test work. It must keep AWS credentials and DynamoDB access out of the browser and use the typed REST API client.
- **Backend agent:** owns `backend/` and the Go, `chi`, AWS SDK, API, connection configuration, and backend test work. It must enforce API versioning, validate inputs, sanitize errors, and never expose or log secrets.
- **Platform/DevEx agent:** owns root build and local-development tooling, including the `Makefile`, `docker/` Compose setup, and environment/configuration examples. It coordinates commands and interfaces with both application agents rather than embedding application logic in tooling.
- **Test/quality agent:** owns cross-stack test strategy and integration-test coordination, including DynamoDB Local fixtures and verification of the documented workflows. It may add tests in either stack in coordination with that stack's owner, but should not change production behavior solely to satisfy a test.

Agents should agree on API schemas and shared commands before parallel implementation, keep changes within their ownership boundaries, and report assumptions or blockers that affect another stack.

- Review the referenced Matt Pocock skills repository and select only skills useful to this project (for example, test-driven implementation, code review, and frontend/backend conventions).
- Record the source URL, selected skill versions/revisions, installation/use instructions, and any license requirements in `agents/README.md`.
- Keep project-specific agent instructions focused on repository commands, architecture boundaries, security rules, and testing expectations. Avoid installing or invoking skills blindly; review their contents before use.

## 8. Risks and safeguards

- **Credential exposure:** keep secrets out of browser bundles, API responses, logs, screenshots, and committed configuration; use secret redaction and a clear storage policy.
- **Expensive reads:** scans may consume significant capacity. Bound page sizes, expose pagination, make filtering limitations clear, and consider a confirmation for broad scans.
- **Record type/key corruption:** preserve DynamoDB attribute types and require the complete primary key for writes; test nested and composite-key records.
- **Destructive edits:** require explicit delete confirmation and report update/delete failures without implying success.
- **Network exposure:** define local binding, CORS, and authentication expectations before deployment beyond a trusted environment.
- **Third-party UI assets:** check license and attribution requirements before copying assets into the project.
