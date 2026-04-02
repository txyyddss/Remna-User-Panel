# Subscription Request History Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/subscription-request-history](#get-api-subscription-request-history) — Get all subscription request history
- [`GET` /api/subscription-request-history/stats](#get-api-subscription-request-history-stats) — Get subscription request history stats

---

## `GET` /api/subscription-request-history

**Get all subscription request history**

**Operation ID**: `UserSubscriptionRequestHistoryController_getSubscriptionRequestHistory`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `size` | query | `number` | ❌ | Page size for pagination |
| `start` | query | `number` | ❌ | Offset for pagination |

### Responses

#### Status `200`: Subscription request history fetched successfully

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
curl -X GET "http://localhost/api/subscription-request-history?size=<size>&start=<start>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/subscription-request-history/stats

**Get subscription request history stats**

**Operation ID**: `UserSubscriptionRequestHistoryController_getSubscriptionRequestHistoryStats`

**Authentication**: `Authorization`

### Responses

#### Status `200`: User subscription request history stats fetched successfully

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
curl -X GET "http://localhost/api/subscription-request-history/stats" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
