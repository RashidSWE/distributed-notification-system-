
````markdown
# Distributed Notification System

## 🚀 Project Overview
This project is a **distributed notification system** designed to send **emails** and **push notifications** using **microservices**. Each service communicates asynchronously through **RabbitMQ** and uses **Redis** for caching and temporary states.

**Key Features:**
- API Gateway for routing requests
- Email and Push notification services
- Template management service
- User service for preferences and contact info
- Retry system with dead-letter queue
- Health checks and idempotency
- Scalable microservices architecture

---

## 🏗 Services & Technologies

| Service           | Language / Framework | Responsibility                                     | Branch Name               |
|------------------|-------------------|--------------------------------------------------|--------------------------|
| API Gateway       | Python / FastAPI   | Entry point, auth, validation, routing, status  | `feature/api-gateway`     |
| User Service      | TypeScript /   | User CRUD, preferences, REST APIs                | `feature/user-service`    |
| Email Service     | Python / FastAPI   | Consume `email.queue`, send emails, handle bounces | `feature/email-service`   |
| Template Service  | Python / FastAPI   | Manage templates, variables, versions           | `feature/template-service`|
| Push Service      | Go                 | Consume `push.queue`, send push notifications   | `feature/push-service`    |

**Shared Tools:**
- RabbitMQ: Message queue
- Redis: Caching and rate-limiting
- PostgreSQL: Service-specific data storage
- Docker: Containerization

---

## ⚙️ Local Setup

### 1. Clone Repository
```bash
git clone https://github.com/RashidSWE/distributed-notification-system-.git
cd distributed-notification-system
````

### 2. Copy Environment Variables

```bash
cp .env.example .env
```

Update `.env` with your local credentials.

---

### 3. Start Docker Services

```bash
docker-compose up --build
```

This starts:

* RabbitMQ (`5672` for AMQP, `15672` for dashboard)
* Redis (`6379`)
* API Gateway (`8000`)

> Other services can be started individually using their respective Dockerfiles.

---

### 4. API Gateway Health Check

Visit:

```
http://localhost:8000/api/health
```

Expected response:

```json
{
  "rabbitmq": "connected",
  "redis": "connected"
}
```

---

## 🧑‍🤝‍🧑 Collaboration & Branch Workflow

### **1. Branch-per-service workflow**

* Each developer works on their **own feature branch**
* Examples:

```bash
git checkout main
git pull origin main
git checkout -b feature/api-gateway
git checkout -b feature/user-service
git checkout -b feature/email-service
git checkout -b feature/template-service
git checkout -b feature/push-service
```

### **3. Push your feature branch**

```bash
git add .
git commit -m "Add initial setup for user servie"
git push -u origin feature/user-service
```

### **4. Pull Request (PR)**

1. On GitHub, create a PR: **base branch** → `main`, **compare branch** → your feature branch
2. Add **title & description**
3. Review changes and merge PR to `main`


### **5. Keep local `main` updated**

```bash
git checkout main
git pull origin main
```

---

## 📦 Docker Commands

* Build and start all services:

```bash
docker-compose up --build
```

* Stop services:

```bash
docker-compose down
```

* Rebuild a single service (example: API Gateway):

```bash
docker-compose up --build api-gateway
```

---

## 📚 Notes

* Each developer runs RabbitMQ and Redis locally for development
* `.env` file contains local credentials (never commit secrets)
* Branches should be **short-lived** and merged via PR
* Logging, monitoring, and metrics will be implemented per service

---

## 🌐 .env.example

```dotenv
# RabbitMQ
RABBITMQ_USER=guest
RABBITMQ_PASS=guest
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_MANAGEMENT_PORT=15672

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# API Gateway
API_HOST=0.0.0.0
API_PORT=8000
```

---

This README serves as the **central guide** for all developers to set up, work on their services.

```