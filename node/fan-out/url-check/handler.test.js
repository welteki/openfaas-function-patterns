'use strict'

const assert = require('node:assert/strict')
const { afterEach, test } = require('node:test')

const handler = require('./handler')

const originalFetch = global.fetch

afterEach(() => {
  global.fetch = originalFetch
})

test('reports a healthy URL', async () => {
  global.fetch = async () => new Response(null, {
    status: 204,
    headers: { 'Content-Type': 'text/plain' }
  })

  const context = new TestContext()
  await handler({ body: 'https://www.openfaas.com/' }, context)

  assert.equal(context.statusCode, 200)
  assert.equal(context.body.reachable, true)
  assert.equal(context.body.healthy, true)
})

test('reports an unhealthy HTTP status', async () => {
  global.fetch = async () => new Response('unavailable', { status: 503 })

  const context = new TestContext()
  await handler({ body: 'https://one.example' }, context)

  assert.equal(context.body.reachable, true)
  assert.equal(context.body.healthy, false)
  assert.equal(context.body.status_code, 503)
})

test('returns network failures as a health result', async () => {
  global.fetch = async () => {
    throw new Error('request timed out')
  }

  const context = new TestContext()
  await handler({ body: 'https://one.example' }, context)

  assert.equal(context.statusCode, 200)
  assert.equal(context.body.reachable, false)
  assert.match(context.body.error, /timed out/)
})

test('rejects a non-HTTP URL', async () => {
  const context = new TestContext()
  await handler({ body: 'file:///etc/passwd' }, context)

  assert.equal(context.statusCode, 400)
})

class TestContext {
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
