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
  config/database.go         # DB connection + migrations
  handlers/                  # HTTP handlers (auth, category, video, upload)
  middlewares/               # JWT and admin checks
  models/models.go           # GORM models
  utils/util.go              # helpers (hash, JWT, JSON helpers)
orchestrate/
  compose.yml                # docker compose for the service
  auth-crud/
    Dockerfile               # multi-stage build
    auth-crud.env.example    # example env
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
```
- Note: URL-encode special characters in password if any (e.g., ! -> %21).

## Run (Docker Compose)
```
docker compose -f orchestrate/compose.yml up -d --build
# view logs
docker compose -f orchestrate/compose.yml logs --tail=100
```
- Logs should include: "Connected to database successfully" and "Running DB migrations..."

## API Endpoints
- Auth
  - POST `/api/v1/auth/register`
    - JSON: {"email":"user@example.com","password":"Passw0rd!"}
  - POST `/api/v1/auth/login`
    - JSON: {"email":"user@example.com","password":"Passw0rd!"}
    - Returns: {"token":"<jwt>"}
- Categories
  - GET `/api/v1/categories`
  - GET `/api/v1/categories/{id}`
  - POST `/api/admin/v1/categories` (admin)
    - Headers: Authorization: Bearer <jwt>
    - JSON: {"name":"Tutorials"}
- Videos
  - GET `/api/v1/videos`
  - GET `/api/v1/videos/{id}`
  - POST `/api/admin/v1/videos` (admin)
    - JSON: {"title":"Intro","duration":"10m","url":"https://...","thumbnailPath":"/uploads/xyz.png","categoryId":1}
  - PUT `/api/admin/v1/videos/{id}` (admin)
    - Partial update JSON allowed
- Uploads
  - POST `/api/admin/v1/uploads` (admin)
    - multipart/form-data: file=<your file>
    - Returns: {"path":"/uploads/<stored-name>"}
  - GET `/uploads/<file>` (public)

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

## Notes
- The service expects Postgres reachable at `host.docker.internal:5432` from inside the container. Adjust DB_URL if your DB runs elsewhere.
- You can serve uploads through any reverse proxy; the app serves `/uploads/*` directly from the container filesystem.
- For production, rotate a strong `JWT_SECRET` and secure DB credentials.
