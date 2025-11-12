from __future__ import annotations

import asyncio
import contextlib
import logging
from typing import Any

import aio_pika
from aio_pika.abc import AbstractIncomingMessage
from pydantic import ValidationError

from .config import Settings
from .schemas import EmailQueuePayload
from .services.email_service import EmailDeliveryError
from .services.email_queue_service import EmailQueueService

logger = logging.getLogger(__name__)


class EmailQueueConsumer:
    """Consumes email payloads from RabbitMQ and delegates to EmailService."""

    def __init__(self, settings: Settings, email_queue_service: EmailQueueService) -> None:
        self._settings = settings
        self._email_queue_service = email_queue_service
        self._connection: aio_pika.RobustConnection | None = None
        self._channel: aio_pika.RobustChannel | None = None
        self._consumer_tag: str | None = None
        self._task: asyncio.Task[Any] | None = None
        self._stopping = asyncio.Event()

    async def start(self) -> None:
        if self._task and not self._task.done():
            return
        self._stopping.clear()
        self._task = asyncio.create_task(self._run(), name="email-queue-consumer")
        logger.info("email queue consumer starting")

    async def stop(self) -> None:
        self._stopping.set()
        if self._task:
            await self._task
            self._task = None
        await self._cleanup()
        logger.info("email queue consumer stopped")

    async def _run(self) -> None:
        while not self._stopping.is_set():
            try:
                await self._connect()
                await self._stopping.wait()
                break
            except asyncio.CancelledError:
                raise
            except Exception as exc:
                logger.exception(f"email queue consumer error: {exc}")
                await self._cleanup()
                await asyncio.sleep(self._settings.rabbitmq_reconnect_delay_seconds)
        await self._cleanup()

    async def _connect(self) -> None:
        self._connection = await aio_pika.connect_robust(
            host=self._settings.rabbitmq_host,
            port=self._settings.rabbitmq_port,
            login=self._settings.rabbitmq_username,
            password=self._settings.rabbitmq_password,
            virtualhost=self._settings.rabbitmq_virtual_host,
        )
        self._channel = await self._connection.channel()
        await self._channel.set_qos(prefetch_count=self._settings.rabbitmq_prefetch_count)
        queue = await self._channel.declare_queue(
            self._settings.rabbitmq_queue_name,
            durable=True,
        )
        self._consumer_tag = await queue.consume(self._handle_message, no_ack=False)
        logger.info(f"Connected to rabbitmq queue. Queue: {self._settings.rabbitmq_queue_name}")

    async def _cleanup(self) -> None:
        if self._channel and self._consumer_tag:
            with contextlib.suppress(Exception):
                await self._channel.cancel(self._consumer_tag)
        if self._channel:
            with contextlib.suppress(Exception):
                await self._channel.close()
        if self._connection:
            with contextlib.suppress(Exception):
                await self._connection.close()
        self._connection = None
        self._channel = None
        self._consumer_tag = None

    async def _handle_message(self, message: AbstractIncomingMessage) -> None:
        try:
            email_request = EmailQueuePayload.model_validate_json(message.body.decode())
        except ValidationError as exc:
            logger.warning(f"invalid email payload dropped: {exc}")
            await message.reject(requeue=False)
            return

        try:
            await self._email_queue_service.process_queue_data(email_request)
        except EmailDeliveryError as exc:
            logger.warning(f"email delivery error, requeueing: {exc}")
            await message.nack(requeue=True)
        except Exception as exc:  # pragma: no cover - safety net
            logger.exception(f"unexpected error processing email: {exc}")
            await message.nack(requeue=False)
        else:
            await message.ack()
