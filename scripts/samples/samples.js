import http from "k6/http";
import { check } from 'k6';

export const options = {
  iterations: 1,
};

const username = 'sigma';
const password = 'sigma';

const host = "https://sigma.tosone.cn";

export default function () {
  let response = http.post(`${host}/api/v1/users/login`, JSON.stringify({ username, password }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(response, {
    'user login status is 200': r => r.status === 200,
  });

  const token = JSON.parse(response.body).token;

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test", "description": "test" }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': r => r.status === 201,
  });

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test-size-limit", "description": "test size limit", "size_limit": 104857600 }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': r => r.status === 201,
  });

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test-repo-cnt-limit", "description": "test repo count limit", "repository_limit": 3 }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': r => r.status === 201,
  });

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test-tag-count-limit", "description": "test tag count limit", "tag_limit": 3 }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': r => r.status === 201,
  });

  response = http.post(`${host}/api/v1/namespaces/`, JSON.stringify({ "name": "test-all", "description": "test all", "tag_limit": 3 ,"repository_limit": 3,"size_limit": 104857600 }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });
  check(response, {
    'create namespace status is 201': r => r.status === 201,
  });
}
