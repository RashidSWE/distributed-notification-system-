from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, EmailStr
import httpx
import uuid
import os
from datetime import datetime

router = APIRouter()

USER_SERVICE_URL = os.getenv("USER_SERVICE_URL")
API_GATEWAY_URL = os.getenv("API_GATEWAY_URL")

class RegisterRequest(BaseModel):
    name: str
    email: EmailStr
    password: str

@router.post("/", status_code=status.HTTP_201_CREATED)
async def register_user(payload: RegisterRequest):
    # 1️⃣ Call User Service to register the user
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.post(f"{USER_SERVICE_URL}/api/users/register", json=payload.dict())
            if resp.status_code != 200:
                raise HTTPException(
                    status_code=resp.status_code,
                    detail=f"User Service registration failed: {resp.text}"
                )
            user_data = resp.json().get("data")
            print(user_data)
        except httpx.RequestError as e:
            raise HTTPException(status_code=503, detail=f"User Service unreachable: {e}")

    # 2️⃣ Call the existing notification endpoint in API Gateway
    notification_payload = {
        "notification_type": "email",
        "id": str(uuid.uuid4()),
        "user_id": user_data["id"],
        "template_code": "WELCOME_EMAIL",
        "variables": {
            "name": user_data["name"],
            "email": [user_data["email"]],
            "link": "https://example.com/welcome"
        },
        "request_id": str(uuid.uuid4())
    }

    routing_keys = ["email", "push"]
    async with httpx.AsyncClient() as client:
        for key in routing_keys:
            payload = notification_payload.copy()
            payload["notification_type"] = key
            
            print(payload)
            if key == "push":
                payload.update({
                    "device_tokens":["eeDkhnlpSbuDFfCEwB1fPf:APA91bFZcSTuNJJMB-UqdgHUwOsUP0FnvdYR16e0TGi6vrS494-YwH3T3dp_lEQaS1U3TkM13afI9paNpePHV7HmjbkB8y8mbeFy1vJeBSTyUVFn-iifwqU"],
                    "platform": "android",
                    "priority": "high",
                    "correlation_id": str(uuid.uuid4()),
                    "scheduled_at": datetime.utcnow().isoformat() + "Z",
                    "created_at": datetime.utcnow().isoformat() + "Z"
                })
                print(payload)
            try:
                notif_resp = await client.post(f"{API_GATEWAY_URL}/api/v1/notifications/", json=payload)
                print(notif_resp)
                if notif_resp.status_code != 202:
                    raise HTTPException(
                        status_code=notif_resp.status_code,
                        detail=f"Notification failed: {notif_resp.text}"
                    )
            except httpx.RequestError as e:
                raise HTTPException(status_code=503, detail=f"Notification service unreachable: {e}")

    return {
        "success": True,
        "data": user_data,
        "message": "User registered successfully and welcome notification enqueued"
    }
