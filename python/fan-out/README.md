# Fan-out pattern — URL health checks

This is the Python implementation of the fan-out example documented in the
OpenFaaS function patterns guide. It submits each URL asynchronously and can
forward an optional callback URL for individual results.

Documentation page: `docs/languages/patterns/fan-out.md` in the
openfaas-docs repository.

```bash
faas-cli template store pull python3-http
faas-cli up --tag=digest
```
