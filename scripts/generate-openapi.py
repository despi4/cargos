#!/usr/bin/env python3
"""Generate the committed OpenAPI contract without third-party dependencies."""
import json
from pathlib import Path


def ref(name):
    return {'$ref': '#/components/schemas/' + name}


def string(**kw):
    return {'type': 'string', **kw}


def integer(**kw):
    return {'type': 'integer', **kw}


def obj(properties, required=None):
    schema = {'type': 'object', 'additionalProperties': False, 'properties': properties}
    fields = list(properties) if required is None else required
    if fields:
        schema['required'] = fields
    return schema


uid = string(format='uuid')
date = string(format='date-time')
quantity = integer(minimum=1, maximum=1000000000)
base = {'id': uid, 'created_at': date, 'updated_at': date}
product_input = {'sku': string(minLength=1, maxLength=64, example='TYRE-001'),
                 'name': string(minLength=1, maxLength=200, example='Winter tyre'),
                 'unit': string(minLength=1, maxLength=32, example='pcs'),
                 'min_stock': integer(minimum=0, example=10),
                 'target_stock': integer(minimum=0, maximum=1000000000, example=50)}
supplier_input = {'name': string(minLength=1, maxLength=200, example='Parts supplier'),
                  'email': string(format='email', example='sales@example.com'),
                  'lead_days': integer(minimum=0, maximum=365, example=7)}
order_input = {'product_id': uid, 'supplier_id': uid, 'quantity': quantity,
               'unit_price': integer(minimum=0, maximum=1000000000, description='Price per unit in minor currency units.', example=150000),
               'currency': string(pattern='^[A-Z]{3}$', example='KZT')}
demand_input = {'product_id': uid, 'quantity': quantity, 'occurred_at': date,
                'note': string(maxLength=500, example='Daily sales')}
schemas = {
    'Error': obj({'error': string()}),
    'RegisterInput': obj({'username': string(minLength=3, maxLength=100, example='buyer'),
                          'email': string(format='email', example='buyer@example.com'),
                          'password': string(minLength=8, maxLength=128, format='password', example='cargos-demo-password')}),
    'LoginInput': obj({'email': string(format='email', example='buyer@example.com'),
                       'password': string(format='password', example='cargos-demo-password')}),
    'RefreshInput': obj({'refresh_token': string()}),
    'Tokens': obj({'access_token': string(), 'refresh_token': string(),
                   'token_type': string(enum=['Bearer']), 'expires_in': integer(example=900)}),
    'User': obj({**base, 'username': string(), 'email': string(format='email')}),
    'ProductInput': obj(product_input),
    'Product': obj({**base, **product_input, 'stock': integer(minimum=0), 'active': {'type': 'boolean'}}),
    'MovementInput': obj({'quantity': integer(minimum=-1000000000, maximum=1000000000, description='Nonzero delta; negative values consume stock.', example=20),
                          'reason': string(minLength=1, maxLength=500, example='Initial stock'),
                          'reference': string(minLength=1, maxLength=200, example='opening-stock-001')}),
    'Movement': obj({**base, 'product_id': uid, 'quantity': integer(), 'reason': string(), 'reference': string(), 'balance': integer()}),
    'SupplierInput': obj(supplier_input),
    'Supplier': obj({**base, **supplier_input, 'active': {'type': 'boolean'}}),
    'OrderInput': obj(order_input),
    'Order': obj({**base, **order_input, 'status': string(enum=['draft', 'ordered', 'receiving', 'received', 'cancelled']), 'expected_at': date}),
    'OrderStatusInput': obj({'status': string(enum=['ordered', 'received', 'cancelled'])}),
    'DemandInput': obj(demand_input, ['product_id', 'quantity']),
    'Demand': obj({**base, **demand_input}),
    'PredictionInput': obj({'window_days': integer(minimum=1, maximum=365, default=30), 'supplier_id': uid}, []),
    'Recommendation': obj({'product_id': uid, 'product_name': string(), 'stock': integer(),
                            'daily_demand': {'type': 'number'}, 'reorder_point': integer(),
                            'suggested_quantity': integer(), 'supplier_id': uid, 'lead_days': integer()},
                           ['product_id', 'product_name', 'stock', 'daily_demand', 'reorder_point', 'suggested_quantity', 'lead_days']),
    'Prediction': obj({**base, 'window_days': integer(), 'recommendations': {'type': 'array', 'items': ref('Recommendation')}, 'notification_sent': {'type': 'boolean'}}),
    'Notification': obj({**base, 'event_id': uid, 'title': string(), 'message': string(), 'read': {'type': 'boolean'}}),
    'ReadInput': obj({'read': {'type': 'boolean'}}),
}
for name in ['Product', 'Movement', 'Supplier', 'Order', 'Demand', 'Prediction', 'Notification']:
    schemas[name + 'List'] = obj({'items': {'type': 'array', 'items': ref(name)},
                                  'limit': integer(), 'offset': integer()})

spec = {'openapi': '3.0.3', 'info': {'title': 'Cargos API', 'version': '1.0.0',
        'description': 'Inventory and procurement platform. Register and log in, then paste access_token into Authorize. Data is isolated by account. Quantities are integers. Prices use minor currency units. All lists use limit/offset. DELETE archives products/suppliers. Demand records statistics; inventory consumption is a separate negative movement. Notifications are delivered asynchronously to the in-app inbox.'},
        'servers': [{'url': '/', 'description': 'Current API Gateway'}],
        'tags': [{'name': tag} for tag in ['Auth', 'Inventory', 'Procurement', 'Demand', 'Prediction', 'Notification', 'System']],
        'security': [{'bearerAuth': []}], 'paths': {},
        'components': {'securitySchemes': {'bearerAuth': {'type': 'http', 'scheme': 'bearer', 'bearerFormat': 'JWT'}}, 'schemas': schemas}}


