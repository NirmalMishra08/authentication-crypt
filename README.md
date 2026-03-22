# authentication-crypt 🔐  
Secure Authentication Service (Go + PostgreSQL + Redis)

A minimal authentication backend written in **Go** that focuses on practical security controls for login systems:

- Password hashing with **bcrypt**
- **User enumeration mitigation**
- **Brute-force protection** using Redis:
  - Per-user login attempt tracking + temporary lockout
  - Per-IP rate limiting (basic)

---

## Features

- **Register** users (`POST /register`)
- **Login** users (`POST /login`)
- Passwords stored hashed (bcrypt)
- Generic auth errors (helps prevent user enumeration)
- Redis-backed:
  - failed-attempt counters with TTL
  - account lock keys with TTL
  - per-IP attempt counter with TTL

---

## Tech Stack

- **Go** (HTTP server)
- **Chi** router + middleware logger
- **PostgreSQL** (user storage)
- **Redis** (rate limiting + lockouts)
- **sqlc** style SQL queries (config present)

---

## Project Structure

```text
.
├── main.go          # HTTP server + handlers
├── go.mod / go.sum
├── schema.sql       # users table
├── query.sql        # SQL queries used by sqlc-generated code
├── sqlc.yaml        # sqlc configuration
└── db/              # sqlc-generated Go code (package db)
```

> Note: `db/` is expected to contain sqlc-generated files used by `main.go` (`main.go/db` import).

---

## Requirements

- Go (your `go.mod` specifies `go 1.25.0`)
- PostgreSQL
- Redis
- (Optional but recommended) `sqlc` if you want to regenerate the `db/` package

---

## Environment Variables

Create a `.env` file in the repository root:

```env
POSTGRES_CONN=postgres://user:password@localhost:5432/dbname
REDIS_CONN=redis://localhost:6379
```

Examples:
- `POSTGRES_CONN=postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable`
- `REDIS_CONN=redis://localhost:6379/0`

---

## Database Setup (PostgreSQL)

1) Create the database (example):

```bash
createdb authdb
```

2) Apply schema:

```bash
psql "$POSTGRES_CONN" -f schema.sql
```

This creates the `users` table:

- `id BIGSERIAL PRIMARY KEY`
- `username TEXT NOT NULL`
- `password TEXT`
- `createdAt TIMESTAMPTZ DEFAULT now()`

---

## Redis Setup

Make sure Redis is running locally:

```bash
redis-server
```

The service uses Redis keys like:

- `login:attempts:user:<userID>` (TTL 15 min)
- `login:lock:user:<userID>` (TTL 15 min)
- `login:attempts:ip:<ip>` (TTL 15 min)

---

## Run the Server

Install dependencies:

```bash
go mod tidy
```

Start the server:

```bash
go run main.go
```

Server listens on:
- `http://localhost:8080`

---

## API

### Health check / root

**GET** `/`

Response:
- `Hello world`

---

### Register

**POST** `/register`

Request body:

```json
{
  "username": "testuser",
  "password": "securepassword"
}
```

Responses:
- `201 Created` → `created new user`
- `400 Bad Request` → invalid body or DB insert error

Example:

```bash
curl -i -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"securepassword"}'
```

---

### Login

**POST** `/login`

Request body:

```json
{
  "username": "testuser",
  "password": "securepassword"
}
```

Typical responses:
- `200 OK` → `Login successful`
- `401 Unauthorized` → `Invalid username or password`
- `429 Too Many Requests` → `Too many requests` (per-IP threshold)

Security behavior:
- If the username does not exist, the handler sleeps briefly before responding (reduces user enumeration via timing).
- Failed logins increment counters; too many failures can lock the account temporarily.

Example:

```bash
curl -i -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"securepassword"}'
```

---

## Notes / Known Limitations

- **IP detection** currently uses `r.RemoteAddr` directly. In real deployments behind a proxy/load balancer, you’d typically use `X-Forwarded-For` (carefully, only if you trust the proxy).
- The “progressive delay” is currently a fixed `1s` delay on failures (the commented code suggests an intent to scale by attempt count).
- Login lockout currently returns a `500` status in code; typically this should be `403 Forbidden` or `423 Locked`.

---

## Future Improvements (Ideas)

- JWT access + refresh tokens
- MFA
- Better IP parsing / proxy-aware IP detection
- Structured logging + metrics
- Configurable thresholds (attempt limits, lock durations) via env vars
- Proper HTTP status codes for lockouts

---

## Contributing

PRs/issues welcome. If you add new endpoints, please update the API section and include curl examples.

---

## License

MIT (if you plan to use MIT, consider adding a `LICENSE` file to the repo).