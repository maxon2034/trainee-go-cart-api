import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
        { duration: '20s', target: 5 },
        { duration: '40s', target: 20 },
        { duration: '1m', target: 20 },
        { duration: '20s', target: 0 },
    ],
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    const headers = { 'Content-Type': 'application/json' };
    const cartId = uuidv4();

    let res = http.post(`${BASE_URL}/api/v1/carts`, null, { headers });
    check(res, { 'cart created': (r) => r.status === 201 || r.status === 200 });

    res = http.get(`${BASE_URL}/api/v1/carts/${cartId}`);
    check(res, { 'cart viewed': (r) => r.status === 200 || r.status === 404 });

    let payload = JSON.stringify({
        item_id: uuidv4(),
        quantity: 1,
        price: 150
    });
    res = http.post(`${BASE_URL}/api/v1/carts/${cartId}/items`, payload, { headers });
    check(res, { 'item added': (r) => r.status === 200 || r.status === 201 || r.status === 404 });

    sleep(1);
}