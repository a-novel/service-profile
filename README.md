# Service Profile

The backend of Agora's public profiles.

[![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/agorastoryverse)](https://twitter.com/agorastoryverse)
[![Discord](https://img.shields.io/discord/1315240114691248138?logo=discord)](https://discord.gg/rp4Qr8cA)

<hr />

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/a-novel/service-profile)
![GitHub repo file or directory count](https://img.shields.io/github/directory-file-count/a-novel/service-profile)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/a-novel/service-profile)
[![License: RPL-1.5](https://img.shields.io/badge/license-RPL--1.5-blue)](LICENSE)

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/a-novel/service-profile/main.yaml)
[![codecov](https://codecov.io/gh/a-novel/service-profile/graph/badge.svg)](https://codecov.io/gh/a-novel/service-profile)

![Coverage graph](https://codecov.io/gh/a-novel/service-profile/graphs/sunburst.svg)

## What it does

The service stores each user's public profile: a username and a square picture, WebP or JPEG. Any caller holding an access token reads a profile by its ID, an anonymous token included, and only its owner writes it. No response names the user behind a profile. It serves one public REST API (`cmd/rest`), for any HTTP client, and checks each caller's access token with [JSON Keys](https://github.com/a-novel/service-json-keys).

## Deploying

The service runs as published OCI images plus a PostgreSQL database. The server is stateless and scales to as many replicas as you need behind a load balancer. All state lives in Postgres, pictures included.

Deploy the images with any container orchestrator. The composition below shows which images to run, how they wire together, and the environment they expect.

| Image                             | Role                                                                        |
| --------------------------------- | --------------------------------------------------------------------------- |
| `service-profile/rest`            | Public REST API.                                                            |
| `service-profile/jobs/migrations` | One-shot schema migration job; runs to completion before the server starts. |
| `service-profile/database`        | PostgreSQL with pgBackRest, or bring your own Postgres.                     |

Pin every image to the same release tag. A production deployment runs `database`, then `migrations` to completion, then any number of `rest` replicas. The server reaches a JSON Keys gRPC server, deployed from that service's own images. Provide `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DATABASE` through the orchestrator; store the password as a secret.

<!-- TODO(project-docs): replace v0.0.0 with the new service's release tag -->

```yaml
services:
  postgres-profile:
    image: ghcr.io/a-novel/service-profile/database:v1.1.0
    networks: [api]
    environment:
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD}"
      POSTGRES_USER: "${POSTGRES_USER}"
      POSTGRES_DB: "${POSTGRES_DATABASE}"
      POSTGRES_HOST_AUTH_METHOD: scram-sha-256
      POSTGRES_INITDB_ARGS: --auth=scram-sha-256
    volumes:
      - profile-postgres-data:/var/lib/postgresql/
    # Podman ignores the image's own check, so the dependents' wait needs this one.
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: 1s
      timeout: 5s
      retries: 10

  migrations-profile:
    image: ghcr.io/a-novel/service-profile/jobs/migrations:v1.1.0
    depends_on:
      postgres-profile: { condition: service_healthy }
    environment:
      POSTGRES_HOST: postgres-profile
      POSTGRES_PORT: "5432"
      POSTGRES_USER: "${POSTGRES_USER}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD}"
      POSTGRES_DATABASE: "${POSTGRES_DATABASE}"
      POSTGRES_TLS_ENABLED: "false"
    networks: [api]

  service-profile:
    image: ghcr.io/a-novel/service-profile/rest:v1.1.0
    ports: ["${REST_PORT}:8080"] # the container always listens on 8080
    depends_on:
      postgres-profile: { condition: service_healthy }
      migrations-profile: { condition: service_completed_successfully }
    environment:
      POSTGRES_HOST: postgres-profile
      POSTGRES_PORT: "5432"
      POSTGRES_USER: "${POSTGRES_USER}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD}"
      POSTGRES_DATABASE: "${POSTGRES_DATABASE}"
      POSTGRES_TLS_ENABLED: "false"
      SERVICE_JSON_KEYS_HOST: service-json-keys
      SERVICE_JSON_KEYS_PORT: "8080"
    networks: [api]

networks:
  api:

volumes:
  profile-postgres-data:
```

### Database image

The database image packages PostgreSQL and pgBackRest on Wolfi. The container build assembles its [package manifest](./builds/database.apko.yaml) with apko, from [locks](./builds/database.apko.lock.json) that pin every package, so an unchanged image rebuilds to the same digest. Building it takes Docker or Podman alone. The image keeps PostgreSQL's standard `POSTGRES_*` initialization variables, the `uuid-ossp` extension and the volume at `/var/lib/postgresql`. Backup scheduling and repository credentials stay with the operator.

### Configuration

Every variable is read from the process environment. `SERVICE_PROFILE_ENV_PREFIX` prepends a prefix to every name, so the service can share an environment without collisions.

| Name                     | Description                                                                                                                   | Images                    |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| `POSTGRES_HOST`          | PostgreSQL hostname or IP address. Selects the discrete connection fields. **Required for new deployments.**                  | all                       |
| `POSTGRES_PORT`          | PostgreSQL port. Defaults to `5432`.                                                                                          | all                       |
| `POSTGRES_USER`          | PostgreSQL login role. **Required when `POSTGRES_HOST` is set.**                                                              | all                       |
| `POSTGRES_PASSWORD`      | PostgreSQL login password. **Required when `POSTGRES_HOST` is set; inject it as a secret.**                                   | all                       |
| `POSTGRES_DATABASE`      | PostgreSQL database name. **Required when `POSTGRES_HOST` is set.**                                                           | all                       |
| `POSTGRES_TLS_ENABLED`   | Encrypt the PostgreSQL connection. Defaults to `true`; disable only when another trusted boundary protects the database link. | all                       |
| `POSTGRES_DSN`           | Connection-URL fallback, read only when `POSTGRES_HOST` is empty.                                                             | all                       |
| `SERVICE_JSON_KEYS_HOST` | Hostname of the JSON Keys gRPC server, which verifies access tokens. Defaults to `localhost`.                                 | `rest`, `standalone-rest` |
| `SERVICE_JSON_KEYS_PORT` | Port of the JSON Keys gRPC server. Defaults to `8080`.                                                                        | `rest`, `standalone-rest` |

<details>
<summary>Optional configuration (server tuning, service tokens, OpenTelemetry)</summary>

REST tuning (images `rest`, `standalone-rest`):

| Name                          | Description                                                                                                                                                                                                                                       | Default                          |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| `REST_MAX_REQUEST_SIZE`       | Maximum request body size, in bytes, on every route. The default fits the largest picture the service stores.                                                                                                                                     | `262144` (256KiB)                |
| `REST_TLS_CERT_FILE`          | PEM certificate chain the server presents, leaf first. With `REST_TLS_KEY_FILE`, the server serves TLS on its port, and loads the certificate again at the first connection after the file changes, as when cert-manager renews a mounted Secret. |                                  |
| `REST_TLS_KEY_FILE`           | PEM private key of that certificate.                                                                                                                                                                                                              |                                  |
| `REST_CLIENT_IP_HEADER`       | Header a trusted proxy sets to the visitor IP. Set it only when every request comes through that proxy.                                                                                                                                           | TCP peer                         |
| `REST_LIVENESS_PATH`          | Ping path, which still answers during a planned downtime and skips the rate limit.                                                                                                                                                                | `/v1/ping`                       |
| `REST_RATE_LIMITS_ENABLED`    | Rate-limit each visitor on every route but the ping. Turn off only in test stacks.                                                                                                                                                                | `true`                           |
| `REST_RATE_LIMIT`             | Requests each visitor sends to each route, as `<requests>/<window>`.                                                                                                                                                                              | `5/1s`                           |
| `REST_TIMEOUT_READ`           | Read timeout.                                                                                                                                                                                                                                     | `15s`                            |
| `REST_TIMEOUT_READ_HEADER`    | Header read timeout.                                                                                                                                                                                                                              | `3s`                             |
| `REST_TIMEOUT_WRITE`          | Write timeout.                                                                                                                                                                                                                                    | `30s`                            |
| `REST_TIMEOUT_IDLE`           | Idle keep-alive timeout.                                                                                                                                                                                                                          | `60s`                            |
| `REST_TIMEOUT_REQUEST`        | Per-request timeout.                                                                                                                                                                                                                              | `60s`                            |
| `REST_TIMEOUT_SHUTDOWN`       | Graceful shutdown budget. With the 2-second telemetry flush after it, it fits Kubernetes' 30 seconds before `SIGKILL`.                                                                                                                            | `6s`                             |
| `REST_CORS_ALLOWED_ORIGINS`   | CORS allowed origins.                                                                                                                                                                                                                             | `*`                              |
| `REST_CORS_ALLOWED_HEADERS`   | CORS allowed headers.                                                                                                                                                                                                                             | `*`                              |
| `REST_CORS_ALLOWED_METHODS`   | CORS allowed methods, comma-separated.                                                                                                                                                                                                            | `HEAD,GET,POST,PUT,PATCH,DELETE` |
| `REST_CORS_ALLOW_CREDENTIALS` | CORS allow-credentials flag.                                                                                                                                                                                                                      | `false`                          |
| `REST_CORS_MAX_AGE`           | CORS max-age, in seconds.                                                                                                                                                                                                                         | `3600`                           |

Service tokens (images `rest`, `standalone-rest`). Without `SERVICE_TOKEN_ISSUER`, the server refuses to start, unless `SERVICE_ISSUER_OPTIONAL` lets it take every request: the `standalone-rest` image sets it for local, test and consumer stacks, and production never does. With it, the server refuses a request unless its `X-Service-Authorization` header carries a token from an allowed caller. The check runs before the server reads `REST_CLIENT_IP_HEADER` or counts a rate limit, so only an allowed caller names the visitor IP. The ping stays open for probes. Each service the server calls has its own audience, so each connection sends its own token:

| Name                           | Description                                                                                                                                                                                                                                                                 | Default |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `SERVICE_TOKEN_ISSUER`         | Issuer of the service tokens the server accepts, such as a GKE cluster's `https://container.googleapis.com/v1/projects/<project>/locations/<location>/clusters/<cluster>`. The server reads its keys through its OpenID discovery document, so it needs egress to its host. |         |
| `SERVICE_TOKEN_AUDIENCE`       | Audience the server answers to. Each caller's projected token names it.                                                                                                                                                                                                     |         |
| `SERVICE_TOKEN_CALLERS`        | Service accounts allowed to call, comma-separated, each as `system:serviceaccount:<namespace>:<name>`.                                                                                                                                                                      |         |
| `SERVICE_ISSUER_OPTIONAL`      | Set to `true` to start without `SERVICE_TOKEN_ISSUER` and take every request. The `standalone-rest` image sets it; never set it in production.                                                                                                                              | `false` |
| `SERVICE_JSON_KEYS_TOKEN_PATH` | File the kubelet projects the service's token for the JSON Keys audience into. When set, the JSON Keys connection sends that token with every call, over TLS once `SERVICE_JSON_KEYS_TLS` is set.                                                                           |         |
| `SERVICE_JSON_KEYS_TLS`        | Dial JSON Keys over TLS once `SERVICE_JSON_KEYS_TOKEN_PATH` is set, verifying its certificate for `SERVICE_JSON_KEYS_HOST` against the system roots, which `SSL_CERT_DIR` extends.                                                                                          | `false` |

Database connection pool (server images). The limits are **per process**. The database's `max_connections` has to cover every replica plus the migration job; the stock `postgres` default is 100.

| Name                      | Description                               | Default |
| ------------------------- | ----------------------------------------- | ------- |
| `POSTGRES_MAX_OPEN_CONNS` | Maximum open connections to the database. | `20`    |
| `POSTGRES_MAX_IDLE_CONNS` | Maximum connections kept open while idle. | `20`    |

Planned downtime (server images):

| Name             | Description                                                                                                                                                                             | Default |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `DOWNTIME_START` | When a planned downtime starts, in RFC 3339. From then until it is removed, even past the announced end, the server answers `503` except for the ping, and starts without the database. |         |

Logs and tracing: OpenTelemetry exports to standard output, or over OTLP/HTTP once an endpoint is set (all server images):

| Name                             | Description                                                                                                      | Default           |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------------- |
| `OTEL`                           | Enable OTel tracing; the variables below pick the exporter.                                                      | `false`           |
| `OTEL_EXPORTER_OTLP_ENDPOINT`    | Base URL of the OTLP/HTTP backend. When set, traces and logs export there, and the servers write no request log. |                   |
| `OTEL_EXPORTER_OTLP_HEADERS`     | Headers each export sends, such as the backend's key. Inject it as a secret.                                     |                   |
| `OTEL_EXPORTER_OTLP_COMPRESSION` | `gzip` compresses each export.                                                                                   |                   |
| `OTEL_SERVICE_NAME`              | Service name every span and log record carries.                                                                  |                   |
| `APP_NAME`                       | Application name attached to traces and logs.                                                                    | `service-profile` |

</details>

## Using the client packages

The JavaScript / TypeScript client talks REST, for frontends and Node services. The snippet is the **minimum viable call**. Your editor's intellisense and the [API reference](https://a-novel.github.io/service-profile) cover the full surface.

### JavaScript / TypeScript (REST)

The package is published to GitHub Packages, which requires a Personal Access Token with the `read:packages` scope even for public packages ([why](https://github.com/orgs/community/discussions/23386#discussioncomment-3240193)). Add to `.npmrc` (project root or `$HOME`):

```ini
@a-novel:registry=https://npm.pkg.github.com
@a-novel-kit:registry=https://npm.pkg.github.com
//npm.pkg.github.com/:_authToken=${YOUR_PERSONAL_ACCESS_TOKEN}
```

```bash
pnpm add @a-novel/service-profile-rest
```

```typescript
import { ProfileApi } from "@a-novel/service-profile-rest";

const api = new ProfileApi("http://service-profile:8080");

await api.ping();
```

Every resource helper takes an access token from the authentication service. A read accepts any token, an anonymous one included; a write acts for the signed-in user:

```typescript
import { profileGet, profileUsernameUpdate } from "@a-novel/service-profile-rest";

const mine = await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });
const profile = await profileGet(api, anonymousToken, mine.id);
```

## Running locally

The **standalone** images bundle a server with the migrations, which they apply on every boot. Use them for a throwaway instance without the dev toolchain, never for a deployment with several replicas. Run a JSON Keys server beside it, as its own README shows.

```yaml
services:
  postgres-profile:
    image: ghcr.io/a-novel/service-profile/database:v1.1.0
    networks: [api]
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_USER: postgres
      POSTGRES_DB: postgres
      POSTGRES_HOST_AUTH_METHOD: scram-sha-256
      POSTGRES_INITDB_ARGS: --auth=scram-sha-256
    # Podman ignores the image's own check, so the dependents' wait needs this one.
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: 1s
      timeout: 5s
      retries: 10

  service-profile:
    image: ghcr.io/a-novel/service-profile/standalone-rest:v1.1.0
    ports: ["${REST_PORT}:8080"]
    depends_on:
      postgres-profile: { condition: service_healthy }
    environment:
      POSTGRES_DSN: "postgres://postgres:postgres@postgres-profile:5432/postgres?sslmode=disable"
      SERVICE_JSON_KEYS_HOST: service-json-keys
    networks: [api]

networks:
  api:
```

## Contributing

Platform setup and the day-to-day commands live in the [developer guide](https://a-novel-kit.github.io/stack/). This repository's own concepts and local commands are in [CONTRIBUTING.md](./CONTRIBUTING.md).