def operation(method, path, tag, summary, output=None, request=None, status=200, public=False, description=None):
    op = {'tags': [tag], 'summary': summary,
          'operationId': method + '_' + path.replace('/api/v1/', '').replace('/', '_').replace('{', '').replace('}', ''),
          'responses': {str(status): {'description': 'Successful operation'}}}
    if description:
        op['description'] = description
    if output:
        op['responses'][str(status)]['content'] = {'application/json': {'schema': ref(output)}}
    for code, title in [(400, 'Invalid input'), (401, 'Authentication required'), (404, 'Resource not found'), (409, 'Duplicate or invalid state'), (429, 'Authentication rate limit'), (500, 'Internal error'), (503, 'Dependency unavailable')]:
        op['responses'][str(code)] = {'description': title, 'content': {'application/json': {'schema': ref('Error')}}}
    if public:
        op['security'] = []
    if request:
        op['requestBody'] = {'required': True, 'content': {'application/json': {'schema': ref(request)}}}
    params = []
    if '{id}' in path:
        params.append({'name': 'id', 'in': 'path', 'required': True, 'schema': uid})
    if output and output.endswith('List'):
        params += [{'name': 'limit', 'in': 'query', 'schema': integer(minimum=1, maximum=500, default=100)},
                   {'name': 'offset', 'in': 'query', 'schema': integer(minimum=0, default=0)}]
    if params:
        op['parameters'] = params
    spec['paths'].setdefault(path, {})[method] = op


operation('post', '/api/v1/auth/register', 'Auth', 'Register account', 'User', 'RegisterInput', 201, True)
operation('post', '/api/v1/auth/login', 'Auth', 'Log in', 'Tokens', 'LoginInput', public=True)
operation('post', '/api/v1/auth/refresh', 'Auth', 'Rotate refresh token and revoke previous token pair', 'Tokens', 'RefreshInput', public=True)
operation('post', '/api/v1/auth/logout', 'Auth', 'Revoke current session', status=204)
operation('get', '/api/v1/auth/me', 'Auth', 'Get current account', 'User')
for path, model, tag in [('products', 'Product', 'Inventory'), ('suppliers', 'Supplier', 'Procurement'), ('orders', 'Order', 'Procurement'), ('demand', 'Demand', 'Demand'), ('predictions', 'Prediction', 'Prediction'), ('notifications', 'Notification', 'Notification')]:
    operation('get', '/api/v1/' + path, tag, 'List ' + path, model + 'List')
    operation('get', '/api/v1/' + path + '/{id}', tag, 'Get ' + model.lower(), model)
    if path != 'notifications':
        desc = None
        if path == 'demand':
            desc = 'Statistical record only; does not decrement inventory. occurred_at defaults to current time. Record consumption with a negative stock movement.'
        if path == 'predictions':
            desc = 'Uses average demand over window_days. Reorder point=max(min_stock,ceil(daily_demand*lead_days)); target=max(target_stock,ceil(daily_demand*(lead_days+7))). Ordered incoming stock is deducted. Uses the selected active supplier or the fastest active supplier; default lead is 7 days. Saves a durable event; notification delivery retries every 2 seconds.'
        operation('post', '/api/v1/' + path, tag, 'Create ' + model.lower(), model, model + 'Input', 201, description=desc)
for path, model, tag in [('products', 'Product', 'Inventory'), ('suppliers', 'Supplier', 'Procurement')]:
    operation('put', '/api/v1/' + path + '/{id}', tag, 'Update ' + model.lower(), model, model + 'Input', description='Full metadata replacement. Product SKU is immutable. Stock changes require a movement.')
    operation('delete', '/api/v1/' + path + '/{id}', tag, 'Archive ' + model.lower(), status=204)
operation('get', '/api/v1/movements', 'Inventory', 'List stock ledger', 'MovementList')
operation('post', '/api/v1/products/{id}/movements', 'Inventory', 'Adjust stock with idempotent reference', 'Movement', 'MovementInput', description='The reference is unique per account. Repeating the same product, quantity, reason and reference returns the original movement. Reusing a reference with different data returns 409. Negative balances are rejected. Stock and ledger commit in one transaction.')
operation('post', '/api/v1/orders/{id}/status', 'Procurement', 'Change purchase status', 'Order', 'OrderStatusInput', description='draft → ordered/cancelled; ordered → receiving → received, or ordered → cancelled. Repeating the current state succeeds. Receiving adds inventory with reference purchase:{order_id}; retries cannot add stock twice. The receiving state is persisted before crediting stock; a background worker retries every 2 seconds after transient errors. Cancellation is blocked during receiving. You can also repeat the received request. expected_at is an estimate set at order creation.')
operation('patch', '/api/v1/notifications/{id}', 'Notification', 'Mark notification as read or unread', 'Notification', 'ReadInput')
for path, description in [('/healthz', 'Gateway liveness'), ('/readyz', 'Readiness of all six services and their databases')]:
    operation('get', path, 'System', description, public=True)
    spec['paths'][path]['get']['responses']['200']['content'] = {'application/json': {'schema': {'type': 'object', 'additionalProperties': {'type': 'string'}}}}
operation('get', '/metrics', 'System', 'Prometheus request counter', public=True)
spec['paths']['/metrics']['get']['responses']['200']['content'] = {'text/plain': {'schema': string()}}
Path(__file__).resolve().parents[1].joinpath('services/api_gateway/docs/openapi.json').write_text(json.dumps(spec, indent=2, ensure_ascii=False) + '\n')
