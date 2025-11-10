from fastapi import APIRouter
from app.core.connections import test_rabbitmq_connection, test_redis_connection

router = APIRouter()

@router.get("/health")
async def health_check():
    rabbit_ok = await test_rabbitmq_connection()
    redis_ok = await test_redis_connection()
    return {
        "rabbitmq": "connected" if rabbit_ok else "failed",
        "redis": "connected" if redis_ok else "failed"
    }
