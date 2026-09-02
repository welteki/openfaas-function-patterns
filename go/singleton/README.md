# Singleton pattern — live notification hub

The `notification-hub` function broadcasts notifications to every client with
an open Server-Sent Events (SSE) connection. Its scaling labels keep exactly
one replica running, so publishers and subscribers share the same in-memory
subscriber list.

WebSockets have the same connection-local state concern and provide
bidirectional communication. See the OpenFaaS article
[How to Integrate WebSockets with Serverless Functions and OpenFaaS](https://www.openfaas.com/blog/serverless-websockets/).

The subscriber list is deliberately connection-local. Clients reconnect when
the function is restarted, rescheduled, or redeployed. To scale the hub across
multiple replicas, replace the in-memory broadcast with an external message
bus.

Documentation page: `docs/languages/patterns/singleton.md` in the
openfaas-docs repository.

## Build and deploy

```bash
faas-cli up --tag=digest
```

## Invoke

In one terminal, subscribe to notifications:

```bash
curl -N -H "Accept: text/event-stream" \
  http://127.0.0.1:8080/function/notification-hub
```

In another terminal, publish a notification:

```bash
curl -s -d "deployment complete" \
  http://127.0.0.1:8080/function/notification-hub
```

The subscriber receives `data: deployment complete` immediately.
