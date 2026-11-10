# dynager — DynamoDB manager frontend

A responsive React + TypeScript workspace for exploring DynamoDB connections, tables, and typed items.

## Development

Requires Node.js 22.12+ and pnpm.

```sh
pnpm install
pnpm dev
```

Vite proxies `/api` to `http://localhost:8080` by default. Set `DYNAMODB_MANAGER_API_TARGET` to proxy to a different backend address (for example, `http://localhost:8081`). Set `VITE_API_BASE_URL` only when the API is hosted at a different origin. All API requests include `X-API-Version: v1`.

## Checks

```sh
pnpm run check
pnpm run lint
pnpm run format:check
pnpm test
pnpm run build
```

ESLint extends the Airbnb JavaScript/React style guide with TypeScript support. Use `pnpm run format` to apply Prettier formatting.

## Credential safety

The browser never asks for, stores, or displays AWS access keys or session tokens. Connections use the backend's AWS SDK credential chain or a named profile; the frontend submits only the connection name, region, optional profile, and optional endpoint. Endpoint URLs containing user information, query strings, or fragments are rejected. Theme preference is the only application state persisted by the browser.
