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


class PushTemplateResponse(BaseModel):
    code: str
    name: str
    description: str | None = None
    category: str | None = None
    type: str | None = None
    title: str | None = None
    body: str | None = None
    image_url: str | None = None
    icon_url: str | None = None
    data: dict[str, Any] = Field(default_factory=dict)
    color: str | None = None
    sound: str | None = None
    badge: int | None = None
    priority: int | None = None


class HealthResponse(BaseModel):
    status: Literal["ok"]
