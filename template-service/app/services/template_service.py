from dataclasses import asdict, dataclass, field
from typing import Literal

from jinja2 import Environment, FileSystemLoader, TemplateNotFound, select_autoescape

from ..config import Settings
from ..schemas import PushTemplateResponse, RenderRequest, TemplateKey


class TemplateError(RuntimeError):
    """Base class for template rendering issues."""


class TemplateNotRegisteredError(TemplateError):
    """Template key is unknown."""


class TemplateFormatUnsupportedError(TemplateError):
    """Requested format is not available for template."""


class TemplateSourceMissingError(TemplateError):
    """Template file is missing on disk."""


@dataclass(frozen=True)
class TemplateDefinition:
    key: TemplateKey
    subject_template: str
    files: dict[Literal["html", "text"], str]


@dataclass(frozen=True)
class PushTemplateDefinition:
    code: str
    name: str
    description: str
    category: str
    type: str
    title: str
    body: str
    image_url: str | None = None
    icon_url: str | None = None
    data: dict[str, str] = field(default_factory=dict)
    color: str | None = None
    sound: str | None = None
    badge: int | None = None
    priority: int | None = None


TEMPLATE_REGISTRY: dict[TemplateKey, TemplateDefinition] = {
    TemplateKey.WELCOME_EMAIL: TemplateDefinition(
        key=TemplateKey.WELCOME_EMAIL,
        subject_template="Welcome to Notifications Hub, {{ user_name | default('there') }}!",
        files={
            "html": "welcome_email.html",
            "text": "welcome_email.txt",
        },
    ),
    TemplateKey.PASSWORD_RESET: TemplateDefinition(
        key=TemplateKey.PASSWORD_RESET,
        subject_template="Reset your password",
        files={
            "html": "password_reset.html",
            "text": "password_reset.txt",
        },
    ),
}


PUSH_TEMPLATE_REGISTRY: dict[str, PushTemplateDefinition] = {
    "PASSWORD_RESET_CODE": PushTemplateDefinition(
        code="PASSWORD_RESET_CODE",
        name="Password Reset Verification Code",
        description="Push notification guiding the user through password recovery.",
        category="security",
        type="push",
        title="Verify your password reset",
        body="Use code {{reset_code}} to finish resetting your password. It expires at {{expires_at}}.",
        image_url=None,
        icon_url=None,
        data={
            "type": "security_alert",
            "alert_type": "reset_code",
            "reset_code": "{{reset_code}}",
            "expires_at": "{{expires_at}}",
            "request_time": "{{request_time}}",
            "action_type": "deeplink",
            "action_url": "myapp://reset-password",
        },
        color="#FF5722",
        sound="notification.mp3",
        badge=1,
        priority=10,
    ),
    "WELCOME_EMAIL": PushTemplateDefinition(
        code="WELCOME_EMAIL",
        name="Welcome Notification",
        description="Greets new users right after signup.",
        category="engagement",
        type="push",
        title="Welcome to {{product_name | default('Notifications Hub')}}",
        body="Hey {{user_name | default('there')}}, tap to explore your new workspace.",
        image_url=None,
        icon_url=None,
        data={
            "type": "onboarding",
            "alert_type": "welcome",
            "request_time": "{{request_time}}",
            "action_type": "deeplink",
            "action_url": "myapp://home",
        },
        color="#2E7D32",
        sound="notification.mp3",
        badge=1,
        priority=5,
    ),
}


class TemplateService:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        loader = FileSystemLoader(searchpath=str(settings.templates_dir))
        self._env = Environment(
            loader=loader,
            autoescape=select_autoescape(["html", "xml"]),
            enable_async=False,
        )
        if not settings.template_cache:
            self._env.cache = {} 
        self._subject_env = Environment(autoescape=False)

    def render(self, request: RenderRequest) -> tuple[str, str]:
        definition = TEMPLATE_REGISTRY.get(request.template_key)
        if not definition:
            raise TemplateNotRegisteredError(f"template {request.template_key.value} is not registered.")

        template_name = definition.files.get(request.format)
        if not template_name:
            raise TemplateFormatUnsupportedError(
                f"template {request.template_key.value} does not support format {request.format}."
            )

        render_context = {
            **request.context,
            "locale": request.locale or self._settings.default_locale,
        }

        try:
            template = self._env.get_template(template_name)
        except TemplateNotFound as exc:
            raise TemplateSourceMissingError(f"template source {template_name} is missing.") from exc

        content = template.render(**render_context)
        subject = self._subject_env.from_string(definition.subject_template).render(**render_context)
        return subject.strip(), content

    def get_push_template(self, template_code: str) -> PushTemplateResponse:
        definition = PUSH_TEMPLATE_REGISTRY.get(template_code)
        if not definition:
            raise TemplateNotRegisteredError(f"push template {template_code} is not registered.")
        return PushTemplateResponse(**asdict(definition))
