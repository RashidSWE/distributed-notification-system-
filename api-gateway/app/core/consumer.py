# app/core/consumer.py
import aio_pika
import asyncio
import json
from datetime import datetime
import logging
from app.core.db import AsyncSessionLocal, NotificationStatus
from app.core.redis_client import get_redis
from app.models.status import NotificationStatusUpdate
import asyncio
from sqlalchemy import insert


logger = logging.getLogger(__name__)
QUEUE_STATUS = "status.queue"

async def handle_message(message: aio_pika.IncomingMessage):
    async with message.process():
        try:
            data = json.loads(message.body)
            status_update = NotificationStatusUpdate.parse_obj(data)

            # Save to Postgres
            async with AsyncSessionLocal() as session:
                stmt = insert(NotificationStatus).values(
                    notification_id=status_update.notification_id,
                    status=status_update.status,
                    error=status_update.error,
                    updated_at=status_update.timestamp or datetime.utcnow()
                ).on_conflict_do_update(
                    index_elements=[NotificationStatus.notification_id],
                    set_={
                        "status": status_update.status,
                        "error": status_update.error,
                        "updated_at": status_update.timestamp or datetime.utcnow()
                    }
                )
                await session.execute(stmt)
                await session.commit()

            # Cache in Redis
            redis = await get_redis()
            key = f"notification_status:{status_update.notification_id}"
            await redis.hset(key, mapping={
                "status": status_update.status.value,
                "error": status_update.error or "",
                "updated_at": (status_update.timestamp or datetime.utcnow()).isoformat()
            })
            await redis.expire(key, 3600)

            logger.info(f"Updated status: {status_update.notification_id} -> {status_update.status}")

        except Exception as e:
            logger.error(f"Failed to process message: {e}")

async def start_status_consumer():
    from app.core.rabbitmq import get_connection

    conn = await get_connection()
    if not conn:
        logger.error("RabbitMQ unavailable")
        return

    channel = await conn.channel()
    queue = await channel.declare_queue(QUEUE_STATUS, durable=True)
    await queue.consume(handle_message)

    while True:
        await asyncio.sleep(1)
