#!/usr/bin/env node
/**
 * deploy_n8n_workflow.js
 *
 * Deploys the Smart Shopper price-updater workflow to an n8n instance via
 * the n8n REST API.
 *
 * Required environment variables (read from ../.env):
 *   N8N_API_KEY  – n8n API key (Settings → API → Create API Key)
 *   N8N_HOST     – n8n base URL, e.g. http://localhost:5678 (default)
 *
 * Usage:
 *   node scripts/deploy_n8n_workflow.js
 */

'use strict';

const fs   = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');

// ── Config ────────────────────────────────────────────────────────────────────

const ROOT        = path.resolve(__dirname, '..');
const ENV_FILE    = path.join(ROOT, '.env');
const WORKFLOW_FILE = path.join(ROOT, 'internal', 'automation', 'n8n_price_updater_workflow.json');

// ── Helpers ───────────────────────────────────────────────────────────────────

/** Parse a .env file into a key→value map. Ignores comments and blank lines. */
function parseEnvFile(filePath) {
  if (!fs.existsSync(filePath)) return {};
  const env = {};
  const lines = fs.readFileSync(filePath, 'utf8').split('\n');
  for (const raw of lines) {
    const line = raw.trim();
    if (!line || line.startsWith('#')) continue;
    const eqIdx = line.indexOf('=');
    if (eqIdx === -1) continue;
    const key = line.slice(0, eqIdx).trim();
    let val = line.slice(eqIdx + 1).trim();
    // Strip inline comments (# after value)
    const commentIdx = val.indexOf(' #');
    if (commentIdx !== -1) val = val.slice(0, commentIdx).trim();
    // Strip surrounding quotes
    if ((val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))) {
      val = val.slice(1, -1);
    }
    env[key] = val;
  }
  return env;
}

/** Perform an HTTP/HTTPS request; returns { statusCode, body }. */
function request(method, urlStr, headers, body) {
  return new Promise((resolve, reject) => {
    const url    = new URL(urlStr);
    const isHttps = url.protocol === 'https:';
    const lib    = isHttps ? https : http;
    const payload = body ? JSON.stringify(body) : '';

    const opts = {
      hostname: url.hostname,
      port:     url.port || (isHttps ? 443 : 80),
      path:     url.pathname + url.search,
      method,
      headers: {
        'Content-Type':  'application/json',
        'Content-Length': Buffer.byteLength(payload),
        ...headers,
      },
    };

    const req = lib.request(opts, (res) => {
      let data = '';
      res.on('data', (chunk) => (data += chunk));
      res.on('end', () => {
        try {
          resolve({ statusCode: res.statusCode, body: JSON.parse(data) });
        } catch {
          resolve({ statusCode: res.statusCode, body: data });
        }
      });
    });

    req.on('error', reject);
    if (payload) req.write(payload);
    req.end();
  });
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  // 1. Load config from .env
  const env       = parseEnvFile(ENV_FILE);
  const N8N_HOST  = (env.N8N_HOST || 'http://localhost:5678').replace(/\/$/, '');
  const N8N_API_KEY = env.N8N_API_KEY;

  if (!N8N_API_KEY || N8N_API_KEY === '******') {
    console.error('❌  N8N_API_KEY is not set or is masked in .env');
    console.error('    Add a real API key: Settings → API in your n8n instance.');
    process.exit(1);
  }

  const authHeaders = { 'X-N8N-API-Key': N8N_API_KEY };

  // 2. Load workflow definition
  if (!fs.existsSync(WORKFLOW_FILE)) {
    console.error(`❌  Workflow file not found: ${WORKFLOW_FILE}`);
    process.exit(1);
  }
  const workflowDef = JSON.parse(fs.readFileSync(WORKFLOW_FILE, 'utf8'));
  console.log(`📂  Loaded workflow: "${workflowDef.name}"`);

  // 3. Create the workflow via n8n REST API
  console.log(`\n🚀  POST ${N8N_HOST}/api/v1/workflows  (create)`);
  let createRes;
  try {
    createRes = await request('POST', `${N8N_HOST}/api/v1/workflows`, authHeaders, workflowDef);
  } catch (err) {
    console.error(`❌  Cannot reach n8n at ${N8N_HOST}: ${err.message}`);
    console.error('    Ensure n8n is running and N8N_HOST is correct.');
    process.exit(1);
  }

  if (createRes.statusCode !== 200 && createRes.statusCode !== 201) {
    console.error(`❌  Workflow creation failed (HTTP ${createRes.statusCode}):`);
    console.error(JSON.stringify(createRes.body, null, 2));
    process.exit(1);
  }

  const workflowId = createRes.body.id;
  console.log(`✅  Workflow created  →  id: ${workflowId}`);

  // 4. Activate the workflow
  console.log(`\n▶️   POST ${N8N_HOST}/api/v1/workflows/${workflowId}/activate`);
  let activateRes;
  try {
    activateRes = await request(
      'POST',
      `${N8N_HOST}/api/v1/workflows/${workflowId}/activate`,
      authHeaders,
      null,
    );
  } catch (err) {
    console.error(`❌  Activation request failed: ${err.message}`);
    process.exit(1);
  }

  if (activateRes.statusCode !== 200) {
    console.error(`❌  Activation failed (HTTP ${activateRes.statusCode}):`);
    console.error(JSON.stringify(activateRes.body, null, 2));
    process.exit(1);
  }

  console.log(`✅  Workflow activated  →  active: ${activateRes.body.active}`);
  console.log('\n🎉  Deployment complete!');
  console.log(`    Workflow ID : ${workflowId}`);
  console.log(`    Name        : ${activateRes.body.name}`);
  console.log(`    Active      : ${activateRes.body.active}`);
  console.log(`    Next run    : daily at 02:00 (${workflowDef.nodes[0]?.parameters?.rule?.interval?.[0]?.expression || '0 2 * * *'})`);
}

main().catch((err) => {
  console.error('Unexpected error:', err);
  process.exit(1);
});
