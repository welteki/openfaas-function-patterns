# Director pattern — telemetry triage

This is the Python implementation of the telemetry workflow documented in the
OpenFaaS function patterns guide. It uses `python3-http` for all four
functions and runs the temperature and battery checks concurrently.

Documentation page: `docs/languages/patterns/director.md` in the
openfaas-docs repository.

```bash
faas-cli template store pull python3-http
faas-cli up --tag=digest
```

```bash
curl -s http://127.0.0.1:8080/function/telemetry-workflow \
  -H "Content-Type: application/json" \
  -d '{"device_id":"pump-17","temperature_c":82.4,"battery_percent":12}' | \
  jq
```
