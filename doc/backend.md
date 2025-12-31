# SmashQueue Backend Documentation 🏸

This directory contains technical documentation regarding the backend architecture, database workflows, and security standards for the SmashQueue application.

## 🏗️ System Overview

The backend is built using **Go (Golang)** and follows a **Clean / Layered Architecture**. It is designed with production readiness in mind, prioritizing security and scalability from the start.

### Key Architectural Decisions

- **HTTPS First:** The server architecture is designed to support **HTTPS/TLS** protocols to ensure data integrity and security for future production deployment.
- **RESTful API:** Structured API endpoints handling JSON payloads.

---

## 💾 Database & Data Persistence Strategy

The project currently uses a local PostgreSQL instance for development, with a file-based strategy for version controlling data states.

### 1. Local Database (PostgreSQL)

- **Current State:** The application connects to a PostgreSQL instance running on `localhost`.
- **Driver:** (e.g., `pgx` or `gorm`) configured to handle connection pooling.

### 2. Mock Data & Git Workflow

To facilitate collaboration without needing a shared remote database server, we utilize a text-file export strategy:

- **Concept:** Database records (e.g., users, match stats) are exported/serialized into **Text Files** (stored in `database/db/`).
- **Benefit:** These files are committed to **Git**, acting as a "Mockup Database" or "Snapshot."
- **Seeding:** Developers can reconstruct the database state by running the mock generator, which reads these text files and populates the local PostgreSQL instance.

---

## 🛠️ Project Structure

```text
backend/
├── cmd/
│   ├── app/            # Main entry point (Production HTTPS Server)
│   └── mock/           # Utility to seed DB from text files
├── config/             # Environment configs (DB Host, HTTPS Certs paths)
├── database/
│   ├── db/             # Text-based data snapshots (Git-tracked)
│   └── generate/       # Logic to parse text files and insert into PSQL
├── handler/            # HTTP Handlers
├── middleware/         # Security middleware (CORS, Auth, TLS)
├── model/              # Database Structs & DTOs
├── service/            # Business Logic
└── main.go             # Root entry point
```
