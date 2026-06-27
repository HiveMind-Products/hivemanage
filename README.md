# Fivemanage Lite

Fivemanage Lite is an open-source, lightweight management service designed for gaming communities. It provides essential features for file storage, structured logging, and community organization.

<kbd>
<img width="1790" height="1001" style="border-radius: 15px;" alt="Screenshot 2026-01-18 at 01 43 59" src="https://github.com/user-attachments/assets/6cef883d-3961-495f-b5f6-e3204aac6664" />
</kbd>


## Quick Start

Run the latest version of Fivemanage Lite using Docker:

```bash
docker run -p 8080:8080 \
  -e DSN=postgres://user:pass@host:5432/db \
  -e ADMIN_PASSWORD=your_secure_password \
  -e API_TOKEN_HMAC_SECRET=your_32_byte_secret \
  ghcr.io/fivemanage/lite:0.1.0-beta.18
```

---

## Features

- Multi-tenant organization support.
- File storage with S3-compatible providers (AWS S3, Cloudflare R2, MinIO).
- High-performance structured logging powered by ClickHouse.
- Built-in authentication and session management.
- OpenTelemetry integration for tracing.
- Modern React-based administrative dashboard.

---

## Hosting Options

### Docker Compose
For production environments, using Docker Compose is the most straightforward method to manage the application along with its dependencies (PostgreSQL and ClickHouse). You can find a template in the `deployments/docker-compose.yml` file.

### Kubernetes
Fivemanage Lite is stateless and can be easily deployed on Kubernetes. It is recommended to use a standard Deployment for the application and managed services for the database and storage. Ensure you configure the required environment variables via Secrets or ConfigMaps.

---

## Configuration

The application is configured via environment variables.

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Port the server listens on | `8080` |
| `DSN` | PostgreSQL connection string | - |
| `ADMIN_PASSWORD` | Initial password for the 'admin' user | `password` |
| `API_TOKEN_HMAC_SECRET` | 32-byte secret for signing API tokens | - |
| `S3_PROVIDER` | S3 provider (`s3`, `r2`, or `minio`) | `minio` |
| `AWS_ACCESS_KEY_ID` | S3 access key ID | - |
| `AWS_SECRET_ACCESS_KEY` | S3 secret access key | - |
| `AWS_ENDPOINT` | S3 endpoint URL | - |
| `AWS_BUCKET` | S3 bucket name | - |
| `AWS_REGION` | S3 region | - |
| `BUCKET_DOMAIN` | Public CDN base URL for served files. Used as the V3 `url` field; the V3 `originalUrl` is derived from `AWS_ENDPOINT`/`AWS_BUCKET`. | - |
| `CLICKHOUSE_HOST` | ClickHouse host address | `localhost:19000` |
| `CLICKHOUSE_USERNAME` | ClickHouse username | `default` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | `password` |
| `CLICKHOUSE_DATABASE` | ClickHouse database name | `default` |
| `ENV` | Environment mode (`production` or `dev`) | `dev` |

---

## Public API (V3)

