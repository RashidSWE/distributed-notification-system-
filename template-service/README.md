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
- `GET /templates/push/{template_code}` – returns push notification metadata (e.g., `PASSWORD_RESET_CODE`)
- `POST /templates/push/render` – renders a push template with the supplied context and returns the substituted payload

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

Example push response (`GET /templates/push/PASSWORD_RESET_CODE`):

```json
{
  "code": "PASSWORD_RESET_CODE",
  "name": "Password Reset Verification Code",
  "description": "Push notification guiding the user through password recovery.",
  "category": "security",
  "type": "push",
  "title": "Verify your password reset",
  "body": "Use code {{reset_code}} to finish resetting your password. It expires at {{expires_at}}.",
  "image_url": null,
  "icon_url": null,
  "data": {
    "type": "security_alert",
    "alert_type": "reset_code",
    "reset_code": "{{reset_code}}",
    "expires_at": "{{expires_at}}",
    "request_time": "{{request_time}}",
    "action_type": "deeplink",
    "action_url": "myapp://reset-password"
  },
  "color": "#FF5722",
  "sound": "notification.mp3",
  "badge": 1,
  "priority": 10
}
```

Example push render request/response (`POST /templates/push/render`):

```jsonc
// Request
{
  "template_code": "PASSWORD_RESET_CODE",
  "context": {
    "reset_code": "483921",
    "expires_at": "2024-05-12T14:30:00Z",
    "request_time": "2024-05-12T14:00:00Z"
  }
}

// Response (fields substituted)
{
  "code": "PASSWORD_RESET_CODE",
  "title": "Verify your password reset",
  "body": "Use code 483921 to finish resetting your password. It expires at 2024-05-12T14:30:00Z.",
  "data": {
    "type": "security_alert",
    "alert_type": "reset_code",
    "reset_code": "483921",
    "expires_at": "2024-05-12T14:30:00Z",
    "request_time": "2024-05-12T14:00:00Z",
    "action_type": "deeplink",
    "action_url": "myapp://reset-password"
  }
}
```

Example welcome push (`GET /templates/push/WELCOME_EMAIL`):

```json
{
  "code": "WELCOME_EMAIL",
  "name": "Welcome Notification",
  "description": "Greets new users right after signup.",
  "category": "engagement",
  "type": "push",
  "title": "Welcome to {{product_name | default('Notifications Hub')}}",
  "body": "Hey {{user_name | default('there')}}, tap to explore your new workspace.",
  "image_url": null,
  "icon_url": null,
  "data": {
    "type": "onboarding",
    "alert_type": "welcome",
    "request_time": "{{request_time}}",
    "action_type": "deeplink",
    "action_url": "myapp://home"
  },
  "color": "#2E7D32",
  "sound": "notification.mp3",
  "badge": 1,
  "priority": 5
}
```
