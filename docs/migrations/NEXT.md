# Upgrading to NEXT

A request that parses but fails validation answers 422 instead of 400, and the images no longer read `POSTGRES_DSN`.

## Do I need to act?

**Required** for a client that tells a refused request apart by its status, and for a deployment that sets `POSTGRES_DSN` without `POSTGRES_HOST`. A client that handles every 4xx alike, and a deployment that sets `POSTGRES_HOST`, take this release as-is.

## Read 422 for a request that fails validation

The service now answers as the authentication service does: 400 for a request it cannot parse, 422 for one that parses but breaks a rule. The problem body keeps its shape, with `status` 422 and the title `Unprocessable Entity`.

| Request                                             | Before | After |
| --------------------------------------------------- | ------ | ----- |
| `PUT /v1/profile` with a refused username           | 400    | 422   |
| `PUT /v1/profile/picture` with an empty body        | 400    | 422   |
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

A picture the contract refuses answered 422 already, with its `reason` tag.

## Replace POSTGRES_DSN with the discrete settings

Split the URL `postgres://<user>:<password>@<host>:<port>/<database>?sslmode=<mode>` into one variable per part:

| URL part                     | Variable                                                    |
| ---------------------------- | ----------------------------------------------------------- |
| `<host>`                     | `POSTGRES_HOST`                                             |
| `<port>`                     | `POSTGRES_PORT`, which defaults to `5432`                   |
| `<user>`                     | `POSTGRES_USER`                                             |
| `<password>`                 | `POSTGRES_PASSWORD`, injected as a secret                   |
| `<database>`                 | `POSTGRES_DATABASE`                                         |
| `sslmode=disable`            | `POSTGRES_TLS_ENABLED=false`                                |
| any other `sslmode`, or none | leave `POSTGRES_TLS_ENABLED` unset, which encrypts the link |

```sh
# before
POSTGRES_DSN=postgres://agora:p%40ss@db.internal:5432/agora?sslmode=disable

# after
POSTGRES_HOST=db.internal
POSTGRES_PORT=5432
POSTGRES_USER=agora
POSTGRES_PASSWORD=p@ss
POSTGRES_DATABASE=agora
POSTGRES_TLS_ENABLED=false
```

`POSTGRES_PASSWORD` holds the raw password: decode any percent-escape the URL carried, such as `%40` for `@`.

Without `POSTGRES_HOST`, the first database connection fails with `pgdriver: invalid scheme`. With `POSTGRES_HOST` but no `POSTGRES_PASSWORD`, the process stops at startup with `postgres password is empty`. Once the new configuration is deployed and healthy, delete `POSTGRES_DSN` and any secret that held it.
