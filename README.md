Examples for the OpenFaaS documentation. Each example is
self-contained: it has its own `stack.yaml` and function directories, so it
can be built and deployed on its own.

## Examples

- [**Director**](https://docs.openfaas.com/languages/patterns/director/) — [Go](go/director/) · [Python](python/director/) · [Node.js](node/director/)
- [**Fan-out**](https://docs.openfaas.com/languages/patterns/fan-out/) — [Go](go/fan-out/) · [Python](python/fan-out/) · [Node.js](node/fan-out/)
- [**Singleton**](https://docs.openfaas.com/languages/patterns/singleton/) — [Go](go/singleton/) · [Python](python/singleton/)

## Prerequisites

- [faas-cli](https://docs.openfaas.com/cli/install/)
- Docker, for building images
- A running OpenFaaS cluster, for e2e testing — set `OPENFAAS_URL` and
  provide credentials, or point the `gateway` in each `stack.yaml` at a
  local gateway

## Build and test an example

```bash
cd go/director
faas-cli up --tag=digest
```

`faas-cli up` builds each function image, pushes it to the registry, and
deploys it to the gateway.

The examples default to the public [ttl.sh](https://ttl.sh) registry so they
can be pushed without credentials. The registry and owner are overridable
through environment variables:

```bash
REGISTRY=ghcr.io OWNER=welteki faas-cli up
```
