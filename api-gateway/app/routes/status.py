from pydantic import BaseModel
from typing import Optional
from enum import Enum
from datetime import datetime

class NotificationStatus(str, Enum):
    delivered = "delivered"
    pending = "pending"
    failed = "failed"

class NotificationStatusUpdate(BaseModel):
    notification_id: str
    status: NotificationStatus
    timestamp: Optional[datetime] = None
    error: Optional[str] = None
