const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
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

test('properties endpoint can proxy to a configured OwnerRez API', async () => {
  const ownerRezMock = http.createServer((req, res) => {
    assert.equal(req.url, '/v1/properties');
    assert.equal(req.headers['x-api-key'], 'demo-key');
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ properties: [{ name: 'Proxy Property', nightlyRate: 315 }] }));
  });

  ownerRezMock.listen(0, '127.0.0.1');
  await once(ownerRezMock, 'listening');

  const previousBaseUrl = process.env.OWNERREZ_API_BASE_URL;
  const previousApiKey = process.env.OWNERREZ_API_KEY;
  const previousEndpoint = process.env.OWNERREZ_PROPERTIES_ENDPOINT;

  process.env.OWNERREZ_API_BASE_URL = `http://127.0.0.1:${ownerRezMock.address().port}`;
  process.env.OWNERREZ_API_KEY = 'demo-key';
  process.env.OWNERREZ_PROPERTIES_ENDPOINT = '/v1/properties';

  const server = await startServer();
  const address = server.address();

  try {
    const response = await fetch(`http://127.0.0.1:${address.port}/api/properties`);
    const body = await response.json();

    assert.equal(response.status, 200);
    assert.equal(body.properties[0].name, 'Proxy Property');
  } finally {
    await stopServer(server);
    await new Promise((resolve, reject) => ownerRezMock.close((error) => (error ? reject(error) : resolve())));
    if (previousBaseUrl === undefined) delete process.env.OWNERREZ_API_BASE_URL; else process.env.OWNERREZ_API_BASE_URL = previousBaseUrl;
    if (previousApiKey === undefined) delete process.env.OWNERREZ_API_KEY; else process.env.OWNERREZ_API_KEY = previousApiKey;
    if (previousEndpoint === undefined) delete process.env.OWNERREZ_PROPERTIES_ENDPOINT; else process.env.OWNERREZ_PROPERTIES_ENDPOINT = previousEndpoint;
  }
});
