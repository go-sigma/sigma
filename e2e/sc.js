import http from "k6/http";
import { check } from 'k6';

export const options = {
  iterations: 3,
};

const username = 'ximager';
const password = 'ximager';

const host = "http://127.0.0.1:3000";

export default function () {
  let response = http.post(`${host}/user/login`, JSON.stringify({ username, password }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(response, {
    'user login status is 200': (r) => r.status === 200,
  });

  const token = JSON.parse(response.body).token;

  response = http.post(`${host}/namespaces/`, JSON.stringify({ "name": "test", "description": "test" }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': (r) => r.status === 201,
  });
  const namespaceId = JSON.parse(response.body).id;

  response = http.del(`${host}/namespaces/${namespaceId}`, null, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'delete namespace status is 204': (r) => r.status === 204,
  });

  let page_size = 100;
  let page_num = 1;
  response = http.get(`${host}/namespaces/?page_size=${page_size}&page_num=${page_num}`, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'list namespace status is 200': (r) => r.status === 200,
  });
}
