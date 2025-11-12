from fastapi import FastAPI
from app.routes.health import router as health_router
from app.routes.notifications import router as notification_router
from app.core.rabbitmq import setup_rabbitmq

app = FastAPI(title="API Gateway", version="1.0.0")


@app.on_event("startup")
async def startup_event():
    await setup_rabbitmq()

app.include_router(notification_router, prefix="/api/v1/notifications")
app.include_router(health_router, prefix="/api", tags=["Health"])
