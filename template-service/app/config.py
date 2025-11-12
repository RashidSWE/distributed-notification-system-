from functools import lru_cache
from pathlib import Path

from pydantic import Field, field_validator
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    templates_dir: Path = Field(default=Path("templates"), alias="TEMPLATES_DIR")
    default_locale: str = Field(default="en", alias="DEFAULT_LOCALE")
    template_cache: bool = Field(default=True, alias="TEMPLATE_CACHE")

    model_config = {
        "env_file": ".env",
        "env_file_encoding": "utf-8",
        "case_sensitive": False,
        "populate_by_name": True,
    }

    @field_validator("templates_dir", mode="before")
    @classmethod
    def _expand_dir(cls, value: str | Path) -> Path:
        return Path(value).expanduser().resolve()


@lru_cache
def get_settings() -> Settings:
    return Settings()
