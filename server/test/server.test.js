const test = require('node:test');
const assert = require('node:assert/strict');
const { once } = require('node:events');
const { createServer } = require('../index');

async function startServer() {
  const server = createServer();
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  return server;
}

async function stopServer(server) {
  await new Promise((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
}

test('health endpoint returns ok', async () => {
  const server = await startServer();
  const address = server.address();

  try {
    const response = await fetch(`http://127.0.0.1:${address.port}/api/health`);
    const body = await response.json();

    assert.equal(response.status, 200);
    assert.equal(body.status, 'ok');
  } finally {
    await stopServer(server);
  }
});

test('properties endpoint returns sample property data', async () => {
  const server = await startServer();
  const address = server.address();

  try {
    const response = await fetch(`http://127.0.0.1:${address.port}/api/properties`);
    const body = await response.json();

    assert.equal(response.status, 200);
    assert.equal(body.properties[0].name, 'Ocean View Retreat');
  } finally {
    await stopServer(server);
  }
});

test('booking endpoint accepts payload and returns confirmation', async () => {
  const server = await startServer();
  const address = server.address();

  try {
    const response = await fetch(`http://127.0.0.1:${address.port}/api/book`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ guestName: 'Ada', checkIn: '2026-08-01', checkOut: '2026-08-03' }),
    });
    const body = await response.json();

    assert.equal(response.status, 200);
    assert.equal(body.message, 'Booking request received');
  } finally {
    await stopServer(server);
  }
});
