# Agent boundaries and working notes

This repository is a trusted, local, single-user DynamoDB manager. Read the root `README.md` and `make help` before making changes. Do not install third-party agent skills, download templates, or copy template assets as part of agent setup.

## Ownership

- **Frontend work:** `frontend/` — UI, browser tests, and API client. The browser must not hold AWS credentials or call DynamoDB directly.
- **Backend work:** `backend/` — Go API, AWS SDK use, connection configuration, and backend tests. Bind to loopback by default; do not expose an unauthenticated API to an untrusted network.
- **Platform/DevEx work:** root-level files plus `docker/`, `assets/`, and `agents/` — Make targets, local service configuration, original project assets, and cross-cutting documentation. Coordinate interfaces with the frontend and backend owners.
- **Integration/quality work:** coordinate with the owner of each application directory before changing its code or tests. Keep unit tests independent of Docker; run integration tests against the local DynamoDB service.
- **Coordinator:** resolve shared API/configuration decisions and assign cross-stack work explicitly. Do not edit another owner's area without coordination.

## Shared decisions

- The initial deployment boundary is a trusted local machine and one user. Bind services to `127.0.0.1`; there is no authentication layer, so do not bind to a LAN/public interface or otherwise expose the API without adding appropriate access control.
- Persist only non-secret connection settings in the backend (for example, region, endpoint, and AWS profile name). Resolve credentials through the AWS SDK's normal environment/shared-profile credential chain. Never persist AWS credentials in app-managed storage, return them from the API, or place them in browser storage.
- Record and follow the backend's eventual configuration contract; `make test-integration` sets `DYNAMODB_LOCAL_ENDPOINT` for the local test endpoint; the integration test uses disposable dummy credentials and never contacts AWS.
- Record browsing as a bounded, paginated DynamoDB `Scan`, with at most one optional equality filter. A Scan can still read many records and consume capacity; keep limits enforced server-side and communicate the limitation in the UI.
- Keep item add/edit flows in the shared item modal. Adding defaults to the typed attribute builder and also supports syntax-highlighted DynamoDB JSON; validate the complete table key before writes.
- Use the reusable confirmation dialog for destructive actions instead of `window.confirm`. Keep a single React Toastify container at the app root; use toasts for operation success/failure and inline feedback for local validation and background loading errors.
- TailAdmin is visual inspiration only. Do not copy its source, components, icons, screenshots, or other assets into this repository. The project icon in `assets/` is original.
- Keep credentials and other secrets out of `.env.example`, logs, documentation, screenshots, and commits. The checked-in example contains only non-secret local settings.

## Template and skill review

- **TailAdmin:** reviewed [TailAdmin's free template repository](https://github.com/TailAdmin/tailadmin-free-tailwind-dashboard-template) on 2026-11-10. Its README identifies the community edition as MIT-licensed. The upstream template is Alpine.js/Webpack-oriented and current docs mention Tailwind CSS v4, while this app is React/Vite with Tailwind CSS v3; it is used only as visual inspiration. No code/assets are copied, so the template's attribution/redistribution terms do not apply to project files.

## Reviewed agent skills

Reviewed the [Matt Pocock skills repository](https://github.com/mattpocock/skills) at commit `4ae2317ddb22b046e26455e4f9c9b3aaaf62212d` (2026-10-31). The repository is MIT-licensed; see its [`LICENSE`](https://github.com/mattpocock/skills/blob/4ae2317ddb22b046e26455e4f9c9b3aaaf62212d/LICENSE). Reviewed the `tdd` and `code-review` skills because they directly support this project's backend/frontend tests and cross-stack review. No third-party skill files are vendored or installed by default; use the existing project commands and tests without requiring those skills.

- `tdd` (`skills/engineering/tdd`): useful for new behavior and regression fixes; upstream's instructions use a red-green-refactor loop and include companion test/mocking guidance. Reviewed at the commit above.
- `code-review` (`skills/engineering/code-review`): useful for a final two-axis standards/spec review after implementation. Reviewed at the commit above.

To install through an agent-aware skills installer, follow the upstream instructions at the source URL and select only `tdd` and `code-review` (for example, `npx skills@latest add mattpocock/skills --skill=tdd --skill=code-review`, then choose the target agent when prompted). That command follows the current upstream revision, not the reviewed pin. For reproducible use, check out commit `4ae2317ddb22b046e26455e4f9c9b3aaaf62212d` and install the complete `skills/engineering/tdd` and `skills/engineering/code-review` directories, including referenced companion files, into the agent's documented skill directory. Review the upstream license and changed skill contents again before updating. Do not install the unrelated setup, issue-tracker, or factory-orchestration skills for this repository.

## Working conventions

- Use root Make targets (`install`, `run`, `run-backend`, `build`, `test-unit`, `test-integration`, `lint`, `format`, and `clean`) rather than inventing parallel root workflows. Frontend targets execute only scripts present in `frontend/package.json`.
- Agree on API shape, environment variable names, and package scripts with the owning stack before changing shared tooling.
- Keep changes focused; add or update frontend tests when changing item editors, confirmation flows, or toast behavior. Do not install agent skills or third-party UI templates without explicit project-owner approval and license review.
