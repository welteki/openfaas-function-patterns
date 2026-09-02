import json
import queue

from flask import Flask

try:
    from . import handler as h
except ImportError:
    import handler as h


app = Flask(__name__)


def setup_function():
    with h.subscribers_lock:
        h.subscribers.clear()


def test_publish_broadcasts_to_subscribers():
    messages = queue.Queue(maxsize=1)
    with h.subscribers_lock:
        h.subscribers.add(messages)

    with app.test_request_context(method="POST"):
        body, status, headers = h.handle("deployment complete")

    assert status == 200
    assert headers["Content-Type"] == "application/json"
    assert json.loads(body)["delivered"] == 1
    assert messages.get_nowait() == "deployment complete"


def test_subscribe_registers_and_removes_client():
    with app.test_request_context(method="GET"):
        response = h.handle("")
        stream = response.response
        assert next(stream) == ": connected\n\n"
        assert len(h.subscribers) == 1
        stream.close()

    assert len(h.subscribers) == 0


def test_subscribe_sends_heartbeat_while_idle(monkeypatch):
    monkeypatch.setattr(h, "HEARTBEAT_INTERVAL", 0)

    with app.test_request_context(method="GET"):
        response = h.handle("")
        stream = response.response
        assert next(stream) == ": connected\n\n"
        assert next(stream) == ": keep-alive\n\n"
        stream.close()

    assert len(h.subscribers) == 0


def test_rejects_empty_notification():
    with app.test_request_context(method="POST"):
        body, status = h.handle("\n")

    assert status == 400
    assert "must not be empty" in body
