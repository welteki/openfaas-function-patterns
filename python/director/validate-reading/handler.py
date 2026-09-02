import json


def handle(event, context):
    try:
        reading = json.loads(event.body)
    except (TypeError, ValueError):
        return error("expected a JSON sensor reading")

    device_id = str(reading.get("device_id", "")).strip()
    temperature = reading.get("temperature_c", 0)
    battery = reading.get("battery_percent", 0)

    if not device_id:
        return error("device_id is required")
    if not is_number(temperature) or temperature < -100 or temperature > 200:
        return error("temperature_c must be between -100 and 200")
    if (
        not isinstance(battery, int)
        or isinstance(battery, bool)
        or battery < 0
        or battery > 100
    ):
        return error("battery_percent must be between 0 and 100")

    return {
        "statusCode": 200,
        "body": {
            "device_id": device_id,
            "temperature_c": temperature,
            "battery_percent": battery,
        },
    }


def is_number(value):
    return isinstance(value, (int, float)) and not isinstance(value, bool)


def error(message):
    return {"statusCode": 400, "body": message}
