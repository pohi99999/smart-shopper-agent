#!/usr/bin/env node
/**
 * simulate_webhook.js
 *
 * Simulates the n8n Price Updater workflow by sending a POST request with
 * test price data directly to the backend's admin ingest endpoint.
 *
 * What it does:
 *   1. Reads ADMIN_TOKEN and BACKEND_HOST from ../.env
 *   2. POSTs a modified test prices payload to /api/v1/admin/prices
 *   3. Validates the HTTP 200 response
 *   4. Prints the diff between test and production prices
 *
 * After running this script, restore the original prices.json with:
 *   git checkout -- internal/data/prices.json
 *
 * Usage:
 *   node scripts/simulate_webhook.js
 */

'use strict';

const fs    = require('fs');
const path  = require('path');
const http  = require('http');
const https = require('https');

// ── Config ────────────────────────────────────────────────────────────────────

const ROOT    = path.resolve(__dirname, '..');
const ENV_FILE = path.join(ROOT, '.env');

/** Test prices – intentionally different from production values to verify write. */
const TEST_PRICES = {
  Aldi: {
    coordinates: { latitude: 46.8451, longitude: 16.8455 },
    prices: {
      'tojás':  99.0,   // prod: 450.0
      'kenyér': 199.0,  // prod: 450.0
      'tej':    149.0,  // prod: 280.0
    },
  },
  Interspar: {
    coordinates: { latitude: 46.8413, longitude: 16.8521 },
    prices: {
      'tojás':  109.0,  // prod: 520.0
      'kenyér': 209.0,  // prod: 520.0
      'tej':    159.0,  // prod: 310.0
    },
  },
};

// ── Helpers ───────────────────────────────────────────────────────────────────

function parseEnvFile(filePath) {
  if (!fs.existsSync(filePath)) return {};
  const env = {};
  for (const raw of fs.readFileSync(filePath, 'utf8').split('\n')) {
    const line = raw.trim();
    if (!line || line.startsWith('#')) continue;
    const eqIdx = line.indexOf('=');
    if (eqIdx === -1) continue;
    const key = line.slice(0, eqIdx).trim();
    let val = line.slice(eqIdx + 1).trim();
    const commentIdx = val.indexOf(' #');
    if (commentIdx !== -1) val = val.slice(0, commentIdx).trim();
    if ((val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))) {
      val = val.slice(1, -1);
    }
    env[key] = val;
  }
  return env;
}

function httpRequest(method, urlStr, headers, body) {
  return new Promise((resolve, reject) => {
    const url     = new URL(urlStr);
    const isHttps = url.protocol === 'https:';
    const lib     = isHttps ? https : http;
    const payload = body ? JSON.stringify(body) : '';

    const opts = {
      hostname: url.hostname,
      port:     url.port || (isHttps ? 443 : 80),
      path:     url.pathname + url.search,
      method,
      headers: {
        'Content-Type':   'application/json',
        'Content-Length': Buffer.byteLength(payload),
        ...headers,
      },
    };

    const req = lib.request(opts, (res) => {
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () => {
        try {
          resolve({ statusCode: res.statusCode, body: JSON.parse(data) });
        } catch {
          resolve({ statusCode: res.statusCode, body: data });
        }
      });
    });

    req.setTimeout(8000, () => {
      req.destroy(new Error('Request timed out after 8s'));
    });

    req.on('error', reject);
    if (payload) req.write(payload);
    req.end();
  });
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  // 1. Load config
  const env          = parseEnvFile(ENV_FILE);
  const ADMIN_TOKEN  = env.ADMIN_TOKEN;
  const BACKEND_HOST = (env.BACKEND_HOST || 'http://localhost:8080').replace(/\/$/, '');
  const TARGET_URL   = `${BACKEND_HOST}/api/v1/admin/prices`;

  if (!ADMIN_TOKEN) {
    console.error('❌  ADMIN_TOKEN is not set in .env');
    process.exit(1);
  }

  console.log('🧪  Smart Shopper – Webhook Simulation Test');
  console.log('──────────────────────────────────────────');
  console.log(`🎯  Target : ${TARGET_URL}`);
  console.log(`🔑  Token  : ${ADMIN_TOKEN.slice(0, 6)}${'*'.repeat(Math.max(0, ADMIN_TOKEN.length - 6))}`);
  console.log('\n📦  Test payload:');
  console.log(JSON.stringify(TEST_PRICES, null, 2));

  // 2. Send POST request
  console.log('\n📡  Sending POST request...');
  let res;
  try {
    res = await httpRequest(
      'POST',
      TARGET_URL,
      { 'X-Admin-Token': ADMIN_TOKEN },
      TEST_PRICES,
    );
  } catch (err) {
    console.error(`\n❌  Connection failed: ${err.message}`);
    console.error('    Is the backend running? Start with: go run ./cmd/app/main.go');
    console.error('    Or with Docker:         docker compose up -d');
    process.exit(1);
  }

  // 3. Validate response
  console.log(`\n📬  Response: HTTP ${res.statusCode}`);

  if (res.statusCode === 200) {
    console.log('✅  SUCCESS – Backend accepted the price update (HTTP 200 OK)');
    console.log('   Response body:', JSON.stringify(res.body));
  } else if (res.statusCode === 401) {
    console.error('❌  UNAUTHORIZED (401) – ADMIN_TOKEN mismatch.');
    console.error(`    Token sent: ${ADMIN_TOKEN}`);
    console.error('    Check ADMIN_TOKEN in .env vs backend config.');
    process.exit(1);
  } else {
    console.error(`❌  Unexpected status: HTTP ${res.statusCode}`);
    console.error('    Body:', JSON.stringify(res.body));
    process.exit(1);
  }

  // 4. Show what changed in prices.json
  const pricesFile = path.join(ROOT, 'internal', 'data', 'prices.json');
  if (fs.existsSync(pricesFile)) {
    const written = JSON.parse(fs.readFileSync(pricesFile, 'utf8'));
    console.log('\n📝  prices.json after update:');
    console.log(JSON.stringify(written, null, 2));
  }

  console.log('\n⚠️   Remember to restore original prices.json:');
  console.log('    git checkout -- internal/data/prices.json');
}

main().catch((err) => {
  console.error('Unexpected error:', err);
  process.exit(1);
});
