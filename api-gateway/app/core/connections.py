import aio_pika
import redis.asyncio as aioredis
from .config import *

async def test_rabbitmq_connection():
    try:
        connection = await aio_pika.connect_robust(
            f"amqp://{RABBITMQ_USER}:{RABBITMQ_PASS}@{RABBITMQ_HOST}:{RABBITMQ_PORT}/"
        )
        await connection.close()
        return True
    except Exception as e:
        print(f"RabbitMQ connection failed: {e}")
        return False


async def test_redis_connection():
    try:
        redis = aioredis.Redis(host=REDIS_HOST, port=REDIS_PORT)
        await redis.set("test_key", "ok")
        val = await redis.get("test_key")
        await redis.close()
        return val == b"ok"
    except Exception as e:
        print(f"Redis connection failed: {e}")
        return False