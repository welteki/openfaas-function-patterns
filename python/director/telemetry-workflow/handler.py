import concurrent.futures
import os
import time

import requests


GATEWAY_URL = os.getenv(
    "gateway_url", "http://gateway.openfaas:8080"
).rstrip("/")
STAGE_TIMEOUT = float(os.getenv("stage_timeout", "5"))
if STAGE_TIMEOUT <= 0:
    raise ValueError("stage_timeout must be greater than zero")


def handle(event, context):
    started = time.monotonic()
    body = event.body

    try:
        validated = invoke("validate-reading", body)
    except requests.RequestException as err:
        return error(502, f"failed to invoke validate-reading: {err}")

    if validated.status_code != 200:
        return error(
            validated.status_code,
            f"validate-reading failed: {validated.text}",
        )

    try:
        reading = validated.json()
    except ValueError as err:
        return error(502, f"unexpected response from validate-reading: {err}")

    functions = ("temperature-check", "battery-check")
    completed = {}
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        futures = {
            name: executor.submit(invoke, name, validated.content)
            for name in functions
        }
        for name, future in futures.items():
            try:
                completed[name] = future.result()
            except requests.RequestException as err:
                return error(502, f"failed to invoke {name}: {err}")

    for name in functions:
        response = completed[name]
        if response.status_code != 200:
            return error(
                response.status_code,
                f"{name} failed: {response.text}",
            )

    try:
        temperature = completed["temperature-check"].json()
        battery = completed["battery-check"].json()
    except ValueError as err:
        return error(502, f"unexpected response from check function: {err}")

    has_alert = temperature["alert"] or battery["alert"]
    workflow_status = "alert" if has_alert else "ok"
    return {
        "statusCode": 200,
        "body": {
            "device_id": reading["device_id"],
            "status": workflow_status,
            "temperature": temperature,
            "battery": battery,
            "duration_ms": int((time.monotonic() - started) * 1000),
        },
    }


def invoke(function, body):
    return requests.post(
        f"{GATEWAY_URL}/function/{function}",
        data=body,
        headers={"Content-Type": "application/json"},
        timeout=STAGE_TIMEOUT,
    )


def error(status_code, message):
    return {"statusCode": status_code, "body": message.strip()}
