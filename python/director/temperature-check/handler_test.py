import json
from types import SimpleNamespace

try:
    from . import handler as h
except ImportError:
    import handler as h


def test_sets_alert_above_threshold():
    event = SimpleNamespace(body=json.dumps({"temperature_c": 82.4}).encode())
    response = h.handle(event, {})

    assert response["statusCode"] == 200
    assert response["body"]["alert"] is True
    assert response["body"]["threshold_c"] == 75.0
