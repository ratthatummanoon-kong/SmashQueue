# SmashQueue 🏸

**SmashQueue** is a Badminton Guan (Social Group) Management System designed to streamline matchmaking, queue management, and performance tracking.

![SmashQueue](https://img.shields.io/badge/status-development-green) ![Go](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go) ![Next.js](https://img.shields.io/badge/next.js-16-black?logo=next.js) ![PostgreSQL](https://img.shields.io/badge/postgresql-16-336791?logo=postgresql) ![Docker](https://img.shields.io/badge/docker-ready-2496ED?logo=docker)

---

## 📋 Table of Contents

- [Tech Stack](#-tech-stack)
- [Key Features](#-key-features)
- [Quick Start](#-quick-start)
- [Project Structure](#-project-structure)
- [Shell Scripts](#-shell-scripts)
- [Environment Variables](#️-environment-variables)
- [Database](#️-database)
- [API Endpoints](#-api-endpoints)
- [Security](#-security)
- [Development](#-development)
- [Troubleshooting](#-troubleshooting)

---

## 🚀 Tech Stack

| Layer              | Technology                           |
| ------------------ | ------------------------------------ |
| **Frontend**       | Next.js 16, React 19, Tailwind CSS 4 |
| **Backend**        | Go 1.22+, Chi Router                 |
| **Database**       | PostgreSQL 16                        |
| **Security**       | PASETO v2 Tokens, Argon2id Hashing   |
| **Infrastructure** | Docker, Docker Compose               |

---

## ✨ Key Features

### For Players

- 📊 **Personal Dashboard** - View performance stats and skill levels
- 📈 **Analytics** - Track win rates and records
- 📜 **Match History** - See past opponents and partners
- ⏳ **Queue Status** - Real-time position and wait time

### For Organizers (Hua Guan)

- 🎮 **Queue Management** - Efficient court rotation
- ⚔️ **Smart Matchmaking** - Automated team balancing
- 📝 **Match Recording** - Track results and scores

### For Admins

- 👤 **User Management** - Promote/demote roles
- ⚙️ **System Configuration** - Full system access

---

## 💻 Quick Start

### Prerequisites

- **Go** 1.22+ ([download](https://go.dev/dl/))
- **Node.js** LTS ([download](https://nodejs.org/))
- **PostgreSQL** 16+ ([download](https://www.postgresql.org/download/)) or Docker
- **Docker** (optional, for containerized deployment)

### Option 1: Shell Scripts (Recommended)

```bash
# Clone repository
git clone <repository-url>
cd smashqueue

# First-time setup
./scripts/setup.sh

# Start PostgreSQL
./scripts/db.sh start

# Run database migrations
./scripts/db.sh migrate

# Start all services
./scripts/start.sh
```

### Option 2: Docker Compose

```bash
# Clone and setup
git clone <repository-url>
cd smashqueue

# Copy environment file
cp .env.example .env

# Start all services (PostgreSQL + Backend + Frontend)
docker compose up -d

# View logs
docker compose logs -f
```

### Option 3: Manual Setup

```bash
# Terminal 1: Start PostgreSQL
# (use your local PostgreSQL or Docker)
./scripts/db.sh start

# Terminal 2: Backend
cd backend
cp .env.example .env
go mod tidy
go run main.go

# Terminal 3: Frontend
cd frontend/astro
cp .env.example .env.local
npm install
npm run dev
```

### Access the Application

| Service      | URL                              |
| ------------ | -------------------------------- |
| **Frontend** | http://localhost:3000            |
| **Backend**  | http://localhost:8080            |
| **API Docs** | http://localhost:8080/api/health |

### Default Credentials

| Role      | Username     | Password     |
| --------- | ------------ | ------------ |
| **Admin** | `kong@admin` | `Admin@123!` |

---

## 📁 Project Structure

```
smashqueue/
├── scripts/                 # Shell scripts for automation
│   ├── setup.sh            # First-time setup
│   ├── start.sh            # Start all services
│   ├── stop.sh             # Stop all services
│   └── db.sh               # Database management
│
├── docker-compose.yml       # Docker orchestration
├── Makefile                 # Make commands
├── .env.example             # Root environment template
│
├── frontend/astro/          # Next.js 16 frontend
│   ├── Dockerfile          # Production Docker build
│   ├── .env.example        # Frontend environment template
│   └── app/
│       ├── page.tsx         # Landing page
│       ├── login/           # Login page
│       ├── register/        # Registration with validation
│       ├── dashboard/       # Player dashboard
│       ├── profile/         # Profile management
│       └── components/      # Shared components (Navbar, etc.)
│
├── backend/                 # Go 1.22+ backend
│   ├── Dockerfile          # Production Docker build
│   ├── .env.example        # Backend environment template
│   ├── main.go             # Application entry point
│   ├── config/             # Environment configuration
│   ├── database/           # PostgreSQL connection & repositories
│   ├── model/              # Data models & DTOs
│   ├── service/            # Business logic layer
│   ├── handler/            # HTTP request handlers
│   └── middleware/         # CORS, Auth, Rate limiting
│
└── doc/                     # Documentation
    ├── frontend.md         # Frontend specifications
    └── backend.md          # Backend architecture
```

---

## 📜 Shell Scripts

All scripts are in the `scripts/` directory and are executable.

### Service Management

| Command              | Description               |
| -------------------- | ------------------------- |
| `./scripts/setup.sh` | First-time project setup  |
| `./scripts/start.sh` | Start frontend + backend  |
| `./scripts/stop.sh`  | Stop all running services |

### Database Management

| Command                   | Description                 |
| ------------------------- | --------------------------- |
| `./scripts/db.sh start`   | Start PostgreSQL container  |
| `./scripts/db.sh stop`    | Stop PostgreSQL container   |
| `./scripts/db.sh status`  | Check database connection   |
| `./scripts/db.sh create`  | Create smashqueue database  |
| `./scripts/db.sh migrate` | Run all database migrations |
| `./scripts/db.sh seed`    | Insert sample data          |
| `./scripts/db.sh reset`   | Drop and recreate database  |
| `./scripts/db.sh connect` | Open psql connection        |

---

## ⚙️ Environment Variables

### Root `.env` (for Docker Compose)

```env
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=smashqueue
DB_PORT=5432
SERVER_PORT=8080
FRONTEND_PORT=3000
AUTH_SECRET_KEY=your-32-character-secret-key-here!
CORS_ORIGIN=http://localhost:3000
```

### Backend `backend/.env`

| Variable          | Description                    | Default                 |
| ----------------- | ------------------------------ | ----------------------- |
| `SERVER_PORT`     | HTTP server port               | `8080`                  |
| `DB_HOST`         | PostgreSQL host                | `localhost`             |
| `DB_PORT`         | PostgreSQL port                | `5432`                  |
| `DB_USER`         | Database username              | `postgres`              |
| `DB_PASSWORD`     | Database password              | `postgres`              |
| `DB_NAME`         | Database name                  | `smashqueue`            |
| `DB_SSLMODE`      | SSL mode (disable/require)     | `disable`               |
| `AUTH_SECRET_KEY` | PASETO signing key (32+ chars) | -                       |
| `ADMIN_PASSWORD`  | Initial admin password         | `Admin@123!`            |
| `CORS_ORIGIN`     | Allowed frontend origin        | `http://localhost:3000` |

### Frontend `frontend/astro/.env.local`

| Variable              | Description     | Default                     |
| --------------------- | --------------- | --------------------------- |
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080/api` |

---

## 🗄️ Database

### Schema Overview

```
┌─────────────────┐     ┌─────────────────┐
│     users       │     │   user_stats    │
├─────────────────┤     ├─────────────────┤
│ id              │────▶│ user_id (FK)    │
│ username        │     │ total_matches   │
│ password_hash   │     │ wins / losses   │
│ name / bio      │     │ win_rate        │
│ role            │     │ skill_level     │
└─────────────────┘     └─────────────────┘
        │
        ▼
┌─────────────────┐     ┌─────────────────┐
│ refresh_tokens  │     │  queue_entries  │
├─────────────────┤     ├─────────────────┤
│ user_id (FK)    │     │ user_id (FK)    │
│ token           │     │ position        │
│ expires_at      │     │ status          │
└─────────────────┘     └─────────────────┘
                              │
                              ▼
                        ┌─────────────────┐
                        │    matches      │
                        ├─────────────────┤
                        │ team1[] / team2[]│
                        │ result          │
                        │ started_at      │
                        └─────────────────┘
                              │
                              ▼
                        ┌─────────────────┐
                        │  match_scores   │
                        ├─────────────────┤
                        │ match_id (FK)   │
                        │ game_number     │
                        │ team1/2_score   │
                        └─────────────────┘
```

### Setup Database

```bash
# Using Docker (recommended)
./scripts/db.sh start
./scripts/db.sh migrate
./scripts/db.sh seed

# Using existing PostgreSQL
psql -U postgres -c "CREATE DATABASE smashqueue;"
./scripts/db.sh migrate
```

---

## 🔌 API Endpoints

### Public Endpoints

| Method | Endpoint        | Description              |
| ------ | --------------- | ------------------------ |
| GET    | `/api/health`   | Health check             |
| POST   | `/api/register` | Create new user          |
| POST   | `/api/login`    | Authenticate & get token |
| POST   | `/api/refresh`  | Refresh access token     |

### Protected Endpoints (Requires Bearer Token)

| Method | Endpoint              | Description         |
| ------ | --------------------- | ------------------- |
| POST   | `/api/logout`         | Invalidate session  |
| GET    | `/api/profile`        | Get user profile    |
| PUT    | `/api/profile`        | Update profile      |
| GET    | `/api/profile/stats`  | Get user statistics |
| GET    | `/api/queue`          | Get queue status    |
| POST   | `/api/queue/join`     | Join the queue      |
| POST   | `/api/queue/leave`    | Leave the queue     |
| GET    | `/api/matches`        | Get match history   |
| GET    | `/api/matches/active` | Get ongoing matches |

### Organizer/Admin Endpoints

| Method | Endpoint              | Description         |
| ------ | --------------------- | ------------------- |
| POST   | `/api/queue/call`     | Call next 4 players |
| POST   | `/api/matches`        | Create new match    |
| PUT    | `/api/matches/result` | Record match result |

### Example API Usage

```bash
# Register a new user
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"Test@123!","confirm_password":"Test@123!"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"Test@123!"}'

# Get profile (with token)
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer <your-access-token>"
```

---

## 🔐 Security

| Feature              | Implementation                      |
| -------------------- | ----------------------------------- |
| **Authentication**   | PASETO v2 (symmetric encryption)    |
| **Password Hashing** | Argon2id (memory-hard)              |
| **Access Token**     | 15 min expiry, Bearer header        |
| **Refresh Token**    | 7 days expiry, HttpOnly cookie      |
| **Rate Limiting**    | 10/min (auth), 100/min (API)        |
| **CORS**             | Strict origin policy                |
| **Password Rules**   | 8+ chars, upper/lower/number/symbol |

---

## 👤 User Roles

| Role          | Permissions                            |
| ------------- | -------------------------------------- |
| **Player**    | View profile, stats, join queue        |
| **Organizer** | + Manage queues, create/record matches |
| **Admin**     | + User management, full system access  |

---

## 🛠 Development

### Make Commands

```bash
make help          # Show all available commands
make setup         # First-time setup
make dev           # Start development servers
make build         # Build Docker images
make up            # Start Docker services
make down          # Stop Docker services
make logs          # View Docker logs
make test          # Run tests
make clean         # Remove Docker resources
```

### Code Structure

```
backend/
├── main.go           # Entry point, route registration
├── config/           # Environment configuration
├── model/            # Data structures & DTOs
├── service/          # Business logic (auth, user, queue, match)
├── handler/          # HTTP handlers (request/response)
├── middleware/       # CORS, auth validation, rate limiting
└── database/         # PostgreSQL connection & repositories
```

### Adding New Features

1. **Model** - Define data structure in `model/`
2. **Repository** - Add database operations in `database/`
3. **Service** - Implement business logic in `service/`
4. **Handler** - Create HTTP endpoints in `handler/`
5. **Routes** - Register routes in `main.go`

---

## ❓ Troubleshooting

### Backend won't start

```bash
# Check if port 8080 is in use
lsof -i :8080

# Check database connection
./scripts/db.sh status

# View backend logs
cat logs/backend.log
```

### Frontend won't start

```bash
# Check if port 3000 is in use
lsof -i :3000

# Clear Next.js cache
cd frontend/astro && rm -rf .next && npm run dev
```

### Database connection failed

```bash
# Start PostgreSQL container
./scripts/db.sh start

# Check if database exists
./scripts/db.sh connect
\l  # List databases

# Reset and recreate
./scripts/db.sh reset
```

### Docker issues

```bash
# Rebuild images
docker compose build --no-cache

# View all logs
docker compose logs -f

# Reset everything
docker compose down -v
docker compose up -d
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Made with ❤️ for the badminton community
</p>
