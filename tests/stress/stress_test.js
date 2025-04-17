import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 20, // количество виртуальных пользователей
    duration: '10s', // сколько длится тест
    thresholds: {
        http_req_failed: ['rate<0.0001'],      // не более 0.01% ошибок
        http_req_duration: ['p(99)<100'],      // 99% запросов быстрее 100мс
    },
};

// Выполняется один раз перед началом теста всеми VU
export function setup() {
    const loginPayload = JSON.stringify({ role: 'staff' });

    const loginRes = http.post('http://localhost:8080/dummyLogin', loginPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    const token = loginRes.json('token');
    if (!token) {
        throw new Error('Не удалось получить токен');
    }

    return { token };
}

export default function (data) {
    const url = 'http://localhost:8080/pvz' +
        '?startDate=2020-01-01T00:00:00Z' +
        '&endDate=2030-01-01T00:00:00Z' +
        '&page=1&limit=10';

    const res = http.get(url, {
        headers: {
            Authorization: `Bearer ${data.token}`,
        },
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 100ms': (r) => r.timings.duration < 100,
    });
}
