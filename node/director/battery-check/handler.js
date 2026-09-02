'use strict'

const thresholdPercent = 20

module.exports = async (event, context) => {
  let reading
  try {
    reading = parseBody(event.body)
  } catch (error) {
    return fail(context)
  }

  const value = reading.battery_percent
  if (!Number.isInteger(value)) {
    return fail(context)
  }

  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed({
      value_percent: value,
      threshold_percent: thresholdPercent,
      alert: value < thresholdPercent
    })
}

function parseBody (body) {
  if (Buffer.isBuffer(body)) {
    return JSON.parse(body.toString())
  }
  return typeof body === 'string' ? JSON.parse(body) : body
}

function fail (context) {
  return context
    .status(400)
    .headers({ 'Content-Type': 'text/plain' })
    .succeed('expected a JSON sensor reading')
}
