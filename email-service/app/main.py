import logging

from fastapi import Depends, FastAPI, status
from fastapi.responses import JSONResponse

from .config import Settings, get_settings
from .queue_consumer import EmailQueueConsumer
from .schemas import EmailRequest, EmailResponse, HealthResponse
from .services.email_queue_service import EmailQueueService
from .services.email_service import EmailDeliveryError, EmailService
from .status_publisher import StatusPublisher

logging.basicConfig(level=logging.INFO)

app = FastAPI(title="Email Service", version="0.1.0")


def get_status_publisher() -> StatusPublisher | None:
    return getattr(app.state, "status_publisher", None)


def get_email_service(
    settings: Settings = Depends(get_settings),
    status_publisher: StatusPublisher | None = Depends(get_status_publisher),
) -> EmailService:
    return EmailService(settings=settings, status_publisher=status_publisher)


@app.exception_handler(EmailDeliveryError)
async def email_delivery_exception_handler(_, exc: EmailDeliveryError):
    return JSONResponse(
        status_code=status.HTTP_502_BAD_GATEWAY,
        content={"detail": str(exc)},
    )


@app.get("/health", response_model=HealthResponse)
async def health() -> HealthResponse:
    return HealthResponse(status="ok")


@app.post("/emails/send", response_model=EmailResponse, status_code=status.HTTP_202_ACCEPTED)
async def send_email(
    payload: EmailRequest,
    service: EmailService = Depends(get_email_service),
) -> EmailResponse:
    request_id = await service.send_email(payload)
    return EmailResponse(request_id=request_id, status="sent")

@app.on_event("startup")
async def start_queue_consumer() -> None:
    settings = get_settings()
    status_publisher = StatusPublisher(settings=settings)
    await status_publisher.connect()
    app.state.status_publisher = status_publisher

    if not settings.email_queue_enabled:
        return

    email_service = EmailService(settings=settings, status_publisher=status_publisher)
    queue_service = EmailQueueService(settings=settings, email_service=email_service)
    consumer = EmailQueueConsumer(settings=settings, email_queue_service=queue_service)
    await consumer.start()
    app.state.email_queue_consumer = consumer


@app.on_event("shutdown")
async def stop_queue_consumer() -> None:
    consumer: EmailQueueConsumer | None = getattr(app.state, "email_queue_consumer", None)
    if consumer:
        await consumer.stop()
    publisher: StatusPublisher | None = getattr(app.state, "status_publisher", None)
    if publisher:
        await publisher.close()
