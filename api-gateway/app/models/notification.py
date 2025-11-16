from pydantic import BaseModel, Field, HttpUrl
from typing import Optional, Dict
from enum import Enum
import uuid


class NotificationType(str, Enum):
    email = "email"
    push = "push"


class UserData(BaseModel):
    name: str
    link: HttpUrl
    meta: Optional[Dict] = None


class NotificationRequest(BaseModel):
    notification_type: NotificationType
    user_id: Optional[str] = None
    template_code: str
    variables: UserData
    request_id: str
    priority: Optional[str] = None
    metadata: Optional[Dict] = None

