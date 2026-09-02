'use strict'

const assert = require('node:assert/strict')
const { afterEach, test } = require('node:test')

const handler = require('./handler')

const originalFetch = global.fetch

afterEach(() => {
  global.fetch = originalFetch
})

test('runs validation before both checks and combines results', async () => {
  const calls = []
  global.fetch = async (url) => {
    const name = url.split('/').pop()
    calls.push(name)
    if (name === 'validate-reading') {
      return jsonResponse({
        device_id: 'pump-17',
        temperature_c: 82.4,
        battery_percent: 12
      })
    }
    if (name === 'temperature-check') {
      return jsonResponse({
        value_c: 82.4,
        threshold_c: 75,
        alert: true
      })
    }
    return jsonResponse({
      value_percent: 12,
      threshold_percent: 20,
      alert: true
    })
  }

  const context = new TestContext()
  await handler({ body: inputReading() }, context)

  assert.equal(context.statusCode, 200)
  assert.equal(context.body.status, 'alert')
  assert.deepEqual(calls, [
    'validate-reading',
    'temperature-check',
    'battery-check'
  ])
})

test('passes through a validation error', async () => {
  global.fetch = async () => new Response('device_id is required', {
    status: 400
  })

  const context = new TestContext()
  await handler({ body: inputReading() }, context)

  assert.equal(context.statusCode, 400)
  assert.match(context.body, /device_id is required/)
})

test('returns bad gateway when a function cannot be invoked', async () => {
  global.fetch = async () => {
    throw new Error('connection refused')
  }

  const context = new TestContext()
  await handler({ body: inputReading() }, context)

  assert.equal(context.statusCode, 502)
  assert.match(context.body, /connection refused/)
})

function inputReading () {
  return {
    device_id: 'pump-17',
    temperature_c: 82.4,
    battery_percent: 12
  }
}

function jsonResponse (body) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' }
  })
}

class TestContext {
  constructor () {
    this.statusCode = 200
    this.headerValues = {}
    this.body = undefined
  }

  status (value) {
    this.statusCode = value
    return this
  }

  headers (value) {
    this.headerValues = value
    return this
  }

  succeed (value) {
    this.body = value
    return value
  }
}
