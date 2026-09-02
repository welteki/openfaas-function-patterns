'use strict'

const thresholdC = 75.0

module.exports = async (event, context) => {
  let reading
  try {
    reading = parseBody(event.body)
  } catch (error) {
    return fail(context)
  }

  const value = reading.temperature_c
  if (!Number.isFinite(value)) {
    return fail(context)
  }

  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed({
      value_c: value,
      threshold_c: thresholdC,
      alert: value > thresholdC
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
