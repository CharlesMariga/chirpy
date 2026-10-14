# Chirpy

Chirpy is a small Twitter-style social network backend written in Go. Users sign up, log in, and post short messages called *chirps*. It ships with JWT authentication, refresh tokens, a webhook-driven premium tier, and a Postgres data layer generated with sqlc.

## Why Chirpy?

Chirpy is a compact, realistic reference for building an HTTP API in Go with no web framework. It covers the pieces most production services need, in a codebase small enough to read in one sitting:

- **Standard library routing** with Go 1.22+ method-aware `http.ServeMux` patterns.
- **Real authentication**: Argon2id password hashing, short-lived JWT access tokens, and long-lived, revocable refresh tokens.
- **Type-safe SQL** via [sqlc](https://sqlc.dev), with schema migrations managed by [goose](https://github.com/pressly/goose).
- **Authorization**: users can only edit their own account and delete their own chirps.
- **Webhooks**: a third-party payment provider ("Polka") upgrades users to *Chirpy Red* through an API-key-protected endpoint.
- **Filtering and sorting**: list chirps by author and order them by creation time.

If you are learning Go or want a clean starting point for your own API, Chirpy is a good place to begin.

## Requirements

- Go 1.22 or newer
- PostgreSQL 14 or newer
- [goose](https://github.com/pressly/goose) for migrations
- [sqlc](https://sqlc.dev) only if you change the SQL queries

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Installation

1. **Clone the repository**

   ```sh
   git clone https://github.com/CharlesMariga/chirpy.git
   cd chirpy
   ```

2. **Create a database**

   ```sh
   createdb chirpy
   ```

3. **Configure the environment**

   Create a `.env` file in the project root:

   ```env
   DB_URL="postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   JWT_SECRET="a-long-random-string"
   POLKA_KEY="your-polka-api-key"
   ```

   | Variable     | Purpose |
   |--------------|---------|
   | `DB_URL`     | Postgres connection string. |
   | `PLATFORM`   | Set to `dev` to enable the `POST /admin/reset` endpoint. Any other value disables it. |
   | `JWT_SECRET` | Secret used to sign access tokens. Generate one with `openssl rand -base64 64`. |
   | `POLKA_KEY`  | API key the Polka webhook must present in the `Authorization` header. |

4. **Run the migrations**

   ```sh
   cd sql/schema
   goose postgres "$DB_URL" up
   cd ../..
   ```

5. **Build and run**

   ```sh
   go build -o chirpy && ./chirpy
   ```

   The server listens on `http://localhost:8080`. Confirm it is up:

   ```sh
   curl http://localhost:8080/api/healthz
   ```

## API reference

Authenticated endpoints expect `Authorization: Bearer <token>`. Access tokens expire after one hour. Refresh tokens expire after sixty days.

### Health and admin

| Method | Path             | Description |
|--------|------------------|-------------|
| GET    | `/api/healthz`   | Readiness check. Returns `OK`. |
| GET    | `/admin/metrics` | HTML page showing how many times `/app/` has been visited. |
| POST   | `/admin/reset`   | Resets the hit counter and deletes all users. Only works when `PLATFORM=dev`. |
| GET    | `/app/`          | Serves the static landing page. |

### Users

| Method | Path          | Auth   | Description |
|--------|---------------|--------|-------------|
| POST   | `/api/users`  | none   | Create a user from `{"email", "password"}`. |
| PUT    | `/api/users`  | bearer | Update the authenticated user's email and password. |
| POST   | `/api/login`  | none   | Exchange `{"email", "password"}` for a user object plus `token` and `refresh_token`. |
| POST   | `/api/refresh`| bearer (refresh token) | Issue a new access token. Returns `{"token"}`. |
| POST   | `/api/revoke` | bearer (refresh token) | Revoke the refresh token. Returns `204`. |

User objects look like:

```json
{
  "id": "4b0bb6f3-5a2e-4ad6-8e2e-5a0b0c7f4e21",
  "created_at": "2026-10-14T09:17:40.330364Z",
  "updated_at": "2026-10-14T09:17:40.330364Z",
  "email": "ada@example.com",
  "is_chirpy_red": false
}
```

### Chirps

| Method | Path                      | Auth   | Description |
|--------|---------------------------|--------|-------------|
| POST   | `/api/chirps`             | bearer | Create a chirp from `{"body"}`. Max 140 characters. Certain words are replaced with `****`. |
| GET    | `/api/chirps`             | none   | List chirps. Supports the query parameters below. |
| GET    | `/api/chirps/{chirpID}`   | none   | Fetch a single chirp. |
| DELETE | `/api/chirps/{chirpID}`   | bearer | Delete a chirp. Only the author may delete it. Returns `204`. |

`GET /api/chirps` query parameters:

| Parameter   | Values          | Default | Description |
|-------------|-----------------|---------|-------------|
| `author_id` | user UUID       | all     | Only return chirps by this user. |
| `sort`      | `asc` or `desc` | `asc`   | Order by `created_at`. |

Example:

```sh
curl "http://localhost:8080/api/chirps?author_id=4b0bb6f3-5a2e-4ad6-8e2e-5a0b0c7f4e21&sort=desc"
```

### Webhooks

| Method | Path                   | Auth   | Description |
|--------|------------------------|--------|-------------|
| POST   | `/api/polka/webhooks`  | `Authorization: ApiKey <POLKA_KEY>` | Handles `{"event": "user.upgraded", "data": {"user_id"}}` by setting `is_chirpy_red` to `true`. Other events return `204` and are ignored. Unknown user IDs return `404`. |

## Quick start walkthrough

```sh
# Create a user
curl -X POST localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"correct-horse"}'

# Log in and grab the access token
TOKEN=$(curl -s -X POST localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"correct-horse"}' | jq -r .token)

# Post a chirp
curl -X POST localhost:8080/api/chirps \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"body":"Hello, Chirpy!"}'

# Read it back, newest first
curl "localhost:8080/api/chirps?sort=desc"
```

## Development

Run the test suite:

```sh
go test ./...
```

After editing anything in `sql/queries/` or `sql/schema/`, regenerate the database package:

```sh
sqlc generate
```

Project layout:

```
.
├── main.go                 # Server setup and routing
├── handler_*.go            # One file per endpoint group
├── json.go                 # JSON response helpers
├── internal/
│   ├── auth/               # Password hashing, JWT, token helpers
│   └── database/           # sqlc-generated query code (do not edit)
├── sql/
│   ├── schema/             # goose migrations
│   └── queries/            # sqlc query definitions
├── index.html, assets/     # Static landing page served at /app/
└── sqlc.yaml
```
