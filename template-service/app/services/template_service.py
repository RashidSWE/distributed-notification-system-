from dataclasses import dataclass
from typing import Literal

from jinja2 import Environment, FileSystemLoader, TemplateNotFound, select_autoescape

from ..config import Settings
from ..schemas import RenderRequest, TemplateKey


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
