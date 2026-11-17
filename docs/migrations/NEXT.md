# Upgrading to NEXT

A request that parses but fails validation answers 422 instead of 400, and the images no longer read `POSTGRES_DSN`.

## Do I need to act?

**Required** for a client that tells a refused request apart by its status, and for a deployment that sets `POSTGRES_DSN` without `POSTGRES_HOST`. A client that handles every 4xx alike, and a deployment that sets `POSTGRES_HOST`, take this release as-is.

## Read 422 for a request that fails validation

The service now answers as the authentication service does: 400 for a request it cannot parse, 422 for one that parses but breaks a rule. The problem body keeps its shape, with `status` 422 and the title `Unprocessable Entity`.

| Request                                             | Before | After |
| --------------------------------------------------- | ------ | ----- |
| `PUT /v1/profile` with a refused username           | 400    | 422   |
| `GET /v1/profile` without an `id`                   | 400    | 422   |
| `GET /v1/profile/picture` without an `id`           | 400    | 422   |
| an `id` that is not a UUID, a body that is not JSON | 400    | 400   |

```diff
 const error = await profileUsernameUpdate(api, accessToken, { username }).catch((error: unknown) => error);
-if (isHttpError(error) && error.status === 400) {
+if (isHttpError(error) && error.status === 422) {
   showInvalidFields(error.tags.invalidFields);
 }
```

A picture the contract refuses, an empty body included (`reason: format`), answered 422 already.

## Replace POSTGRES_DSN with the discrete settings

Remove `POSTGRES_DSN`, and set `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DATABASE` and `POSTGRES_TLS_ENABLED` from the parts of its value:

| Part of `POSTGRES_DSN`       | Variable                                                                                          |
| ---------------------------- | ------------------------------------------------------------------------------------------------- |
| host                         | `POSTGRES_HOST`                                                                                   |
| port                         | `POSTGRES_PORT`, which defaults to `5432`                                                         |
| user                         | `POSTGRES_USER`                                                                                   |
| password                     | `POSTGRES_PASSWORD`, injected as a secret, with any percent-escape decoded, such as `%40` for `@` |
| database                     | `POSTGRES_DATABASE`                                                                               |
| `sslmode=disable`            | `POSTGRES_TLS_ENABLED=false`                                                                      |
| any other `sslmode`, or none | leave `POSTGRES_TLS_ENABLED` unset, which encrypts the link and verifies the server               |

With TLS on, the service checks the server certificate and its host name against the system trust store, as `sslmode=verify-full` did. The modes `require`, `prefer` and `allow`, and a value without `sslmode`, skipped that check, and `verify-ca` skipped the host name: no unverified mode is left. Before the switch, put the database's CA certificate where Go reads trust, in `SSL_CERT_FILE` or a directory listed in `SSL_CERT_DIR`, and set `POSTGRES_HOST` to a name the certificate covers. Production is unaffected: infra already does both.

Without `POSTGRES_HOST`, the process stops at startup with `panic: pgdriver: invalid scheme`. With `POSTGRES_HOST` but no `POSTGRES_PASSWORD`, it stops at startup with `postgres password is empty`. Once the new configuration is deployed and healthy, delete `POSTGRES_DSN` and any secret that held it.
