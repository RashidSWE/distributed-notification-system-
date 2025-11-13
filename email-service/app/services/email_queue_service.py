import requests

from ..config import Settings
from ..schemas import EmailQueuePayload, EmailRequest, EmailResponse, RenderResponse
from .email_service import EmailService


class EmailQueueService:
    def __init__(self, settings: Settings, email_service: EmailService | None = None) -> None:
        self._settings = settings
        self._email_service = email_service or EmailService(settings)

    async def process_queue_data(self, data: EmailQueuePayload) -> None:
        context = self._build_context(data)
        template = self._fetch_render_template(data.template_code, context)

        email_request = EmailRequest(
            request_id=data.request_id,
            template_code=data.template_code,
            to=[data.data.email],
            subject=template.subject,
            body_html=template.content if template.format == "html" else None,
            body_text=template.content if template.format == "text" else None,
            metadata={"user_id": data.user_id},
        )

        await self._email_service.send_email(email_request)
        return EmailResponse(request_id=data.request_id, status="sent")

    def _fetch_render_template(self, template_code: str, context: dict, format: str = "html") -> RenderResponse:
        response = requests.post(
            f"{self._settings.template_service_endpoint}/templates/render",
            json={"template_key": template_code, "context": context, "format": format},
            timeout=10,
        )
        response.raise_for_status()
        return RenderResponse(**response.json())

    def _build_context(self, data: EmailQueuePayload) -> dict:
        return {
            "name": data.data.name,
            "email": data.data.email,
            **data.data.context,
        }
