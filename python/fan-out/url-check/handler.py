import os
import time
from urllib.parse import urlparse

import requests


MAX_URL_LENGTH = 4096
REQUEST_TIMEOUT = float(os.getenv("request_timeout", "5"))
if REQUEST_TIMEOUT <= 0:
    raise ValueError("request_timeout must be greater than zero")


def handle(event, context):
    body = (
        event.body
        if isinstance(event.body, bytes)
        else str(event.body).encode()
    )
    if len(body) > MAX_URL_LENGTH:
        return error("URL is too long")

    target = body.decode().strip()
    parsed = urlparse(target)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        return error("expected an absolute HTTP or HTTPS URL")

    started = time.monotonic()
    result = {
        "url": target,
        "reachable": False,
        "healthy": False,
    }

    try:
        with requests.get(
            target,
            headers={"User-Agent": "OpenFaaS URL health check"},
            timeout=REQUEST_TIMEOUT,
            stream=True,
        ) as response:
            response.raw.read(1024)
            result.update(
                {
                    "reachable": True,
                    "healthy": 200 <= response.status_code < 400,
                    "status_code": response.status_code,
                    "content_type": response.headers.get("Content-Type", ""),
                    "duration_ms": int((time.monotonic() - started) * 1000),
                }
            )
    except requests.RequestException as err:
        result["duration_ms"] = int((time.monotonic() - started) * 1000)
        result["error"] = str(err)
        return {"statusCode": 200, "body": result}

    return {"statusCode": 200, "body": result}


def error(message):
    return {"statusCode": 400, "body": message}
