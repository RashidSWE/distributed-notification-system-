from datetime import datetime
from enum import Enum
from typing import Any, Literal

from pydantic import BaseModel, EmailStr, Field, HttpUrl, model_validator


class Attachment(BaseModel):
    filename: str
    content_type: str = Field(default="application/octet-stream")
    data_base64: str | None = None
    remote_url: HttpUrl | None = None

    @model_validator(mode="after")
    def validate_payload(self) -> "Attachment":
        if not (self.data_base64 or self.remote_url):
            raise ValueError("Either data_base64 or remote_url must be provided for an attachment.")
        return self


class EmailRequest(BaseModel):
    request_id: str | None = None
    template_code: str | None = None
    to: list[EmailStr] = Field(default_factory=list)
    cc: list[EmailStr] = Field(default_factory=list)
    bcc: list[EmailStr] = Field(default_factory=list)
    subject: str
    body_text: str | None = None
    body_html: str | None = None
    headers: dict[str, str] = Field(default_factory=dict)
    attachments: list[Attachment] = Field(default_factory=list)
    metadata: dict[str, str] | None = None

    @model_validator(mode="after")
    def validate_payload(self) -> "EmailRequest":
        if not self.to:
            raise ValueError("At least one primary recipient is required.")
        if not (self.body_text or self.body_html):
            raise ValueError("Email body requires body_text and/or body_html.")
        return self


class EmailResponse(BaseModel):
    request_id: str
    status: Literal["queued", "sent"]


class HealthResponse(BaseModel):
    status: Literal["ok"]


class EmailQueueData(BaseModel):
    notification_type: Literal["email"]
    user_id: str
    template_code: str
    request_id: str
    name: str
    email: EmailStr
    context: dict[str, Any] = Field(default_factory=dict)


class EmailQueueEnvelope(BaseModel):
    success: bool
    data: EmailQueueData
    error: Any | None = None
    message: str | None = None
    meta: dict[str, Any] | None = None


class TemplateKey(str, Enum):
    WELCOME_EMAIL = "welcome_email"
    PASSWORD_RESET = "password_reset"


class RenderResponse(BaseModel):
    template_key: TemplateKey
    format: Literal["html", "text"]
    subject: str
    content: str


class DeliveryStatus(BaseModel):
    request_id: str
    status: Literal["sent", "failed"]
    recipients: list[EmailStr]
    template_code: str | None = None
    metadata: dict[str, str] | None = None
    error: str | None = None
    attempts: int = 1
    sent_at: datetime | None = None
