'use strict'

const assert = require('node:assert/strict')
const { afterEach, test } = require('node:test')

const handler = require('./handler')

const originalFetch = global.fetch

afterEach(() => {
  global.fetch = originalFetch
})

test('submits one asynchronous invocation per URL', async () => {
  const requests = []
  global.fetch = async (url, options) => {
    requests.push({ url, options })
    return new Response('', {
      status: 202,
      headers: { 'X-Call-Id': `call-${requests.length}` }
    })
  }

  const context = new TestContext()
  await handler({
    body: 'https://one.example\nhttps://two.example\n',
    headers: {}
  }, context)

  assert.equal(context.statusCode, 200)
  assert.equal(context.body.submitted, 2)
  assert.deepEqual(context.body.call_ids, ['call-1', 'call-2'])
  assert.equal(requests[0].options.body, 'https://one.example')
})

test('forwards the callback URL to every invocation', async () => {
  const requests = []
  global.fetch = async (url, options) => {
    requests.push(options)
    return new Response('', { status: 202 })
  }

  const context = new TestContext()
  await handler({
    body: 'https://one.example',
    headers: {
      'x-callback-url': 'https://results.example/callback'
    }
  }, context)

  assert.equal(context.body.callback, true)
  assert.equal(
    requests[0].headers['X-Callback-Url'],
    'https://results.example/callback'
  )
})

test('returns bad gateway when a submission is rejected', async () => {
  global.fetch = async () => new Response('queue unavailable', {
    status: 503
  })

  const context = new TestContext()
  await handler({ body: 'https://one.example', headers: {} }, context)

  assert.equal(context.statusCode, 502)
  assert.match(context.body, /queue unavailable/)
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
