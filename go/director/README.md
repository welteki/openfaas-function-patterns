# Director pattern — telemetry triage

The director function `telemetry-workflow` validates a sensor reading, invokes
temperature and battery checks in parallel, and combines their results into an
`ok` or `alert` response:

```
telemetry-workflow ◄── director: owns the workflow and handles errors
    ├── 1. invoke ──► validate-reading
    ├── 2. invoke in parallel
    │       ├──► temperature-check ──┐
    │       └──► battery-check ──────┤
    │◄──── check results ────────────┘
    └── 3. combine results and return ok or alert
```

Documentation page: `docs/languages/patterns/director.md` in the
openfaas-docs repository.

## Build and deploy

```bash
faas-cli up --tag=digest
```

The stage functions are invoked through the gateway. The `stage_timeout`
environment variable limits each downstream call, while the watchdog timeout
variables limit the complete director invocation.

## Invoke

```bash
curl -s http://127.0.0.1:8080/function/telemetry-workflow \
  -H "Content-Type: application/json" \
  -d '{"device_id":"pump-17","temperature_c":82.4,"battery_percent":12}' | \
  jq
```

Example output:

```json
{
  "device_id": "pump-17",
  "status": "alert",
  "temperature": {
    "value_c": 82.4,
    "threshold_c": 75,
    "alert": true
  },
  "battery": {
    "value_percent": 12,
    "threshold_percent": 20,
    "alert": true
  },
  "duration_ms": 4
}
```