The public API is compatible with the [Fivemanage V3 API](https://docs.fivemanage.com/api-reference/introduction). All routes are served under `/api` and authenticated with an organization API token (created on the dashboard Tokens page):

```
Authorization: <YOUR_API_TOKEN>
```

Successful responses use the envelope `{ "status": "ok", "data": { ... } }`; errors use `{ "error": "..." }` with an appropriate status code (`400`, `401`, `404`, `413`, `500`).

### Files

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v3/file` | Multipart upload (file field is always `file`); optional `filename`, `path`, `metadata` (JSON string), `retentionExempt`. |
| `POST` | `/api/v3/file/base64` | JSON upload `{ base64, filename?, path?, metadata?, retentionExempt? }` (accepts data-URI). |
| `GET` | `/api/v3/file` | List files: `page` (1), `limit` (50, max 100), `type`, `path`. |
| `GET` | `/api/v3/file/*` | Get a file by id or storage key. |
| `DELETE` | `/api/v3/file/*` | Delete a file by id or storage key. |
| `GET` | `/api/v3/file/presigned-url` | Create a presigned upload URL; optional `expiresAt` (unix), `path`. Default expiry 15 min. |
| `POST` | `/api/v3/file/presigned-url/:token` | Upload via a presigned URL. Authenticated by the token, **not** an API key. |

Upload responses return `data: { id, url, originalUrl }`. `originalUrl` is the default storage URL; `url` uses `BUCKET_DOMAIN` (CDN) when configured. `FileItemV3` (list/get) is `{ id, filename, type, size, url, originalUrl, metadata }`.

### Logging

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v3/logs` | Ingest logs. The body **must** be a JSON array of `{ level, message, resource?, timestamp?, metadata? }`. The target dataset is set via the `X-Fivemanage-Dataset` header. |

### V2 → V3 mapping

The legacy V2 routes still work as deprecated aliases (they emit a `Deprecation: true` response header and reuse the V3 logic internally):

| V2 (deprecated) | V3 |
|-----------------|----|
| `POST /api/image`, `/api/video`, `/api/audio`, `/api/file` | `POST /api/v3/file` |
| `POST /api/logs` (single object **or** array) | `POST /api/v3/logs` (array only) |

The legacy media routes keep their original flat `{ "url": ... }` response shape; `/api/logs` additionally accepts a single log object and wraps it into a one-element batch.

### Examples

```bash
# Multipart upload
curl -X POST https://your-host/api/v3/file \
  -H "Authorization: $API_TOKEN" \
  -F "file=@photo.png" \
  -F "path=avatars" \
  -F 'metadata={"alt":"profile"}'

# Base64 upload
curl -X POST https://your-host/api/v3/file/base64 \
  -H "Authorization: $API_TOKEN" -H "Content-Type: application/json" \
  -d '{"base64":"<BASE64>","filename":"note.txt"}'

# List images
curl "https://your-host/api/v3/file?type=image&limit=20" -H "Authorization: $API_TOKEN"

# Get / delete by id
curl https://your-host/api/v3/file/<id> -H "Authorization: $API_TOKEN"
curl -X DELETE https://your-host/api/v3/file/<id> -H "Authorization: $API_TOKEN"

# Presigned URL: generate, then upload without an API key
PRESIGNED=$(curl -s "https://your-host/api/v3/file/presigned-url" \
  -H "Authorization: $API_TOKEN" | jq -r '.data.presignedUrl')
curl -X POST "$PRESIGNED" -F "file=@clip.mp4"

# Ingest logs (array only)
curl -X POST https://your-host/api/v3/logs \
  -H "Authorization: $API_TOKEN" -H "Content-Type: application/json" \
  -H "X-Fivemanage-Dataset: my-dataset" \
  -d '[{"level":"info","message":"hello","metadata":{"requestId":"abc"}}]'
```

---

## Initial Login

Once the application is running, access the dashboard at `http://localhost:8080`.

- **Username:** `admin`
- **Password:** The value of your `ADMIN_PASSWORD` environment variable.

---

## Development & Contribution

Follow these instructions if you want to contribute to the project or build from source.

### Prerequisites

- **Go**: 1.24 or later
- **Node.js**: 22.x or later
- **pnpm**: 9.x or later
- **Air**: For backend hot-reloading
- **Docker**: For running local infrastructure

### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/fivemanage/fivemanage-lite.git
   cd fivemanage-lite
   ```

2. **Install Frontend Dependencies:**
   ```bash
   cd web
   pnpm install
   cd ..
   ```

3. **Download Backend Dependencies:**
   ```bash
   go mod download
   ```

4. **Start Development Infrastructure:**
   ```bash
   docker compose -f deployments/docker-compose.yml up -d
   ```

### Running for Development

Run the backend and frontend in separate terminals.

1. **Backend:**
   ```bash
   air
   ```

2. **Frontend:**
   ```bash
   cd web
   pnpm dev
   ```

---

## Contributing

1. **Branching:** Use descriptive branch names (`feature/*`, `fix/*`).
2. **Formatting:** Use `go fmt` and `pnpm lint`.
3. **Commits:** Provide clear, professional messages.
4. **Pull Requests:** Provide a detailed description of changes.

## Project Structure

- `cmd/lite`: Main entry point and CLI configuration.
- `internal/`: Application logic, API handlers, and services.
- `pkg/`: Reusable packages for storage, logging, and caching.
- `web/`: Frontend React application.
- `deployments/`: Docker Compose and deployment configurations.
- `migrate/`: PostgreSQL schema migrations.
- `build/`: Dockerfile and build scripts.

## License

Fivemanage Lite is released under the [MIT License](LICENSE.md).
