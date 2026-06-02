import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const ACCESS_TOKEN = __ENV.ACCESS_TOKEN || '';

export const options = {
  scenarios: {
    create_payment: {
      executor: 'constant-vus',
      vus: 5,
      duration: '30s',
      exec: 'createPayment',
    },
    duplicate_idempotency_key: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 10,
      exec: 'duplicateIdempotencyKey',
      startTime: '35s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
};

function authHeaders(extraHeaders = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...extraHeaders,
  };
  if (ACCESS_TOKEN) {
    headers.Authorization = `Bearer ${ACCESS_TOKEN}`;
  }
  return headers;
}

export function createPayment() {
  const payload = JSON.stringify({
    amount: 10.5,
    currency: 'USD',
    provider: 'stripe',
  });

  const res = http.post(`${BASE_URL}/payments`, payload, {
    headers: authHeaders({
      'Idempotency-Key': uuidv4(),
    }),
  });

  check(res, {
    'create payment status 201 or 200': (r) => r.status === 201 || r.status === 200,
    'create payment has id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined && body.id !== '';
      } catch (_) {
        return false;
      }
    },
  });

  sleep(0.5);
}

export function duplicateIdempotencyKey() {
  const idempotencyKey = `load-test-${uuidv4()}`;
  const payload = JSON.stringify({
    amount: 25.0,
    currency: 'USD',
    provider: 'stripe',
  });
  const headers = authHeaders({ 'Idempotency-Key': idempotencyKey });

  const first = http.post(`${BASE_URL}/payments`, payload, { headers });
  const second = http.post(`${BASE_URL}/payments`, payload, { headers });

  check(first, {
    'first request succeeds': (r) => r.status === 201 || r.status === 200,
  });

  check(second, {
    'duplicate idempotency returns same payment': (r) => {
      if (first.status !== 201 && first.status !== 200) {
        return true;
      }
      if (r.status !== 201 && r.status !== 200) {
        return false;
      }
      try {
        const firstBody = JSON.parse(first.body);
        const secondBody = JSON.parse(r.body);
        return firstBody.id === secondBody.id;
      } catch (_) {
        return false;
      }
    },
  });

  sleep(0.2);
}

export function setup() {
  if (!ACCESS_TOKEN) {
    console.warn('ACCESS_TOKEN not set; authenticated routes may return 401');
  }
  return { baseUrl: BASE_URL };
}

export default function () {
  createPayment();
}
