# GoDepot

GoDepot is a Go HTTP service for authenticated file indexing and retrieval. Users can register or log in, choose a local folder to synchronize, list indexed files, and fetch file content with optional processing such as image resizing or format conversion.

## Features

- User registration and login with JWT bearer tokens
- PostgreSQL-backed user storage
- Per-user synchronized folder watching
- In-memory file index and short-lived file response cache
- File listing and content serving
- Image processing for JPEG, PNG, and GIF files
- Basic metadata responses for text, Markdown, PDF, and video files

## Tech Stack

- Go 1.25.6
- Gorilla Mux
- PostgreSQL 16
- pgx
- fsnotify
- JWT
- Docker Compose

## Project Structure

```text
.
+-- main.go
+-- service/                         # Application wiring
+-- domain/                          # Entities, rules, use cases, DTOs
+-- infrastructure/
|   +-- datastore/                   # Database, repositories, cache, file index
|   +-- files/                       # File watcher and processors
|   +-- router/                      # HTTP routing and modules
|   +-- script/migrate/              # Database migrations
|   +-- security/                    # JWT and password hashing
+-- docker-compose.yml
+-- .env-example
```

## Requirements

- Go installed
- Docker and Docker Compose

## Getting Started

Create a local environment file:

```bash
cp .env-example .env
```

Start PostgreSQL:

```bash
docker compose up -d
```

Run the API:

```bash
go run .
```

The server listens on:

```text
http://localhost:8080
```

## Environment Variables

```env
DATABASE_URL=postgres://godepot:godepot@localhost:5432/godepot?sslmode=disable
JWT_SECRET=change-me
```

`DATABASE_URL` is required. `JWT_SECRET` is used to sign and validate JWTs and should be changed for non-local environments.

## API

### Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

Returns a JWT token as JSON and in the `Authorization` response header.

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Set Sync Folder

```bash
curl -X POST http://localhost:8080/files/sync-folder \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"path":"C:\\Users\\you\\Pictures"}'
```

### List Files

```bash
curl http://localhost:8080/files \
  -H "Authorization: Bearer <token>"
```

### Get File Content

```bash
curl "http://localhost:8080/files/content?name=photo.jpg" \
  -H "Authorization: Bearer <token>" \
  --output photo.jpg
```

For supported images, optional query parameters can process the output:

```bash
curl "http://localhost:8080/files/content?name=photo.jpg&w=800&h=600&format=jpeg&quality=85" \
  -H "Authorization: Bearer <token>" \
  --output resized.jpg
```

Supported query parameters:

- `name`: indexed file name, required
- `w`: output width
- `h`: output height
- `format`: output format, such as `jpeg` or `png`
- `quality`: JPEG quality from `0` to `100`

## Database

The database container builds from `infrastructure/datastore/db/Dockerfile`. On first startup, it runs:

```text
infrastructure/script/migrate/001-create-tables.up.sql
```

This creates the `users` table and enables the `pgcrypto` extension for UUID generation.

## Development

Run formatting before committing changes:

```bash
go fmt ./...
```

Run tests when tests are added:

```bash
go test ./...
```
