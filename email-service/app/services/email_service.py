import asyncio
import base64
import binascii
import uuid
from datetime import datetime
from typing import Iterable

import httpx
from loguru import logger

from ..config import Settings
from ..sender import EmailSender, ResolvedAttachment
from ..schemas import Attachment, DeliveryStatus, EmailRequest
from ..status_publisher import StatusPublisher

class EmailDeliveryError(RuntimeError):
    """Raised when the service cannot deliver an email."""


class EmailService:
    def __init__(
        self,
        settings: Settings,
        email_sender: EmailSender | None = None,
        status_publisher: StatusPublisher | None = None,
    ) -> None:
        self._settings = settings
        self._email_sender = email_sender or EmailSender(settings)
        self._status_publisher = status_publisher

    async def send_email(self, payload: EmailRequest) -> str:
        if not payload.request_id:
            payload.request_id = str(uuid.uuid4())
        request_id = payload.request_id
        attachments = await self._resolve_attachments(payload.attachments)
        retry_attempts = max(self._settings.email_max_retries, 1)
        last_error: Exception | None = None

        for attempt in range(1, retry_attempts + 1):
            try:
                await self._email_sender.send_email(
                    request=payload,
                    request_id=request_id,
                    attachments=attachments,
                )
                logger.info(f"Email sent \n\nRequest_id: {request_id}\nAttempt: {attempt}")
                await self._publish_status(payload, attempt, status="sent")
                return request_id
            except Exception as exc:
                last_error = exc
                logger.warning(f"Email send failed\n\nRequest_id: {request_id}\nAttempt: {attempt}\nError: {exc}")
                if attempt < retry_attempts:
                    await asyncio.sleep(min(2 ** (attempt - 1), 5))

        message = "Failed to send email after retries."
        if last_error:
            message = f"{message} Last error: {last_error}"
        await self._publish_status(payload, retry_attempts, status="failed", error=str(last_error) if last_error else None, notify_failed=True)
        raise EmailDeliveryError(message)

    async def _resolve_attachments(
        self,
        attachments: Iterable[Attachment],
    ) -> list[ResolvedAttachment]:
        tasks = [self._resolve_attachment(attachment) for attachment in attachments]
        return await asyncio.gather(*tasks)

    async def _resolve_attachment(self, attachment: Attachment) -> ResolvedAttachment:
        if attachment.data_base64:
            return ResolvedAttachment(
                filename=attachment.filename,
                content_type=attachment.content_type,
                data=self._decode_base64(attachment.data_base64),
            )
        if attachment.remote_url:
            data = await self._download_remote(attachment.remote_url)
            return ResolvedAttachment(
                filename=attachment.filename,
                content_type=attachment.content_type,
                data=data,
            )
        raise ValueError("Attachment requires data_base64 or remote_url.")

    async def _download_remote(self, url: str) -> bytes:
        timeout = httpx.Timeout(self._settings.smtp_timeout_seconds)
        async with httpx.AsyncClient(timeout=timeout, follow_redirects=True) as client:
            response = await client.get(url)
            response.raise_for_status()
            return response.content

    @staticmethod
    def _decode_base64(data: str) -> bytes:
        try:
            return base64.b64decode(data, validate=True)
        except (ValueError, binascii.Error) as exc:
            raise ValueError("Invalid base64 attachment payload.") from exc

    async def _publish_status(
        self,
        payload: EmailRequest,
        attempts: int,
        *,
        status: str,
        error: str | None = None,
        notify_failed: bool = False,
    ) -> None:
        if not self._status_publisher:
            return
        status_payload = DeliveryStatus(
            request_id=payload.request_id or "",
            status=status,  # type: ignore[arg-type]
            recipients=self._collect_recipients(payload),
            template_code=payload.template_code,
            metadata=payload.metadata,
            error=error,
            attempts=attempts,
            sent_at=datetime.utcnow(),
        )
        try:
            await self._status_publisher.publish_status(status_payload)
            if status == "failed" or notify_failed:
                await self._status_publisher.publish_failed(status_payload)
        except Exception as exc:  # pragma: no cover - publishing is best effort
            logger.error(f"Failed to publish delivery status: {exc}")

    @staticmethod
    def _collect_recipients(payload: EmailRequest) -> list[str]:
        recipients: list[str] = []
        recipients.extend(payload.to)
        recipients.extend(payload.cc)
        recipients.extend(payload.bcc)
        return recipients
