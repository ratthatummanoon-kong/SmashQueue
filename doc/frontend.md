# SmashQueue 🏸

**SmashQueue** is a Badminton Guan (Social Group) Management System designed to streamline matchmaking, queue management, and performance tracking.

## 🚀 Tech Stack

- **Frontend:** Next.js (React), Tailwind CSS
- **Backend:** Go (Golang)
- **Database:** PostgreSQL
- **Security:** PASETO Tokens, Argon2id Hashing
- **Infrastructure:** Docker
- **CI/CD:** Drone CI (Go-based)

---

## 🔐 Authentication & Security Specification

This system implements a high-security authentication flow separating Frontend and Backend concerns.

### 1. Registration Rules

- **Fields:** `username`, `password`, `confirm_password`
- **Default Role:** `Player` (All new users start as players)
- **Validation Logic:**
  - **Username:** Unique, required.
  - **Password Strength:**
    - Minimum length: **8 characters**
    - Must contain: `a-z`, `A-Z`, `0-9`, and Special Characters (Symbols/Punctuation).
    - **Constraint:** Password must **NOT** be identical to the username.

### 2. Login & Session Management

- **Method:** Username & Password
- **Token Standard:** **PASETO** (Platform-Agnostic Security Tokens) v4
- **Session Flow:**
  - **Access Token:** Short-lived (15 minutes). Sent in Authorization Header (Bearer).
  - **Refresh Token:** Long-lived (7 days). Stored securely in an **HttpOnly, Secure, SameSite Cookie**.
- **Rate Limiting:** Implemented on Login/Register endpoints to prevent Brute Force attacks.
- **CORS:** Strict policy allowing requests only from the trusted Frontend domain.

### 3. Password Security

- **Hashing Algorithm:** **Argon2id** (Resistant to GPU/ASIC cracking).
- **Storage:** Passwords are never stored in plain text.

---

## 👤 User Roles & Permissions

### Role Hierarchy

1.  **Player (Default):** Can view own profile, stats, and join queues.
2.  **Organizer (Hua Guan):** Can manage queues and match players.
3.  **Admin:** Full system access.

### Admin Features

- **Initial Super Admin:**
  - Username: `kong@admin`
  - _Note: Initial password is set via environment variables on deployment._
- **Role Management:** Admin can promote/demote users (e.g., Change `Player` to `Organizer`) via the Admin Dashboard.

---

## 🛠️ Feature Modules

### Profile Management

- **View Profile:** Display user stats, win rate, and match history.
- **Edit Profile:** Update personal information (Name, Bio, Avatar).

---

## 💻 Development Setup

### Prerequisites

- Go 1.22+
- Node.js (LTS) & pnpm/npm
- Docker & Docker Compose

### Getting Started (Frontend)

```bash
cd frontend
npm install
npm run dev
```
