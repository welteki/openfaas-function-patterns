'use strict'

const targetFunction = 'url-check'
const submitTimeout = 30000

module.exports = async (event, context) => {
  const input = requestBody(event.body)
  const records = input
    .trim()
    .split('\n')
    .map((record) => record.trim())
    .filter(Boolean)

  if (records.length === 0) {
    return fail(
      context,
      400,
      'expected one record per line in the request body'
    )
  }

  const gateway = process.env.gateway_url ||
    'http://gateway.openfaas:8080'
  const headers = event.headers || {}
  const callback = String(
    headers['x-callback-url'] || process.env.callback_url || ''
  ).trim()

  const callIDs = []
  for (const [index, record] of records.entries()) {
    let callID
    try {
      callID = await submit(gateway, record, callback)
    } catch (error) {
      return fail(
        context,
        502,
        `record ${index + 1} of ${records.length}: ${error.message}`
      )
    }
    if (callID) {
      callIDs.push(callID)
    }
  }

  const response = {
    submitted: records.length,
    function: targetFunction,
    callback: Boolean(callback)
  }
  if (callIDs.length > 0) {
    response.call_ids = callIDs
  }

  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed(response)
}

async function submit (gateway, record, callback) {
  const headers = { 'Content-Type': 'text/plain' }
  if (callback) {
    headers['X-Callback-Url'] = callback
  }

  const baseURL = gateway.replace(/\/$/, '')
  const response = await fetch(
    `${baseURL}/async-function/${targetFunction}`,
    {
      method: 'POST',
      headers,
      body: record,
      signal: AbortSignal.timeout(submitTimeout)
    }
  )

  if (response.status !== 202) {
    const body = await response.text()
    throw new Error(
      `unexpected status ${response.status} ` +
      `from ${targetFunction}: ${body}`
    )
  }

  return response.headers.get('X-Call-Id') || ''
}

function requestBody (body) {
  if (Buffer.isBuffer(body)) {
    return body.toString()
  }
  return String(body || '')
}

function fail (context, status, message) {
  return context
    .status(status)
    .headers({ 'Content-Type': 'text/plain' })
    .succeed(message)
}
