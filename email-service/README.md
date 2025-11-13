# Email Service

Async FastAPI service responsible for sending ready-made emails via SMTP. Templates and other enrichment happen outside this service.

## Requirements

- Python 3.11+
- Dependencies from `requirements.txt`

## Environment Variables

Copy `.env.example` to `.env` (or export variables) and set:

- `SMTP_HOST`, `SMTP_PORT`
- `SMTP_USERNAME`, `SMTP_PASSWORD`
- `SMTP_USE_TLS` (`true`/`false`)
- `SMTP_USE_SSL` (`true`/`false`)
- `EMAIL_FROM` (default sender address)
- `SMTP_TIMEOUT_SECONDS`
- `EMAIL_MAX_RETRIES`
- `EMAIL_QUEUE_ENABLED` (`true` to start the RabbitMQ consumer)
- `RABBITMQ_HOST`, `RABBITMQ_PORT`
- `RABBITMQ_USERNAME`, `RABBITMQ_PASSWORD`, `RABBITMQ_VHOST`
- `RABBITMQ_QUEUE_NAME`
- `RABBITMQ_PREFETCH`
- `RABBITMQ_RECONNECT_DELAY`
- `RABBITMQ_STATUS_QUEUE`
- `RABBITMQ_FAILED_QUEUE`
- `TEMPLATE_SERVICE_ENDPOINT`

## Run locally

```bash
pip install -r requirements.txt
uvicorn app.main:app --reload --host 0.0.0.0 --port 8080
```

Set `EMAIL_QUEUE_ENABLED=true` to have the app listen to RabbitMQ and process queued emails in the same process.

When queue processing or HTTP sends complete, the service publishes delivery summaries to `RABBITMQ_STATUS_QUEUE` (and `RABBITMQ_FAILED_QUEUE` when the send ultimately fails).

## Docker

```bash
docker build -t email-service .
docker run --env-file .env -p 8080:8080 email-service
```

## API

- `GET /health` – readiness probe.
- `POST /emails/send` – send a single email payload immediately.

Example payload:

```json
{
  "to": ["user@example.com"],
  "subject": "Welcome!",
  "body_text": "Plain text body",
  "body_html": "<p>HTML body</p>"
}
```
