import http from 'k6/http';
import { check, sleep } from 'k6';
import { Options } from 'k6/options';

export const options: Options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        'http_req_duration': ['p(95)<100'],
        'http_req_failed': ['rate<0.01'],
    },
};

// generate a fake IP address
function getRandomIP(): string {
    return `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`;
}

export default function () {
    const url = 'http://localhost:8080/api/v1/packing/calculate';

    const randomOrder: number = Math.floor(Math.random() * 10000) + 1;
    const payload: string = JSON.stringify({
        orderQuantity: randomOrder,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Forwarded-For': getRandomIP(),
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'returned a JSON body': (r) => r.body !== null && String(r.body).length > 0,
        'has packs array': (r) => r.json('packs') !== undefined,
    });

    sleep(1);
}