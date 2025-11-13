from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, EmailStr
import httpx
import uuid
import os

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
            if resp.status_code != 201:
                raise HTTPException(
                    status_code=resp.status_code,
                    detail=f"User Service registration failed: {resp.text}"
                )
            user_data = resp.json()
        except httpx.RequestError as e:
            raise HTTPException(status_code=503, detail=f"User Service unreachable: {e}")

    # 2️⃣ Call the existing notification endpoint in API Gateway
    notification_payload = {
        "notification_type": "email",
        "notification_id": str(uuid.uuid4()),
        "user_id": user_data["id"],
        "template_code": "WELCOME_EMAIL",
        "variables": {
            "name": user_data["name"],
            "link": "https://example.com/welcome"
        },
        "request_id": str(uuid.uuid4())
    }

    async with httpx.AsyncClient() as client:
        try:
            notif_resp = await client.post(f"{API_GATEWAY_URL}/api/v1/notifications", json=notification_payload,
                                           headers={"Authorization": f"Bearer {user_data['token']}"})
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
