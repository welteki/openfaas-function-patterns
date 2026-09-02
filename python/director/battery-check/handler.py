import json


THRESHOLD_PERCENT = 20


def handle(event, context):
    try:
        reading = json.loads(event.body)
        value = reading["battery_percent"]
    except (KeyError, TypeError, ValueError):
        return {"statusCode": 400, "body": "expected a JSON sensor reading"}

    return {
        "statusCode": 200,
        "body": {
            "value_percent": value,
            "threshold_percent": THRESHOLD_PERCENT,
            "alert": value < THRESHOLD_PERCENT,
        },
    }
