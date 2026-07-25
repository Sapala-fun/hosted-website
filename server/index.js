const http = require('node:http');

function getOwnerRezConfig() {
  return {
    apiBaseUrl: process.env.OWNERREZ_API_BASE_URL || '',
    apiKey: process.env.OWNERREZ_API_KEY || '',
    propertiesEndpoint: process.env.OWNERREZ_PROPERTIES_ENDPOINT || '/v1/properties',
    bookingEndpoint: process.env.OWNERREZ_BOOKING_ENDPOINT || '/v1/bookings',
  };
}

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
      const config = getOwnerRezConfig();

      if (!config.apiBaseUrl) {
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

      const targetUrl = new URL(`${config.apiBaseUrl}${config.propertiesEndpoint}`);
      const request = http.request(
        {
          hostname: targetUrl.hostname,
          port: targetUrl.port || (targetUrl.protocol === 'https:' ? 443 : 80),
          path: `${targetUrl.pathname}${targetUrl.search}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': config.apiKey,
          },
        },
        (proxyRes) => {
          let body = '';
          proxyRes.setEncoding('utf8');
          proxyRes.on('data', (chunk) => {
            body += chunk;
          });
          proxyRes.on('end', () => {
            try {
              const parsed = JSON.parse(body);
              sendJson(res, proxyRes.statusCode || 200, parsed);
            } catch (error) {
              sendJson(res, 502, { error: 'Invalid response from OwnerRez' });
            }
          });
        },
      );

      request.on('error', () => {
        sendJson(res, 502, { error: 'Unable to reach OwnerRez API' });
      });
      request.end();
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

          const config = getOwnerRezConfig();
          if (!config.apiBaseUrl) {
            sendJson(res, 200, {
              message: 'Booking request received',
              guestName,
              checkIn,
              checkOut,
              note: 'OwnerRez integration not configured yet. Add OWNERREZ_API_BASE_URL and OWNERREZ_API_KEY to enable live booking.',
            });
            return;
          }

          const targetUrl = new URL(`${config.apiBaseUrl}${config.bookingEndpoint}`);
          const bookingRequest = http.request(
            {
              hostname: targetUrl.hostname,
              port: targetUrl.port || (targetUrl.protocol === 'https:' ? 443 : 80),
              path: `${targetUrl.pathname}${targetUrl.search}`,
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'X-API-Key': config.apiKey,
              },
            },
            (proxyRes) => {
              let responseBody = '';
              proxyRes.setEncoding('utf8');
              proxyRes.on('data', (chunk) => {
                responseBody += chunk;
              });
              proxyRes.on('end', () => {
                try {
                  const parsedResponse = JSON.parse(responseBody);
                  sendJson(res, proxyRes.statusCode || 200, parsedResponse);
                } catch (error) {
                  sendJson(res, 502, { error: 'Invalid booking response from OwnerRez' });
                }
              });
            },
          );

          bookingRequest.on('error', () => {
            sendJson(res, 502, { error: 'Unable to submit booking to OwnerRez API' });
          });
          bookingRequest.write(JSON.stringify(parsed));
          bookingRequest.end();
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
