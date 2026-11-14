
# Hafiz Pintar Store

Hafiz Pintar Store adalah proyek pengembangan platform e-commerce untuk mendukung penjualan berbagai kategori produk, mulai dari elektronik, smart home, Islamic lifestyle, health & fitness, hingga kebutuhan rumah tangga.

Proyek ini dikembangkan sebagai bagian dari tugas mata kuliah Interaksi Manusia dan Komputer (IMK).

## Tech Stack

### Frontend
- Next.js
- TypeScript
- Flutter
- Dart
- Figma

### Backend
- Golang
- PostgreSQL
- pgx
- Goose
- SQLC

### Infrastructure & Tools
- Docker
- Docker Compose
- Adminer
- Git & GitHub
- Postman

## Project Structure

```text
hafiz-pintar-store/
├── apps/
│   ├── mobile/          # Android application (Flutter)
│   └── web/             # Web application (Next.js)
├── backend/
│   └── api/
│       ├── cmd/
│       │   └── server/
│       ├── db/
│       │   ├── migrations/
│       │   └── queries/
│       └── internal/
│           ├── database/
│           ├── handler/
│           ├── repository/
│           └── service/
├── docs/
├── infrastructure/
│   └── docker/
│       └── docker-compose.yaml
├── .env.example
├── .gitignore
└── README.md
```

## Current Development Progress

- [x] Project repository structure
- [x] Next.js initial setup
- [x] Golang REST API initialization
- [x] PostgreSQL configuration
- [x] Docker Compose environment
- [x] Goose database migration setup
- [x] SQLC code generation
- [x] Product database schema
- [x] Product repository layer
- [x] Product service layer
- [x] Product handler and API routes
- [ ] Authentication and authorization
- [ ] Shopping cart
- [ ] Order management
- [ ] Payment integration
- [ ] Admin dashboard
- [ ] Flutter Android application
- [ ] Frontend API integration
- [ ] Deployment

## Local Development

### Requirements

- Go
- Node.js and npm
- Docker Desktop
- Git
- Goose
- SQLC

### Environment Configuration

Copy `.env.example` to `.env`, kemudian sesuaikan konfigurasi database.

Jangan mengunggah `.env` yang berisi kredensial asli ke GitHub.

### Run Docker

Dari root project:

```powershell
docker compose --env-file .env -f infrastructure/docker/docker-compose.yaml up -d --build
```

### Check Containers

```powershell
docker ps
```

### Test API

Health check:

```powershell
Invoke-RestMethod http://localhost:8080/health
```

Readiness check:

```powershell
Invoke-RestMethod http://localhost:8080/ready
```

Product endpoint:

```powershell
Invoke-RestMethod http://localhost:8080/api/v1/products
```

## Database Development

Backend menggunakan PostgreSQL dengan:

- Goose untuk database migrations.
- SQLC untuk menghasilkan kode Go dari SQL queries.
- pgx sebagai PostgreSQL driver.

## Development Notes

Informasi produk, kategori, harga, stok, dan model bisnis akan disesuaikan setelah proses konfirmasi dan analisis UMKM Hafiz Pintar Store.

## License

Project ini dikembangkan untuk keperluan akademik dan pengembangan sistem Hafiz Pintar Store.
