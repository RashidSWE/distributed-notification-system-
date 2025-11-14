from __future__ import annotations

import asyncio
import contextlib
import logging

import aio_pika
from aio_pika import DeliveryMode, Message

from .config import Settings
from .schemas import DeliveryStatus

logger = logging.getLogger(__name__)


class StatusPublisher:
    """Publishes delivery outcomes to RabbitMQ queues."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._connection: aio_pika.RobustConnection | None = None
        self._channel: aio_pika.RobustChannel | None = None
        self._lock = asyncio.Lock()

    async def connect(self) -> None:
        async with self._lock:
            if self._connection and not self._connection.is_closed:
                return
            self._connection = await aio_pika.connect_robust(
                host=self._settings.rabbitmq_host,
                port=self._settings.rabbitmq_port,
                login=self._settings.rabbitmq_username,
                password=self._settings.rabbitmq_password,
                virtualhost=self._settings.rabbitmq_virtual_host,
            )
            self._channel = await self._connection.channel()
            logger.info("status publisher connected to rabbitmq")

    async def close(self) -> None:
        async with self._lock:
            if self._channel:
                with contextlib.suppress(Exception):
                    await self._channel.close()
                self._channel = None
            if self._connection:
                with contextlib.suppress(Exception):
                    await self._connection.close()
                self._connection = None

    async def publish_status(self, payload: DeliveryStatus) -> None:
        await self._publish(self._settings.rabbitmq_status_queue, payload)

    async def publish_failed(self, payload: DeliveryStatus) -> None:
        await self._publish(self._settings.rabbitmq_failed_queue, payload)

    async def _publish(self, queue_name: str, payload: DeliveryStatus) -> None:
        if not self._channel or self._channel.is_closed:
            await self.connect()
        assert self._channel

        message = Message(
            body=payload.model_dump_json().encode("utf-8"),
            delivery_mode=DeliveryMode.PERSISTENT,
            content_type="application/json",
        )
        try:
            await self._channel.default_exchange.publish(message, routing_key=queue_name)
            logger.info("delivery status published", extra={"queue": queue_name, "request_id": payload.request_id})
        except Exception as exc:  # pragma: no cover - network failure logging
            logger.exception("failed to publish delivery status: %s", exc)
