import contextlib
from email.message import EmailMessage
from typing import Iterable, Sequence

import aiosmtplib

from .config import Settings
from .schemas import EmailRequest
from loguru import logger



class ResolvedAttachment:
    def __init__(self, filename: str, content_type: str, data: bytes) -> None:
        self.filename = filename
        self.content_type = content_type
        self.data = data


class EmailSender:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings

    async def send_email(
        self,
        request: EmailRequest,
        request_id: str,
        attachments: Iterable[ResolvedAttachment] | None = None,
    ) -> None:
        resolved_attachments = list(attachments or [])
        logger.debug(f"Building email to send \nRequest_id: {request_id}\nAttachments: {len(resolved_attachments)}")
        message = self._build_message(request, request_id, resolved_attachments)
        recipients = self._collect_recipients(request)
        await self._deliver(message, recipients)

    def _build_message(
        self,
        request: EmailRequest,
        request_id: str,
        attachments: Iterable[ResolvedAttachment],
    ) -> EmailMessage:
        message = EmailMessage()
        message["Subject"] = request.subject
        message["From"] = self._settings.email_from
        message["To"] = ", ".join(request.to)
        if request.cc:
            message["Cc"] = ", ".join(request.cc)
        message["X-Request-ID"] = request_id

        for header, value in request.headers.items():
            message[header] = value

        if request.body_text:
            message.set_content(request.body_text)
        else:
            message.set_content("This message contains HTML content.")

        if request.body_html:
            message.add_alternative(request.body_html, subtype="html")

        for attachment in attachments:
            maintype, _, subtype = attachment.content_type.partition("/")
            if not subtype:
                maintype = "application"
                subtype = "octet-stream"
            message.add_attachment(
                attachment.data,
                maintype=maintype,
                subtype=subtype,
                filename=attachment.filename,
            )

        return message

    async def _deliver(self, message: EmailMessage, recipients: Sequence[str]) -> None:
        logger.debug(f"Connecting to SMTP server {self._settings.smtp_host}:{self._settings.smtp_port}")
        smtp = aiosmtplib.SMTP(
            hostname=self._settings.smtp_host,
            port=self._settings.smtp_port,
            use_tls=self._settings.smtp_use_tls,
            timeout=self._settings.smtp_timeout_seconds,
        )
        logger.debug("SMTP connection established.")

        try:
            await smtp.connect()
            # if self._settings.smtp_use_tls and not self._settings.smtp_use_ssl:
            #     await smtp.starttls()
            if self._settings.smtp_username:
                await smtp.login(self._settings.smtp_username, self._settings.smtp_password)
                logger.debug("SMTP authentication successful.")
            await smtp.send_message(message, recipients=recipients)
        finally:
            with contextlib.suppress(Exception):
                await smtp.quit()

    @staticmethod
    def _collect_recipients(request: EmailRequest) -> list[str]:
        recipients: list[str] = []
        recipients.extend(request.to)
        recipients.extend(request.cc)
        recipients.extend(request.bcc)
        return recipients
