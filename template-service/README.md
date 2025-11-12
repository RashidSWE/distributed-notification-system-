# Template Service

FastAPI service that renders ready-to-send email bodies using Jinja2 templates. Currently supports:

- `welcome_email` (HTML + text)
- `password_reset` (HTML + text)

## Requirements

- Python 3.11+
- Dependencies from `requirements.txt`

## Environment

Copy `.env.example` and set:

- `TEMPLATES_DIR` – absolute or relative path to template files (defaults to `templates`)
- `DEFAULT_LOCALE` – locale string used when requests omit one (defaults to `en`)

## Run

```bash
pip install -r requirements.txt
uvicorn app.main:app --reload --host 0.0.0.0 --port 8090
```

## API

- `GET /health`
- `POST /templates/render`

Example payload:

```json
{
  "template_key": "welcome_email",
  "format": "html",
  "context": {
    "user_name": "Jane",
    "product_name": "Notifications Hub"
  }
}
```

