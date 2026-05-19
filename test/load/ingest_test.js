import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// ── Metrics ───────────────────────────────────────────────────
const ingestErrors = new Counter('ingest_errors');
const ingestLatency = new Trend('ingest_latency_ms', true);

// ── Config ────────────────────────────────────────────────────
// Jalankan: k6 run test/load/ingest_test.js \
//   --env BASE_URL=http://localhost:8080 \
//   --env API_KEY=<your_api_key> \
//   --env DEVICE_ID=<your_device_id>

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY  = __ENV.API_KEY  || 'your-api-key-here';
const DEVICE_ID = __ENV.DEVICE_ID || 'your-device-id-here';

// ── Load profile ──────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '30s', target: 50  },  // ramp up ke 50 VU
    { duration: '2m',  target: 200 },  // sustained 200 VU
    { duration: '30s', target: 500 },  // spike ke 500 VU
    { duration: '1m',  target: 200 },  // turun balik
    { duration: '30s', target: 0   },  // ramp down
  ],
  thresholds: {
    http_req_duration:        ['p(95)<300'],  // 95% request < 300ms
    http_req_failed:          ['rate<0.01'],  // error rate < 1%
    ingest_latency_ms:        ['p(99)<500'],  // 99th percentile < 500ms
  },
};

// ── Test scenario ─────────────────────────────────────────────
export default function () {
  const payload = JSON.stringify({
    temperature: (20 + Math.random() * 15).toFixed(1),
    humidity:    (40 + Math.random() * 40).toFixed(1),
    battery:     Math.floor(70 + Math.random() * 30),
    ts:          new Date().toISOString(),
  });

  const res = http.post(
    `${BASE_URL}/api/v1/ingest/${DEVICE_ID}`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': API_KEY,
      },
      timeout: '5s',
    }
  );

  ingestLatency.add(res.timings.duration);

  const ok = check(res, {
    'status is 201':          (r) => r.status === 201,
    'response has status ok': (r) => {
      try { return JSON.parse(r.body).data.status === 'ok'; }
      catch { return false; }
    },
  });

  if (!ok) {
    ingestErrors.add(1);
  }

  sleep(0.1); // 100ms antar request per VU
}

// ── Setup: cek endpoint sebelum load test ─────────────────────
export function setup() {
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200) {
    throw new Error(`Server not healthy: ${res.status}`);
  }
  console.log(`Load test target: ${BASE_URL}`);
  console.log(`Device ID: ${DEVICE_ID}`);
}
