import aio_pika
import json
import logging
from .config import *
import asyncio

logger = logging.getLogger(__name__)


EXCHANGE_NAME = "notifications.direct"
QUEUE_EMAIL = "email.queue"
QUEUE_PUSH = "push.queue"
QUEUE_FAILED = "failed.queue"

_rabbit_connection = None

async def get_connection(retries: int = 5, delay: int = 5):
    """
    Establish a robust connection wiith rabbitMQ
    """
    global _rabbit_connection

    if _rabbit_connection and not _rabbit_connection.is_closed:
        return _rabbit_connection
    for attempt in range(1, retries + 1):
        try:

            _rabbit_connection = await aio_pika.connect_robust(
                f"amqp://{RABBITMQ_USER}:{RABBITMQ_PASS}@{RABBITMQ_HOST}:{RABBITMQ_PORT}/"
            )
            logger.info("connected to RabbitMQ")
            return _rabbit_connection
        except Exception as e:
            logger.warning(f" RabbitMQ connection attempt {attempt} failed: {e}")
            await asyncio.sleep(delay)
    
    logger.error("failed to connect to RabbitMQ after several attempts")
    return None


async def setup_rabbitmq():
    """
    initializes RabbitMQ: declares exchange and binds queues.
    """

    conn = await get_connection()
    if not conn:
        logger.error("Could not initialze rabbitMQ(no connection)")
        return

    
    channel = await conn.channel()

    exchange = await channel.declare_exchange(EXCHANGE_NAME, aio_pika.ExchangeType.DIRECT, durable=True)

    await (await channel.declare_queue(QUEUE_EMAIL, durable=True)).bind(exchange, "email")
    await (await channel.declare_queue(QUEUE_PUSH, durable=True)).bind(exchange, "push")
    await (await channel.declare_queue(QUEUE_FAILED, durable=True)).bind(exchange, "failed")

    logger.info("RabbitMQ echange and queues initialized")
    return

async def publish_message(routing_keys, message: dict):
    """
    Publish message to RabbitMQ direct exchange 'notifications.direct'
    routing_key: email or push
    message: dict containing notification payload
    """

    try:
        conn = await get_connection()

        if not conn:
            logger.error("Cannot publish message: rabbitMq unavailable")

        channel = await conn.channel()
        exchange = await channel.get_exchange(EXCHANGE_NAME)

        if isinstance(routing_keys, str):
            routing_keys = [routing_keys]
        
        for key in routing_keys:
            await exchange.publish(
                aio_pika.Message(
                    body=json.dumps(message, default=str).encode(),
                    content_type="application/json",
                    delivery_mode=aio_pika.DeliveryMode.PERSISTENT
                ),

                routing_key=key
            )
        return True
    except Exception as e:
        print(f"Failed to publish message: {e}")
        return False