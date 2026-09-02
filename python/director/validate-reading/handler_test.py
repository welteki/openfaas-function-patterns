import json
from types import SimpleNamespace

try:
    from . import handler as h
except ImportError:
    import handler as h


def event(body):
    return SimpleNamespace(body=json.dumps(body).encode())


def test_normalizes_valid_reading():
    response = h.handle(
        event({"device_id": " pump-17 ", "temperature_c": 48.2, "battery_percent": 78}),
        {},
    )

    assert response["statusCode"] == 200
    assert response["body"]["device_id"] == "pump-17"


def test_rejects_invalid_values():
    response = h.handle(
        event({"device_id": "pump-17", "temperature_c": 250, "battery_percent": 78}),
        {},
    )

    assert response["statusCode"] == 400
    assert "temperature_c" in response["body"]


def test_rejects_invalid_json():
    response = h.handle(SimpleNamespace(body=b"not-json"), {})
    assert response["statusCode"] == 400
