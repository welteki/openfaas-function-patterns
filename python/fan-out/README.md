# Fan-out pattern — URL health checks

This is the Python implementation of the fan-out example documented in the
OpenFaaS function patterns guide. It submits each URL asynchronously and can
forward an optional callback URL for individual results.

Documentation: [Fan-out pattern](https://docs.openfaas.com/languages/patterns/fan-out/)

```bash
faas-cli template store pull python3-http
faas-cli up --tag=digest
```
