import json
import queue
import threading

from flask import Response, request


subscribers = set()
subscribers_lock = threading.Lock()
HEARTBEAT_INTERVAL = 15


def handle(req):
    if request.method == "GET":
        return subscribe()
    if request.method == "POST":
        return publish(req)

    return "method not allowed", 405, {"Allow": "GET, POST"}


def subscribe():
    messages = queue.Queue(maxsize=1)
    with subscribers_lock:
        subscribers.add(messages)

    def stream():
        try:
            yield ": connected\n\n"
            while True:
                try:
                    message = messages.get(timeout=HEARTBEAT_INTERVAL)
                    yield f"data: {message}\n\n"
                except queue.Empty:
                    # Periodic writes let the server detect idle disconnects.
                    yield ": keep-alive\n\n"
        finally:
            with subscribers_lock:
                subscribers.discard(messages)

    return Response(
        stream(),
        mimetype="text/event-stream",
        headers={"Cache-Control": "no-cache"},
    )


def publish(req):
    body = req.decode() if isinstance(req, bytes) else str(req)
    message = " ".join(body.splitlines()).strip()
    if not message:
        return "notification must not be empty", 400

    delivered = broadcast(message)
    return (
        json.dumps({"delivered": delivered}) + "\n",
        200,
        {"Content-Type": "application/json"},
    )


def broadcast(message):
    delivered = 0
    with subscribers_lock:
        for subscriber in subscribers:
            try:
                subscriber.put_nowait(message)
                delivered += 1
            except queue.Full:
                # Do not let a slow subscriber block every other client.
                pass

    return delivered
