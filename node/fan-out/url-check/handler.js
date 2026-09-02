'use strict'

const { performance } = require('node:perf_hooks')

const maxURLLength = 4096
const requestTimeout = configuredTimeout(
  process.env.request_timeout || '5',
  'request_timeout'
)

module.exports = async (event, context) => {
  const target = requestBody(event.body).trim()
  if (Buffer.byteLength(target) > maxURLLength) {
    return fail(context, 'URL is too long')
  }

  let parsed
  try {
    parsed = new URL(target)
  } catch (error) {
    return fail(context, 'expected an absolute HTTP or HTTPS URL')
  }
  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    return fail(context, 'expected an absolute HTTP or HTTPS URL')
  }

  const started = performance.now()
  const result = {
    url: target,
    reachable: false,
    healthy: false
  }

  let response
  try {
    response = await fetch(target, {
      headers: { 'User-Agent': 'OpenFaaS URL health check' },
      signal: AbortSignal.timeout(requestTimeout)
    })
  } catch (error) {
    result.duration_ms = Math.round(performance.now() - started)
    result.error = error.message
    return succeed(context, result)
  }

  if (response.body) {
    await response.body.cancel()
  }
  result.reachable = true
  result.healthy = response.status >= 200 && response.status < 400
  result.status_code = response.status
  result.content_type = response.headers.get('Content-Type') || ''
  result.duration_ms = Math.round(performance.now() - started)
  return succeed(context, result)
}

function requestBody (body) {
  if (Buffer.isBuffer(body)) {
    return body.toString()
  }
  return String(body || '')
}

function configuredTimeout (value, name) {
  const seconds = Number(value)
  if (!Number.isFinite(seconds) || seconds <= 0) {
    throw new Error(`${name} must be greater than zero`)
  }
  return seconds * 1000
}

function succeed (context, body) {
  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed(body)
}

function fail (context, message) {
  return context
    .status(400)
    .headers({ 'Content-Type': 'text/plain' })
    .succeed(message)
}
