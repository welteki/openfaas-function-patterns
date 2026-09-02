# Fan-out pattern — URL health checks

The `fan-out` function accepts one trusted URL per line and submits each URL as
an asynchronous invocation of the `url-check` function. It returns a summary
immediately without waiting for the checks to complete.

Documentation: [Fan-out pattern](https://docs.openfaas.com/languages/patterns/fan-out/)

## Build and deploy

```bash
faas-cli up --tag=digest
```

Environment variables consumed by `fan-out`:

| Variable | Default | Description |
|---|---|---|
| `gateway_url` | `http://gateway.openfaas:8080` | In-cluster gateway address |
| `callback_url` | (none) | Optional `X-Callback-Url` for function results |

## Invoke

Newlines matter, so use `--data-binary`:

```bash
printf 'https://www.openfaas.com/\nhttps://docs.openfaas.com/\n' | \
  curl -s --data-binary @- http://127.0.0.1:8080/function/fan-out | jq
```

Example output:

```json
{
  "submitted": 2,
  "function": "url-check",
  "callback": false,
  "call_ids": [
    "9c0b1a12-fdea-4f01-baff-c5d9f50435ea",
    "4111d512-cdf3-4b8f-96b3-1b7f1f376bd7"
  ]
}
```

Each URL is queued for the `url-check` function. Track the number of function
invocations with `faas-cli list`, or cancel an individual check with its call
ID via `DELETE /async-function/<call-id>`.

## Callbacks

Deploy the `printer` function, then pass `X-Callback-Url` with the batch to
receive every health-check result:

```bash
faas-cli store deploy printer

printf 'https://www.openfaas.com/\nhttps://docs.openfaas.com/\n' | \
  curl -s --data-binary @- \
  -H "X-Callback-Url: http://gateway.openfaas:8080/function/printer" \
  http://127.0.0.1:8080/function/fan-out
```

Inspect the individual callback bodies with `faas-cli logs printer`. This
demonstrates result delivery only; it does not wait for all checks or combine
their results. On OpenFaaS Pro, callbacks may be restricted by the
queue-worker's `allowedCallbackURLs` setting.
