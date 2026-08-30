# url-shortener

[English](./README.md) | [한국어](./README.ko.md)

Minimal URL shortener built with Go's standard library and SQLite — no web framework, no ORM.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![SQLite](https://img.shields.io/badge/DB-SQLite-003B57?logo=sqlite&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- Shorten a long URL into a 7-character code
- Redirect a short code to its original URL
- Duplicate URLs reuse their existing short code instead of creating a new one
- URL format validation

## Tech Stack

- **Go** — `net/http` from the standard library, no framework
- **SQLite** — via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite), a pure-Go driver (no CGO)
- **Docker / Docker Compose** — isolated dev environment

## Getting Started

### Prerequisites

- Docker Desktop

### Run

```bash
git clone https://github.com/dev1f965x/url-shortener.git
cd url-shortener
docker compose up --build
```

Server starts on `http://localhost:8080`.

## API

### Shorten a URL

```
POST /shorten
Content-Type: application/json

{"url": "https://example.com/very/long/path"}
```

Returns `201 Created` for a new URL, or `200 OK` if that URL was already shortened before:

```json
{
  "short_code": "q1km5u2",
  "short_url": "http://localhost:8080/q1km5u2",
  "original_url": "https://example.com/very/long/path"
}
```

Returns `400 Bad Request` if the URL is missing or malformed.

### Redirect

```
GET /{short_code}
```

Responds with `302 Found` and a `Location` header pointing at the original URL, or `404 Not Found` if the code doesn't exist.

## Data Model

```sql
CREATE TABLE urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_url TEXT NOT NULL UNIQUE,
    short_code TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Roadmap

- [ ] Custom slugs
- [ ] Click count tracking
- [ ] Link expiration
- [ ] Link deletion/deactivation
- [ ] QR code generation
- [ ] Click analytics dashboard
- [ ] Account system
- [ ] API key issuance

## License

[MIT](./LICENSE)
