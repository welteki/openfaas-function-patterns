from types import SimpleNamespace

try:
    from . import handler as h
except ImportError:
    import handler as h


class FakeResponse:
    status_code = 202
    text = ""
    headers = {"X-Call-Id": "call-1"}


def event(body, callback=""):
    headers = {"X-Callback-Url": callback} if callback else {}
    return SimpleNamespace(body=body.encode(), headers=headers)


def test_submits_each_url_with_callback(monkeypatch):
    submitted = []

    def fake_post(url, data, headers, timeout):
        submitted.append((url, data.decode(), headers, timeout))
        return FakeResponse()

    monkeypatch.setattr(h.requests, "post", fake_post)
    response = h.handle(
        event(
            "https://one.example\nhttps://two.example\n",
            "http://gateway.openfaas:8080/function/printer",
        ),
        {},
    )

    assert response["statusCode"] == 200
    assert response["body"]["submitted"] == 2
    assert response["body"]["callback"] is True
    assert len(submitted) == 2
    assert submitted[0][2]["X-Callback-Url"].endswith("/function/printer")


def test_rejects_empty_batch():
    response = h.handle(event("\n\n"), {})
    assert response["statusCode"] == 400


def test_reports_submission_failure(monkeypatch):
    failed = FakeResponse()
    failed.status_code = 503
    failed.text = "queue unavailable"
    monkeypatch.setattr(h.requests, "post", lambda *args, **kwargs: failed)

    response = h.handle(event("https://one.example\n"), {})
    assert response["statusCode"] == 502
