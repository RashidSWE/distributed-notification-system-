from fastapi import APIRouter, HTTPException, status, Depends
from app.models.notification import NotificationRequest, NotificationType

from app.core.rabbitmq import publish_message
from app.models.response import APIResponse
from app.core.auth import verify_jwt_token, get_current_user

router = APIRouter()

@router.post("/", response_model=APIResponse[dict], status_code=status.HTTP_202_ACCEPTED)
async def create_notification(payload: NotificationRequest, current_user: dict = Depends(get_current_user)):
    """
    Accepts a notification request and publishes it to RabbitMQ
    """

    if payload.notification_type not in [NotificationType.email, NotificationType.push]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="invalid notification type. Must be email or push"
        )
    
    payload.user_id = current_user["id"]
    email = current_user["email"]


    message_dict = payload.model_dump()
    success = await publish_message(payload.notification_type.value, message_dict)

    if not success:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail="Failed to enqueue message")
    
    
    return APIResponse(
        success= True,
        data={
            "notification_type": payload.notification_type,
            "user_id": payload.user_id,
            "template_code": payload.template_code,
            "email": email
        },
        message="Notification enqueued successfully"
    )