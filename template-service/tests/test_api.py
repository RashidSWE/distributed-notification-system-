import pytest
from httpx import ASGITransport, AsyncClient

from app.main import app


def build_client():
    transport = ASGITransport(app=app)
    return AsyncClient(transport=transport, base_url="http://testserver")


@pytest.mark.asyncio
async def test_get_push_template():
    async with build_client() as client:
        response = await client.get("/templates/push/PASSWORD_RESET_CODE")

    assert response.status_code == 200
    assert response.json()["code"] == "PASSWORD_RESET_CODE"


@pytest.mark.asyncio
async def test_render_push_template_endpoint():
    payload = {
        "template_code": "WELCOME_EMAIL",
        "context": {
            "user_name": "Nia",
            "product_name": "Notifier",
            "request_time": "2024-05-10T00:00:00Z"
        },
    }

    async with build_client() as client:
        response = await client.post("/templates/push/render", json=payload)

    assert response.status_code == 200
    body = response.json()
    assert body["code"] == "WELCOME_EMAIL"
    assert "Nia" in body["body"]


@pytest.mark.asyncio
async def test_render_push_template_not_found():
    async with build_client() as client:
        response = await client.post(
            "/templates/push/render",
            json={"template_code": "UNKNOWN", "context": {}},
        )

    assert response.status_code == 404
