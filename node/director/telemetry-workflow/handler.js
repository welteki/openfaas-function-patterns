'use strict'

const { performance } = require('node:perf_hooks')

const gatewayURL = process.env.gateway_url ||
  'http://gateway.openfaas:8080'
const stageTimeout = configuredTimeout(
  process.env.stage_timeout || '5',
  'stage_timeout'
)

module.exports = async (event, context) => {
  const started = performance.now()
  const input = requestBody(event.body)

  let validated
  try {
    validated = await invoke('validate-reading', input)
  } catch (error) {
    return fail(
      context,
      502,
      `failed to invoke validate-reading: ${error.message}`
    )
  }

  if (validated.status !== 200) {
    return fail(
      context,
      validated.status,
      `validate-reading failed: ${await validated.text()}`
    )
  }

  let reading
  try {
    reading = await validated.json()
  } catch (error) {
    return fail(
      context,
      502,
      `unexpected response from validate-reading: ${error.message}`
    )
  }

  const body = JSON.stringify(reading)
  const names = ['temperature-check', 'battery-check']
  let responses
  try {
    responses = await Promise.all(
      names.map(async (name) => [name, await invoke(name, body)])
    )
  } catch (error) {
    return fail(context, 502, `failed to invoke check: ${error.message}`)
  }

  const completed = Object.fromEntries(responses)
  for (const name of names) {
    const response = completed[name]
    if (response.status !== 200) {
      return fail(
        context,
        response.status,
        `${name} failed: ${await response.text()}`
      )
    }
  }

  let temperature
  let battery
  try {
    temperature = await completed['temperature-check'].json()
    battery = await completed['battery-check'].json()
  } catch (error) {
    return fail(
      context,
      502,
      `unexpected response from check function: ${error.message}`
    )
  }

  const status = temperature.alert || battery.alert ? 'alert' : 'ok'
  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed({
      device_id: reading.device_id,
      status,
      temperature,
      battery,
      duration_ms: Math.round(performance.now() - started)
    })
}

function invoke (name, body) {
  const gateway = gatewayURL.replace(/\/$/, '')
  return fetch(`${gateway}/function/${name}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
    signal: AbortSignal.timeout(stageTimeout)
  })
}

function requestBody (body) {
  if (Buffer.isBuffer(body)) {
    return body
  }
  return typeof body === 'string' ? body : JSON.stringify(body)
}

function configuredTimeout (value, name) {
  const seconds = Number(value)
  if (!Number.isFinite(seconds) || seconds <= 0) {
    throw new Error(`${name} must be greater than zero`)
  }
  return seconds * 1000
}

function fail (context, status, message) {
  return context
    .status(status)
    .headers({ 'Content-Type': 'text/plain' })
    .succeed(message.trim())
}
