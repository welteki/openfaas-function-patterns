import json
from types import SimpleNamespace

import requests

try:
    from . import handler as h
except ImportError:
    import handler as h


class FakeResponse:
    def __init__(self, status_code, body):
        self.status_code = status_code
        self._body = body
        self.content = json.dumps(body).encode() if isinstance(body, dict) else body.encode()
        self.text = self.content.decode()

    def json(self):
        return json.loads(self.content)


def event(body):
    return SimpleNamespace(body=json.dumps(body).encode())


def test_combines_parallel_results(monkeypatch):
    responses = {
        "validate-reading": FakeResponse(
            200,
            {"device_id": "pump-17", "temperature_c": 82.4, "battery_percent": 12},
        ),
        "temperature-check": FakeResponse(
            200,
            {"value_c": 82.4, "threshold_c": 75.0, "alert": True},
        ),
        "battery-check": FakeResponse(
            200,
            {"value_percent": 12, "threshold_percent": 20, "alert": True},
        ),
    }
    monkeypatch.setattr(h, "invoke", lambda name, body: responses[name])

    response = h.handle(
        event({"device_id": "pump-17", "temperature_c": 82.4, "battery_percent": 12}),
        {},
    )

    assert response["statusCode"] == 200
    assert response["body"]["status"] == "alert"
    assert response["body"]["temperature"]["alert"] is True
    assert response["body"]["battery"]["alert"] is True


def test_stops_after_validation_failure(monkeypatch):
    calls = []

    def fake_invoke(name, body):
        calls.append(name)
        return FakeResponse(400, "device_id is required")

    monkeypatch.setattr(h, "invoke", fake_invoke)
    response = h.handle(event({}), {})

    assert response["statusCode"] == 400
    assert calls == ["validate-reading"]


def test_reports_transport_failure(monkeypatch):
    def fake_invoke(name, body):
        raise requests.ConnectionError("gateway unavailable")

    monkeypatch.setattr(h, "invoke", fake_invoke)
    response = h.handle(event({"device_id": "pump-17"}), {})

    assert response["statusCode"] == 502
    assert "validate-reading" in response["body"]
