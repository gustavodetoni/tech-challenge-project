import http from 'k6/http';
import { check, group, sleep } from 'k6';

const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const REGISTER_VUS = Number(__ENV.REGISTER_VUS || __ENV.VUS || 40);
const HEALTH_RPS = Number(__ENV.HEALTH_RPS || 30);
const DURATION = __ENV.DURATION || '5m';
const RAMP_UP = __ENV.RAMP_UP || '10s';
const RAMP_DOWN = __ENV.RAMP_DOWN || '20s';
const REGISTER_ENABLED = (__ENV.REGISTER_ENABLED || 'true').toLowerCase() !== 'false';
const SLEEP_SECONDS = Number(__ENV.SLEEP_SECONDS || 0.2);
const EMAIL_DOMAIN = __ENV.EMAIL_DOMAIN || 'k6.local';
const PASSWORD = __ENV.PASSWORD || 'Senha@123';
const RUN_ID = __ENV.RUN_ID || `${Date.now()}`;

export const options = {
  scenarios: {
    health_probe: {
      executor: 'constant-arrival-rate',
      rate: HEALTH_RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.max(10, Math.ceil(HEALTH_RPS / 2)),
      maxVUs: Math.max(20, HEALTH_RPS * 2),
      exec: 'health',
    },
    register_cpu: {
      executor: 'ramping-vus',
      exec: 'register',
      startVUs: 0,
      stages: [
        { duration: RAMP_UP, target: REGISTER_ENABLED ? REGISTER_VUS : 0 },
        { duration: DURATION, target: REGISTER_ENABLED ? REGISTER_VUS : 0 },
        { duration: RAMP_DOWN, target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.30'],
    checks: ['rate>0.70'],
  },
};

function userPayload() {
  const unique = `${RUN_ID}-${__VU}-${__ITER}`;
  return {
    name: `K6 User ${unique}`,
    email: `k6-user-${unique}@${EMAIL_DOMAIN}`,
    password: PASSWORD,
  };
}

export function health() {
  group('healthchecks', () => {
    const res = http.get(`${BASE_URL}/health`, {
      timeout: '10s',
      tags: { name: 'GET /health' },
    });

    check(res, {
      'health status is 200': (r) => r && r.status === 200,
    });
  });
}

export function register() {
  if (!REGISTER_ENABLED) {
    sleep(SLEEP_SECONDS);
    return;
  }

  group('account creation', () => {
    const res = http.post(`${BASE_URL}/auth/register`, JSON.stringify(userPayload()), {
      headers: { 'Content-Type': 'application/json' },
      timeout: '30s',
      tags: { name: 'POST /auth/register' },
    });

    check(res, {
      'register status is 201': (r) => r && r.status === 201,
      'register returned access token': (r) => {
        if (!r || r.status !== 201 || !r.body) {
          return false;
        }

        const body = r.json();
        return Boolean(body && body.access_token);
      },
    });
  });

  sleep(SLEEP_SECONDS);
}

export default register;
