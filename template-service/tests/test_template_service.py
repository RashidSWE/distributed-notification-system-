import pytest

from app.config import Settings
from app.schemas import PushRenderRequest, RenderRequest, TemplateKey
from app.services import template_service
from app.services.template_service import (
    TemplateDefinition,
    TemplateFormatUnsupportedError,
    TemplateNotRegisteredError,
    TemplateService,
)


def build_settings() -> Settings:
    return Settings(TEMPLATES_DIR="templates")


def test_render_email_template_html():
    service = TemplateService(settings=build_settings())
    request = RenderRequest(template_key=TemplateKey.WELCOME_EMAIL, format="html", context={"user_name": "Jo"})

    subject, content = service.render(request)

    assert "Jo" in subject
    assert "Jo" in content


def test_render_email_template_invalid_format():
    service = TemplateService(settings=build_settings())
    original = template_service.TEMPLATE_REGISTRY[TemplateKey.WELCOME_EMAIL]
    template_service.TEMPLATE_REGISTRY[TemplateKey.WELCOME_EMAIL] = TemplateDefinition(
        key=TemplateKey.WELCOME_EMAIL,
        subject_template=original.subject_template,
        files={"html": original.files["html"]},
    )
    request = RenderRequest(template_key=TemplateKey.WELCOME_EMAIL, format="text")

    try:
        with pytest.raises(TemplateFormatUnsupportedError):
            service.render(request)
    finally:
        template_service.TEMPLATE_REGISTRY[TemplateKey.WELCOME_EMAIL] = original


def test_render_push_template_success():
    service = TemplateService(settings=build_settings())
    request = PushRenderRequest(
        template_code="PASSWORD_RESET_CODE",
        context={
            "reset_code": "999111",
            "expires_at": "2024-06-01T00:00:00Z",
            "request_time": "2024-05-31T23:00:00Z",
        },
    )

    response = service.render_push_template(request)

    assert response.data["reset_code"] == "999111"
    assert "999111" in response.body


def test_render_push_template_unknown():
    service = TemplateService(settings=build_settings())
    with pytest.raises(TemplateNotRegisteredError):
        service.render_push_template(PushRenderRequest(template_code="UNKNOWN"))
