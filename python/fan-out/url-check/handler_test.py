from io import BytesIO
from types import SimpleNamespace

import requests

try:
    from . import handler as h
except ImportError:
    import handler as h


class FakeResponse:
    def __init__(self, status_code, content_type="text/plain"):
        self.status_code = status_code
        self.headers = {"Content-Type": content_type}
        self.raw = BytesIO(b"response body")
        self.closed = False

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.closed = True


def event(target):
    return SimpleNamespace(body=target.encode())


def test_reports_healthy_url(monkeypatch):
    fetched = FakeResponse(204)
    monkeypatch.setattr(h.requests, "get", lambda *args, **kwargs: fetched)
    response = h.handle(event("https://www.openfaas.com/"), {})

    assert response["statusCode"] == 200
    assert response["body"]["reachable"] is True
    assert response["body"]["healthy"] is True
    assert fetched.closed is True


def test_reports_unhealthy_status(monkeypatch):
    monkeypatch.setattr(h.requests, "get", lambda *args, **kwargs: FakeResponse(503))
    response = h.handle(event("https://one.example"), {})

    assert response["body"]["reachable"] is True
    assert response["body"]["healthy"] is False


def test_reports_network_error(monkeypatch):
    def fail(*args, **kwargs):
        raise requests.Timeout("request timed out")

    monkeypatch.setattr(h.requests, "get", fail)
    response = h.handle(event("https://one.example"), {})

    assert response["statusCode"] == 200
    assert response["body"]["reachable"] is False
    assert "timed out" in response["body"]["error"]


def test_rejects_invalid_url():
    response = h.handle(event("file:///etc/passwd"), {})
    assert response["statusCode"] == 400
