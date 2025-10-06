# Go Auth + CRUD (Stdlib, GORM, Docker Compose)

## Features
- Authentication
  - Register user (email lowercased, unique, password hashed with bcrypt)
  - Login returns JWT (HS256) with sub=userID and 24h expiry
- Authorization
  - JWT middleware validates Bearer token
  - Admin middleware enforces `is_admin=true` for admin routes
- Categories
  - List categories
  - Get category by id (path param)
  - Create category (admin, unique name validation)
- Videos
  - List videos
  - Get video by id
  - Create video (admin, validates category)
  - Update video (admin, partial updates)
- Uploads
  - Upload file (admin) to `/uploads`, returns stored path
  - Static file serving at `/uploads/*`
- Migrations
  - Auto-migrate `User`, `Category`, `Video` on startup

## Tech Stack
- Go stdlib HTTP server (`net/http`)
- GORM + Postgres
- JWT (github.com/golang-jwt/jwt/v4)
- Dockerfile + Docker Compose

## Project Structure
```
source/auth-crud/
  main.go                    # routes & server startup
  config/database.go         # DB connection + migrations + optional seeding
  handlers/                  # HTTP handlers (auth, category, video, upload)
  middlewares/               # JWT, admin checks, request logging (JSON)
  models/models.go           # GORM models
  utils/util.go              # helpers (hash, JWT, JSON responses, pagination)
  loggers/logger.go          # centralized JSON logger (stdout/file)
orchestrate/
  compose.yml                # docker compose for the service
  auth-crud/
    Dockerfile               # multi-stage build
    auth-crud.env.example    # example env
.docs/
  docs/openapi.yaml          # OpenAPI spec
  docs/postman_collection.json
```

## Prerequisites
- Docker and Docker Compose installed
- A reachable Postgres instance on your host (or change DB_URL accordingly)

## Configuration
1) Create your env from example:
```
cp orchestrate/auth-crud/auth-crud.env.example orchestrate/auth-crud/auth-crud.env
```
2) Edit `orchestrate/auth-crud/auth-crud.env`:
```
DB_URL=postgres://<user>:<password>@host.docker.internal:5432/go_auth_crud?sslmode=disable
JWT_SECRET=change_me
# logging
LOG_OUTPUT=stdout            # or file
LOG_FILE_PATH=/app/logs/app.log
# optional seed demo data at startup
SEED_DATA=false              # set true to insert demo users/categories/videos
```
- Note: URL-encode special characters in password if any (e.g., ! -> %21).

## Run (Docker Compose)
From repo root:
```
docker compose -f orchestrate/compose.yml up -d --build
# view logs (JSON)
docker compose -f orchestrate/compose.yml logs --tail=100
```
Or from `orchestrate/`:
```
cd orchestrate && docker compose up -d --build
```
- Logs are line-delimited JSON. Set `LOG_OUTPUT=file` to write to `LOG_FILE_PATH` inside the container.

## API Endpoints
- Auth
  - POST `/api/v1/auth/register`
    - JSON: {"email":"user@example.com","password":"Passw0rd!"}
  - POST `/api/v1/auth/login`
    - JSON: {"email":"user@example.com","password":"Passw0rd!"}
    - Returns: {"token":"<jwt>"}
- Categories
  - GET `/api/v1/categories?limit=20&cursor=&sort_by=id|created_at&order=asc|desc`
  - GET `/api/v1/categories/{id}`
  - POST `/api/admin/v1/categories` (admin)
    - Headers: Authorization: Bearer <jwt>
    - JSON: {"name":"Tutorials"}
- Videos
  - GET `/api/v1/videos?limit=20&cursor=&sort_by=id|created_at&order=asc|desc`
  - GET `/api/v1/videos/{id}` (preloads `category`)
  - POST `/api/admin/v1/videos` (admin)
    - JSON: {"title":"Intro","duration":"10m","url":"https://...","thumbnailPath":"/uploads/xyz.png","categoryId":1}
  - PUT `/api/admin/v1/videos/{id}` (admin)
    - Partial update JSON allowed
- Uploads
  - POST `/api/admin/v1/uploads` (admin)
    - multipart/form-data: file=<your file>
    - Returns: {"path":"/uploads/<stored-name>"}
  - GET `/uploads/<file>` (public)

## Standard Response Format
All endpoints return a standard envelope:
```
{
  "status": "SUCCESS" | "FAIL",
  "message": "...",
  "data": {}, // object or { items: [...], next_cursor: "..." }
  "error": { "code": "...", "description": "..." },
  "meta": { "timestamp": "...", "request_id": "...", "trace_id": "" }
}
```

## API Documentation (Swagger / Postman)
- OpenAPI (Swagger): `docs/openapi.yaml`
  - View using Swagger UI, Redoc, or VS Code OpenAPI extension.
- Postman collection: `docs/postman_collection.json`
  - Import into Postman, set the `token` variable after login.

## Seeding Demo Data
- Enable seeding: set `SEED_DATA=true` in `orchestrate/auth-crud/auth-crud.env` and restart via docker compose
- Inserts ~30 users, categories, and videos (user1@example.com is admin)

## First-Time Test Flow
1) Register a user
```
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Passw0rd!"}'
```
2) Promote user to admin (via DB)
```
-- example SQL
UPDATE users SET is_admin = true WHERE email = 'admin@example.com';
```
3) Login and copy token
```
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Passw0rd!"}'
```
4) Create a category (admin)
```
curl -s -X POST http://localhost:8080/api/admin/v1/categories \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"name":"Tutorials"}'
```
5) Upload a thumbnail (admin)
```
curl -s -X POST http://localhost:8080/api/admin/v1/uploads \
  -H "Authorization: Bearer <token>" \
  -F file=@/absolute/path/to/file.png
```
6) Create a video (admin)
```
curl -s -X POST http://localhost:8080/api/admin/v1/videos \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"title":"Intro","duration":"10m","url":"https://example.com/video.mp4","thumbnailPath":"/uploads/<stored-name>","categoryId":1}'
```
