# [Protected] Subscriptions Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/subscriptions](#get-api-subscriptions) — Get all subscriptions
- [`GET` /api/subscriptions/by-username/{username}](#get-api-subscriptions-by-username-username) — Get subscription by username
- [`GET` /api/subscriptions/by-short-uuid/{shortUuid}](#get-api-subscriptions-by-short-uuid-shortuuid) — Get subscription by short uuid (protected route)
- [`GET` /api/subscriptions/by-uuid/{uuid}](#get-api-subscriptions-by-uuid-uuid) — Get subscription by uuid
- [`GET` /api/subscriptions/by-short-uuid/{shortUuid}/raw](#get-api-subscriptions-by-short-uuid-shortuuid-raw) — Get Raw Subscription by Short UUID
- [`GET` /api/subscriptions/subpage-config/{shortUuid}](#get-api-subscriptions-subpage-config-shortuuid) — Get Subpage Config by Short UUID
- [`GET` /api/subscriptions/connection-keys/{uuid}](#get-api-subscriptions-connection-keys-uuid) — Get connection keys (base64 format) by uuid

---

## `GET` /api/subscriptions

**Get all subscriptions**

**Operation ID**: `SubscriptionsController_getAllSubscriptions`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `size` | query | `number` | ❌ | Number of subscriptions to return, no more than 500 |
| `start` | query | `number` | ❌ | Start index (offset) of the users to return, default is 0 |

### Responses

#### Status `200`: Users fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions?size=25&start=0" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscriptions/by-username/{username}

**Get subscription by username**

**Operation ID**: `SubscriptionsController_getSubscriptionByUsername`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `username` | path | `string` | ✅ | Username of the user |

### Responses

#### Status `200`: Subscription fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `404`: User not found

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` (date-time) | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/by-username/<username>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscriptions/by-short-uuid/{shortUuid}

**Get subscription by short uuid (protected route)**

**Operation ID**: `SubscriptionsController_getSubscriptionByShortUuidProtected`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `shortUuid` | path | `string` | ✅ | Short uuid of the user |

### Responses

#### Status `200`: Subscription fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `404`: User not found

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` (date-time) | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/by-short-uuid/<shortUuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscriptions/by-uuid/{uuid}

**Get subscription by uuid**

**Operation ID**: `SubscriptionsController_getSubscriptionByUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | Uuid of the user |

### Responses

#### Status `200`: Subscription fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `404`: User not found

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` (date-time) | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/by-uuid/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscriptions/by-short-uuid/{shortUuid}/raw

**Get Raw Subscription by Short UUID**

**Operation ID**: `SubscriptionsController_getRawSubscriptionByShortUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `withDisabledHosts` | query | `boolean` | ❌ | Include disabled hosts in the subscription. Default is false. |
| `shortUuid` | path | `string` | ✅ | Short UUID of the user |

### Responses

#### Status `200`: Raw subscription fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/by-short-uuid/<shortUuid>/raw?withDisabledHosts=<withDisabledHosts>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscriptions/subpage-config/{shortUuid}

**Get Subpage Config by Short UUID**

**Operation ID**: `SubscriptionsController_getSubpageConfigByShortUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `shortUuid` | path | `string` | ✅ | Short UUID of the user |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `requestHeaders` | `object` | ✅ |  |

**Example Body**:

```json
{
  "requestHeaders": null
}
```

### Responses

#### Status `200`: Subpage config fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/subpage-config/<shortUuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "requestHeaders": null
}' \
```


---

## `GET` /api/subscriptions/connection-keys/{uuid}

**Get connection keys (base64 format) by uuid**

**Operation ID**: `SubscriptionsController_getConnectionKeysByUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: Connection keys fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | `object` | ✅ |  |

**Example Response**:

```json
{
  "response": null
}
```

#### Status `400`: Validation error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `message` | `string` | ❌ |  |
| `statusCode` | `number` | ❌ | Example: `400` |
| `errors` | array of `object` | ❌ | Example: `[{'validation': 'uuid', 'code': 'invalid_string', 'message': 'Invalid uuid', 'path': ['uuid']}]` |

**Example Response**:

```json
{
  "message": "string",
  "statusCode": 400,
  "errors": [
    {
      "validation": "uuid",
      "code": "invalid_string",
      "message": "Invalid uuid",
      "path": [
        "uuid"
      ]
    }
  ]
}
```

#### Status `500`: Server error

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `timestamp` | `string` | ❌ |  |
| `path` | `string` | ❌ |  |
| `message` | `string` | ❌ |  |
| `errorCode` | `string` | ❌ |  |

**Example Response**:

```json
{
  "timestamp": "string",
  "path": "string",
  "message": "string",
  "errorCode": "string"
}
```

### Usage Example

```bash
curl -X GET "http://localhost/api/subscriptions/connection-keys/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
