import http from "k6/http";
import { check } from 'k6';

export const options = {
  iterations: 3,
};

const username = 'ximager';
const password = 'ximager';

const host = "http://127.0.0.1:3000";

export default function () {
  let response = http.post(`${host}/api/v1/users/login`, JSON.stringify({ username, password }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(response, {
    'user login status is 200': (r) => r.status === 200,
  });

  const token = JSON.parse(response.body).token;

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test", "description": "test" }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': (r) => r.status === 201,
  });
  const namespaceId = JSON.parse(response.body).id;

  response = http.del(`${host}/api/v1/namespaces/${namespaceId}`, null, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'delete namespace status is 204': (r) => r.status === 204,
  });

  let limit = 100;
  let last = 0;
  response = http.get(`${host}/api/v1/namespaces/?limit=${limit}&last=${last}`, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'list namespace status is 200': (r) => r.status === 200,
  });
}
