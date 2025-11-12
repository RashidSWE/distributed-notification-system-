from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    smtp_host: str = Field(..., alias="SMTP_HOST")
    smtp_port: int = Field(..., alias="SMTP_PORT")
    smtp_username: str = Field(..., alias="SMTP_USERNAME")
    smtp_password: str = Field(..., alias="SMTP_PASSWORD")
    smtp_use_tls: bool = Field(default=True, alias="SMTP_USE_TLS")
    smtp_use_ssl: bool = Field(default=False, alias="SMTP_USE_SSL")
    smtp_timeout_seconds: float = Field(default=30.0, alias="SMTP_TIMEOUT_SECONDS")
    email_from: str = Field(..., alias="EMAIL_FROM")
    email_max_retries: int = Field(default=3, alias="EMAIL_MAX_RETRIES")
    email_queue_enabled: bool = Field(default=False, alias="EMAIL_QUEUE_ENABLED")

    rabbitmq_host: str = Field(default="localhost", alias="RABBITMQ_HOST")
    rabbitmq_port: int = Field(default=5672, alias="RABBITMQ_PORT")
    rabbitmq_username: str = Field(default="guest", alias="RABBITMQ_USERNAME")
    rabbitmq_password: str = Field(default="guest", alias="RABBITMQ_PASSWORD")
    rabbitmq_virtual_host: str = Field(default="/", alias="RABBITMQ_VHOST")
    rabbitmq_queue_name: str = Field(default="email.notifications", alias="RABBITMQ_QUEUE_NAME")
    rabbitmq_prefetch_count: int = Field(default=10, alias="RABBITMQ_PREFETCH")
    rabbitmq_reconnect_delay_seconds: float = Field(default=5.0, alias="RABBITMQ_RECONNECT_DELAY")
    user_service_endpoint: str = Field(..., alias="USER_SERVICE_ENDPOINT")
    template_service_endpoint: str = Field(..., alias="TEMPLATE_SERVICE_ENDPOINT")

    model_config = {
        "case_sensitive": False,
        "env_file": ".env",
        "env_file_encoding": "utf-8",
    }


@lru_cache
def get_settings() -> Settings:
    return Settings()
