from fastapi import APIRouter, HTTPException, status, Depends
from app.models.notification import NotificationRequest, NotificationType

from app.core.rabbitmq import publish_message
from app.models.response import APIResponse
from app.core.auth import verify_jwt_token, get_current_user
import uuid
import logging
from datetime import datetime


logger = logging.getLogger(__name__)

router = APIRouter()

@router.post("/", response_model=APIResponse[dict], status_code=status.HTTP_202_ACCEPTED)
async def create_notification(payload: NotificationRequest):
    """
    Accepts a notification request and publishes it to RabbitMQ
    """
    notification_id = str(uuid.uuid4())

    if payload.notification_type not in [NotificationType.email, NotificationType.push]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="invalid notification type. Must be email or push"
        )
    
    payload.user_id = current_user["id"]
    email = current_user["email"]

    message_dict = payload.model_dump()
    message_dict["notification_id"] = notification_id

    # Publish to both queues
    routing_keys = ["email", "push"]  # always send to both
    success = True
    for key in routing_keys:
        ok = await publish_message(key, message_dict)
        if not ok:
            logger.error(f"Failed to publish message to {key} queue")
            success = False

    if not success:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail="Failed to enqueue message")


    return APIResponse(
        success= True,
        data={
            "notification_id": notification_id,
            "notification_type": payload.notification_type,
            "user_id": payload.user_id,
            "template_code": payload.template_code,
        },
        message="Notification enqueued successfully"
    )