import os
import requests
from dotenv import load_dotenv
from ..config import Settings, get_settings
from .email_service import EmailDeliveryError, EmailService
from ..schemas import EmailQueuePayload,UserProfile, RenderResponse, EmailResponse, EmailRequest

load_dotenv()
class EmailQueueService:
    def __init__(self, settings: Settings, email_service: EmailService | None = None) -> None:
        self._settings = settings
        self._email_service = email_service or EmailService(self._settings)

    async def process_queue_data(self, data: EmailQueuePayload) -> None:
        
        request_id = data.request_id
        metadata = data.metadata
        variables = data.variables
        user_profile = self.fetch_user_profile(data.user_id)
        context = self.prepare_context(variables.dict(), user_profile)
        template_data = self.fetch_render_template(data.template_code, context)
        email_request = {
            "request_id": request_id,
            "template_code": data.template_code,
            "to": [user_profile.email],
            "subject": template_data.subject,
            "body_html": template_data.content if template_data.format == "html" else None,
            "body_text": template_data.content if template_data.format == "text" else None,
            "headers": metadata or {},
            "metadata": metadata or {},
        }
        email_request = EmailRequest(**email_request)
        await self._email_service.send_email(email_request)
        return EmailResponse(request_id=request_id, status="sent")

    def fetch_user_profile(self, user_id: str) -> UserProfile:
        response = requests.post(f"{os.getenv('USER_SERVICE_URL')}/users/profile", json={"user_id": user_id})
        response.raise_for_status()
        user_data = response.json()
        return UserProfile(**user_data)

    def fetch_render_template(self, template_code: str, context: dict, format: str = "html") -> dict:
        response = requests.post(f"{os.getenv('TEMPLATE_SERVICE_ENDPOINT')}/templates/render",
                                 json={"template_key": template_code, "context": context, "format": format})
        response.raise_for_status()
        template_data = response.json()
        return RenderResponse(**template_data)

    def prepare_context(self, variables: dict, user_profile: UserProfile) -> dict:
        context = {
            "name": variables.get("name"),
            # "link": variables.get("link"),
            "meta": variables.get("meta", {}),
            "user_email": user_profile.email
        }
        return context
