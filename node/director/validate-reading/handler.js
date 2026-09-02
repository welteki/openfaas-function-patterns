'use strict'

module.exports = async (event, context) => {
  let reading
  try {
    reading = parseBody(event.body)
  } catch (error) {
    return fail(context, 'expected a JSON sensor reading')
  }

  const deviceID = String(reading.device_id || '').trim()
  const temperature = reading.temperature_c ?? 0
  const battery = reading.battery_percent ?? 0

  if (!deviceID) {
    return fail(context, 'device_id is required')
  }
  if (
    !Number.isFinite(temperature) ||
    temperature < -100 ||
    temperature > 200
  ) {
    return fail(
      context,
      'temperature_c must be between -100 and 200'
    )
  }
  if (!Number.isInteger(battery) || battery < 0 || battery > 100) {
    return fail(
      context,
      'battery_percent must be between 0 and 100'
    )
  }

  return context
    .status(200)
    .headers({ 'Content-Type': 'application/json' })
    .succeed({
      device_id: deviceID,
      temperature_c: temperature,
      battery_percent: battery
    })
}

function parseBody (body) {
  if (Buffer.isBuffer(body)) {
    return JSON.parse(body.toString())
  }
  return typeof body === 'string' ? JSON.parse(body) : body
}

function fail (context, message) {
  return context
    .status(400)
    .headers({ 'Content-Type': 'text/plain' })
    .succeed(message)
}
