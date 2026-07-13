const http = require('node:http');

function sendJson(res, statusCode, payload, extraHeaders = {}) {
  res.writeHead(statusCode, {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    ...extraHeaders,
  });
  res.end(JSON.stringify(payload));
}

function createServer() {
  return http.createServer((req, res) => {
    if (req.method === 'OPTIONS') {
      res.writeHead(204, {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      });
      res.end();
      return;
    }

    const url = new URL(req.url, 'http://localhost');

    if (url.pathname === '/api/health') {
      sendJson(res, 200, {
        status: 'ok',
        service: 'ownerrez-proxy',
        timestamp: new Date().toISOString(),
      });
      return;
    }

    if (url.pathname === '/api/properties') {
      sendJson(res, 200, {
        properties: [
          {
            id: 'sample-property',
            name: 'Ocean View Retreat',
            slug: 'ocean-view-retreat',
            nightlyRate: 210,
            bedrooms: 2,
            bathrooms: 2,
            description: 'A starter property payload for the GitHub Pages frontend.',
          },
        ],
      });
      return;
    }

    if (url.pathname === '/api/book') {
      let body = '';
      req.on('data', (chunk) => {
        body += chunk;
      });

      req.on('end', () => {
        try {
          const parsed = body ? JSON.parse(body) : {};
          const { guestName, checkIn, checkOut } = parsed;

          if (!guestName || !checkIn || !checkOut) {
            sendJson(res, 400, { error: 'Missing required booking fields' });
            return;
          }

          sendJson(res, 200, {
            message: 'Booking request received',
            guestName,
            checkIn,
            checkOut,
            note: 'Replace this stub with a real OwnerRez booking integration.',
          });
        } catch (error) {
          sendJson(res, 400, { error: 'Invalid JSON payload' });
        }
      });
      return;
    }

    sendJson(res, 404, { error: 'not_found' });
  });
}

if (require.main === module) {
  const port = Number(process.env.PORT || 3001);
  const server = createServer();
  server.listen(port, () => {
    console.log(`Server listening on port ${port}`);
  });
}

module.exports = { createServer };
