'use strict'

const assert = require('node:assert/strict')
const { test } = require('node:test')

const handler = require('./handler')

test('raises an alert above the temperature threshold', async () => {
  const context = new TestContext()
  await handler({ body: { temperature_c: 82.4 } }, context)

  assert.equal(context.statusCode, 200)
  assert.deepEqual(context.body, {
    value_c: 82.4,
    threshold_c: 75,
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
