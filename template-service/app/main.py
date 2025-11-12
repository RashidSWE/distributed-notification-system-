from __future__ import annotations

from fastapi import Depends, FastAPI, status
from fastapi.responses import JSONResponse

from .config import Settings, get_settings
from .schemas import HealthResponse, PushTemplateResponse, RenderRequest, RenderResponse
from .services.template_service import (
    TemplateError,
    TemplateFormatUnsupportedError,
    TemplateNotRegisteredError,
    TemplateService,
    TemplateSourceMissingError,
)

app = FastAPI(title="Template Service", version="0.1.0")


def get_template_service(settings: Settings = Depends(get_settings)) -> TemplateService:
    return TemplateService(settings=settings)


@app.exception_handler(TemplateError)
async def template_error_handler(_, exc: TemplateError):
    status_code = status.HTTP_400_BAD_REQUEST
    if isinstance(exc, TemplateNotRegisteredError):
        status_code = status.HTTP_404_NOT_FOUND
    elif isinstance(exc, TemplateSourceMissingError):
        status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
    elif isinstance(exc, TemplateFormatUnsupportedError):
        status_code = status.HTTP_422_UNPROCESSABLE_ENTITY

    return JSONResponse(
        status_code=status_code,
        content={"detail": str(exc)},
    )


@app.get("/health", response_model=HealthResponse)
async def health() -> HealthResponse:
    return HealthResponse(status="ok")


@app.post("/templates/render", response_model=RenderResponse)
async def render_template(
    request: RenderRequest,
    service: TemplateService = Depends(get_template_service),
) -> RenderResponse:
    subject, content = service.render(request)
    return RenderResponse(
        template_key=request.template_key,
        format=request.format,
        subject=subject,
        content=content,
    )


@app.get("/templates/push/{template_code}", response_model=PushTemplateResponse)
async def get_push_template(
    template_code: str,
    service: TemplateService = Depends(get_template_service),
) -> PushTemplateResponse:
    return service.get_push_template(template_code)
