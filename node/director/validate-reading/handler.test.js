'use strict'

const assert = require('node:assert/strict')
const { test } = require('node:test')

const handler = require('./handler')

test('normalizes a valid reading', async () => {
  const context = new TestContext()
  await handler({
    body: {
      device_id: ' pump-17 ',
      temperature_c: 42.5,
      battery_percent: 80
    }
  }, context)

  assert.equal(context.statusCode, 200)
  assert.deepEqual(context.body, {
    device_id: 'pump-17',
    temperature_c: 42.5,
    battery_percent: 80
  })
})

test('rejects a missing device ID', async () => {
  const context = new TestContext()
  await handler({ body: { temperature_c: 20 } }, context)

  assert.equal(context.statusCode, 400)
  assert.match(context.body, /device_id is required/)
})

test('rejects values outside the accepted ranges', async () => {
  const context = new TestContext()
  await handler({
    body: {
      device_id: 'pump-17',
      temperature_c: 250,
      battery_percent: 80
    }
  }, context)

  assert.equal(context.statusCode, 400)
  assert.match(context.body, /temperature_c/)
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
