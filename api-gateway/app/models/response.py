from typing import Optional, Any, Generic, TypeVar
from pydantic import BaseModel

T = TypeVar("T")

class PaginationMeta(BaseModel):
    total: int = 0
    limit: int = 0
    page: int = 0
    total_pages: int = 0
    has_next: bool = False
    has_previous: bool = False

class APIResponse(BaseModel, Generic[T]):
    success: bool
    data: Optional[T] = None
    error: Optional[str] = None
    message: str
    meta: PaginationMeta = PaginationMeta()