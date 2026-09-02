Worked examples for the OpenFaaS documentation. Each example is
self-contained: it has its own `stack.yaml` and function directories, so it
can be built and deployed on its own.

## Structure

```
<language>/
  <example>/
    stack.yaml
    <function>/
      handler.go, handler.py, or handler.js
```

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

The examples use the public [ttl.sh](https://ttl.sh) registry so they can be
pushed without credentials. Replace the prefix in `stack.yaml` with your own
registry for anything you would not publish openly.

## Examples

| Directory | Pattern | Documentation |
|---|---|---|
| `go/director/` | Director: chains functions in a workflow | /languages/patterns/director/ |
| `go/fan-out/` | Fan-out: submits one async invocation per record | /languages/patterns/fan-out/ |
| `go/singleton/` | Singleton: broadcasts notifications from one replica | /languages/patterns/singleton/ |
| `python/director/` | Director: chains functions in a workflow | /languages/patterns/director/ |
| `python/fan-out/` | Fan-out: submits one async invocation per record | /languages/patterns/fan-out/ |
| `python/singleton/` | Singleton: broadcasts notifications from one replica | /languages/patterns/singleton/ |
| `node/director/` | Director: chains functions in a workflow | /languages/patterns/director/ |
| `node/fan-out/` | Fan-out: submits one async invocation per record | /languages/patterns/fan-out/ |
