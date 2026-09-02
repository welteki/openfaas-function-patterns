import json


THRESHOLD_C = 75.0


def handle(event, context):
    try:
        reading = json.loads(event.body)
        value = reading["temperature_c"]
    except (KeyError, TypeError, ValueError):
        return {"statusCode": 400, "body": "expected a JSON sensor reading"}

    return {
        "statusCode": 200,
        "body": {
            "value_c": value,
            "threshold_c": THRESHOLD_C,
            "alert": value > THRESHOLD_C,
        },
    }
