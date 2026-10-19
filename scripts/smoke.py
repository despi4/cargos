#!/usr/bin/env python3
"""Exercise the public API. Creates a uniquely named demo account and its data."""
import json
import os
import time
import uuid
from urllib.request import Request, urlopen

base = os.environ.get('BASE_URL', 'http://localhost:8080').rstrip('/')
token = ''


def call(method, path, payload=None, expected=200):
    data = None if payload is None else json.dumps(payload).encode()
    request = Request(base + path, data=data, method=method,
                      headers={'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token})
    with urlopen(request, timeout=60) as response:
        assert response.status == expected, (path, response.status)
        raw = response.read()
        return json.loads(raw) if raw else None


call('GET', '/readyz')
name = 'demo-' + uuid.uuid4().hex[:12]
email, password = name + '@example.com', 'cargos-demo-password'
call('POST', '/api/v1/auth/register', {'username': name, 'email': email, 'password': password}, 201)
tokens = call('POST', '/api/v1/auth/login', {'email': email, 'password': password})
token = tokens['access_token']
product = call('POST', '/api/v1/products', {'sku': 'TYRE-001', 'name': 'Winter tyre', 'unit': 'pcs', 'min_stock': 10, 'target_stock': 50}, 201)
supplier = call('POST', '/api/v1/suppliers', {'name': 'Demo supplier', 'email': 'sales@example.com', 'lead_days': 7}, 201)
movement = {'quantity': 2, 'reason': 'Opening stock', 'reference': 'opening-stock'}
call('POST', '/api/v1/products/' + product['id'] + '/movements', movement)
call('POST', '/api/v1/products/' + product['id'] + '/movements', movement)
call('POST', '/api/v1/demand', {'product_id': product['id'], 'quantity': 90, 'note': 'Sales history'}, 201)
prediction = call('POST', '/api/v1/predictions', {'window_days': 30, 'supplier_id': supplier['id']}, 201)
assert prediction['recommendations'][0]['suggested_quantity'] == 48, prediction
order = call('POST', '/api/v1/orders', {'product_id': product['id'], 'supplier_id': supplier['id'], 'quantity': 48, 'unit_price': 150000, 'currency': 'KZT'}, 201)
for status in ['ordered', 'received', 'received']:
    call('POST', '/api/v1/orders/' + order['id'] + '/status', {'status': status})
assert call('GET', '/api/v1/products/' + product['id'])['stock'] == 50
for _ in range(30):
    inbox = call('GET', '/api/v1/notifications')['items']
    if inbox:
        break
    time.sleep(1)
assert len(inbox) == 1, inbox
call('PATCH', '/api/v1/notifications/' + inbox[0]['id'], {'read': True})
tokens = call('POST', '/api/v1/auth/refresh', {'refresh_token': tokens['refresh_token']})
token = tokens['access_token']
call('GET', '/api/v1/auth/me')
call('POST', '/api/v1/auth/logout', expected=204)
print('OK: registration, login, inventory, idempotency, suppliers, demand, predictions, purchase receipt, notification, refresh and logout')
print('Demo account:', email)
