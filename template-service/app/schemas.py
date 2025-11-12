from enum import Enum
from typing import Any, Literal

from pydantic import BaseModel, Field


class TemplateKey(str, Enum):
    WELCOME_EMAIL = "welcome_email"
    PASSWORD_RESET = "password_reset"


class RenderRequest(BaseModel):
    template_key: TemplateKey
    format: Literal["html", "text"] = "html"
    context: dict[str, Any] = Field(default_factory=dict)
    locale: str | None = None


class RenderResponse(BaseModel):
    template_key: TemplateKey
    format: Literal["html", "text"]
    subject: str
    content: str


class HealthResponse(BaseModel):
    status: Literal["ok"]
