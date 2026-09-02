# Singleton pattern — notification hub

This is the Python implementation of the singleton SSE example documented in
the OpenFaaS function patterns guide. It uses the `python3-flask` template so
the handler can return a streaming Flask response.

Documentation page: `docs/languages/patterns/singleton.md` in the
openfaas-docs repository.

```bash
faas-cli template store pull python3-flask
faas-cli up --tag=digest
```
