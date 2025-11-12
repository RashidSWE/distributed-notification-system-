```markdown
# User Service — Distributed Notification System

The **User Service** is a core microservice within the **Distributed Notification System**.  
It manages user data, authentication, contact preferences, and exposes APIs for other services (like Notification and Gateway) to interact with user information.

Built with **Fastify**, **TypeScript**, **PostgreSQL**, and **Prisma**, it ensures high performance, scalability, and clean architectural boundaries between domains.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Environment Configuration](#environment-configuration)
- [Database (Prisma + PostgreSQL)](#database-prisma--postgresql)
- [Running the Service](#running-the-service)
- [Docker Setup](#docker-setup)
- [API Endpoints](#api-endpoints)
- [Inter-Service Communication](#inter-service-communication)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

The **User Service** handles:

- User registration and authentication
- Management of user contact info (email, push tokens)
- Notification preference storage
- Permission and role-based access
- Providing user data APIs to other services via REST or message queues

It is part of a modular distributed system that includes:
- **API Gateway** — routes external requests to internal services  
- **Notification Service** — dispatches notifications (email, push, etc.)
- **User Service** — manages user identities and preferences  
- **RabbitMQ / Redis** — message broker and caching layer  
- **PostgreSQL** — primary data store

---

## Features

- JWT-based authentication
- Secure password hashing with bcrypt
- Role-based access management
- User profile CRUD operations
- Notification preferences (email, push)
- Input validation with Joi
- Structured error and success responses
- Built-in Docker support
- Prisma ORM with PostgreSQL

---

## Architecture

The User Service follows a **modular and domain-driven design** (DDD):

##  Tech Stack

| Category | Technology |
|-----------|-------------|
| Language | TypeScript |
| Framework | Fastify |
| Database | PostgreSQL |
| ORM | Prisma |
| Authentication | JWT |
| Message Broker | RabbitMQ |
| Containerization | Docker & Docker Compose |

---
````
## Installation

### Clone the repository

```bash
git clone https://github.com/your-username/distributed-notification-system.git
cd distributed-notification-system/user-service
````

### Install dependencies

```bash
npm install
```

### Generate Prisma Client

```bash
npx prisma generate
```

### Run migrations

```bash
npx prisma migrate dev --name init
```

### Compile TypeScript

```bash
npm run build
```

## Database (Prisma + PostgreSQL)

### Example Prisma Schema

`prisma/schema.prisma`:

After editing the schema:

```bash
npx prisma generate
npx prisma migrate dev --name init
```

---

## Docker Setup

The service is Dockerized for local or containerized deployment.

### Dockerfile Run Locally

```bash
docker-compose up --build user-service
```

Or run everything (gateway, user-service, redis, rabbitmq, postgres):

```bash
docker-compose up --build
```

---

## API Endpoints

| Method | Endpoint                 | Description                    | Auth |
| ------ | ------------------------ | ------------------------------ | ---- |
| POST   | `/api/users/register`    | Register new user              | ❌    |
| POST   | `/api/users/login`       | Authenticate user              | ❌    |
| GET    | `/api/users/profile`     | Retrieve user profile          | ✅    |
| PUT    | `/api/users/profile`     | Update user profile            | ✅    | 
| PUT    |`/api/users/me/push-token`| Update user profile            | ✅    |

> Protected routes require `Authorization: Bearer <token>` header.

---

## Inter-Service Communication

The **User Service** interacts with:

* **Notification Service** via RabbitMQ (publishes events like `user.created`, `preferences.updated`).
* **API Gateway** for request routing.
* **Redis** for caching frequently accessed user data.

Example Event:

```json
{
  "event": "user.created",
  "data": {
    "userId": "e123-uuid",
    "email": "user@example.com",
    "preferences": {
      "email": true,
      "push": false
    }
  }
}
```

---

## Testing

Run unit tests:

```bash
npm test
```

Run integration tests with Docker:

```bash
docker-compose exec user-service npm test
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/user-auth`
3. Commit your changes using conventional commits
4. Push and open a PR


## License

This project is licensed under the **MIT License**.
