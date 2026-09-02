'use strict'

const assert = require('node:assert/strict')
const { test } = require('node:test')

const handler = require('./handler')

test('raises an alert below the battery threshold', async () => {
  const context = new TestContext()
  await handler({ body: { battery_percent: 12 } }, context)

  assert.equal(context.statusCode, 200)
  assert.deepEqual(context.body, {
    value_percent: 12,
    threshold_percent: 20,
    alert: true
  })
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
