import os

import requests


TARGET_FUNCTION = "url-check"
SUBMIT_TIMEOUT = 30


def handle(event, context):
    body = (
        event.body.decode()
        if isinstance(event.body, bytes)
        else str(event.body)
    )
    records = [
        record.strip()
        for record in body.strip().splitlines()
        if record.strip()
    ]
    if not records:
        return error(400, "expected one record per line in the request body")

    gateway = os.getenv("gateway_url", "http://gateway.openfaas:8080")
    callback = event.headers.get("X-Callback-Url", "").strip()
    if not callback:
        callback = os.getenv("callback_url", "").strip()

    call_ids = []
    for index, record in enumerate(records):
        try:
            call_id = submit(gateway, record, callback)
        except requests.RequestException as err:
            return error(502, f"record {index + 1} of {len(records)}: {err}")
        except RuntimeError as err:
            return error(502, f"record {index + 1} of {len(records)}: {err}")

        if call_id:
            call_ids.append(call_id)

    response = {
        "submitted": len(records),
        "function": TARGET_FUNCTION,
        "callback": bool(callback),
    }
    if call_ids:
        response["call_ids"] = call_ids

    return {"statusCode": 200, "body": response}


def submit(gateway, record, callback):
    headers = {"Content-Type": "text/plain"}
    if callback:
        headers["X-Callback-Url"] = callback

    response = requests.post(
        f"{gateway.rstrip('/')}/async-function/{TARGET_FUNCTION}",
        data=record.encode(),
        headers=headers,
        timeout=SUBMIT_TIMEOUT,
    )
    if response.status_code != 202:
        raise RuntimeError(
            f"unexpected status {response.status_code} "
            f"from {TARGET_FUNCTION}: {response.text}"
        )

    return response.headers.get("X-Call-Id", "")


def error(status_code, message):
    return {"statusCode": status_code, "body": message}
