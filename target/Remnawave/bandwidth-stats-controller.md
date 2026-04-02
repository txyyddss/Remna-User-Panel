# Bandwidth Stats Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/bandwidth-stats/nodes/{uuid}/users/legacy](#get-api-bandwidth-stats-nodes-uuid-users-legacy) — Get Node User Usage by Range and Node UUID (Legacy)
- [`GET` /api/bandwidth-stats/nodes/realtime](#get-api-bandwidth-stats-nodes-realtime) — Get Nodes Realtime Usage
- [`GET` /api/bandwidth-stats/nodes/{uuid}/users](#get-api-bandwidth-stats-nodes-uuid-users) — Get Node Users Usage by Node UUID
- [`GET` /api/bandwidth-stats/users/{uuid}/legacy](#get-api-bandwidth-stats-users-uuid-legacy) — Get User Usage by Range (Legacy)
- [`GET` /api/bandwidth-stats/users/{uuid}](#get-api-bandwidth-stats-users-uuid) — Get User Usage by Range
- [`GET` /api/bandwidth-stats/nodes](#get-api-bandwidth-stats-nodes) — Get Nodes Usage by Range

---

## `GET` /api/bandwidth-stats/nodes/{uuid}/users/legacy

**Get Node User Usage by Range and Node UUID (Legacy)**

**Operation ID**: `BandwidthStatsNodesController_getNodeUserUsage`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `start` | query | `string` (date-time) | ✅ | Start date |
| `end` | query | `string` (date-time) | ✅ | End date |
| `uuid` | path | `string` | ✅ | UUID of the node |

### Responses

#### Status `200`: Nodes users usage by range (legacy) fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "userUuid": "550e8400-e29b-41d4-a716-446655440000",
      "username": "string",
      "nodeUuid": "550e8400-e29b-41d4-a716-446655440000",
      "total": 0,
      "date": "string"
    }
  ]
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
curl -X GET "http://localhost/api/bandwidth-stats/nodes/<uuid>/users/legacy?start=<start>&end=<end>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/bandwidth-stats/nodes/realtime

**Get Nodes Realtime Usage**

**Operation ID**: `BandwidthStatsNodesController_getNodesRealtimeUsage`

**Authentication**: `Authorization`

### Responses

#### Status `200`: Nodes realtime usage fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "nodeUuid": "550e8400-e29b-41d4-a716-446655440000",
      "nodeName": "string",
      "countryCode": "string",
      "downloadBytes": 0,
      "uploadBytes": 0,
      "totalBytes": 0,
      "downloadSpeedBps": 0,
      "uploadSpeedBps": 0,
      "totalSpeedBps": 0
    }
  ]
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
curl -X GET "http://localhost/api/bandwidth-stats/nodes/realtime" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/bandwidth-stats/nodes/{uuid}/users

**Get Node Users Usage by Node UUID**

**Operation ID**: `BandwidthStatsNodesController_getStatsNodeUsersUsage`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `topUsersLimit` | query | `number` | ✅ | Limit of top users to return |
| `start` | query | `string` (date) | ✅ | Start date (YYYY-MM-DD) |
| `end` | query | `string` (date) | ✅ | End date (YYYY-MM-DD) |
| `uuid` | path | `string` | ✅ | UUID of the node |

### Responses

#### Status `200`: Stats node users usage fetched successfully

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
curl -X GET "http://localhost/api/bandwidth-stats/nodes/<uuid>/users?topUsersLimit=<topUsersLimit>&start=2026-01-31&end=2026-01-01" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/bandwidth-stats/users/{uuid}/legacy

**Get User Usage by Range (Legacy)**

**Operation ID**: `BandwidthStatsUsersController_getUserUsageByRange`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `start` | query | `string` (date-time) | ✅ | Start date |
| `end` | query | `string` (date-time) | ✅ | End date |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User usage by range (legacy) fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "userUuid": "550e8400-e29b-41d4-a716-446655440000",
      "nodeUuid": "550e8400-e29b-41d4-a716-446655440000",
      "nodeName": "string",
      "countryCode": "string",
      "total": 0,
      "date": "string"
    }
  ]
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
curl -X GET "http://localhost/api/bandwidth-stats/users/<uuid>/legacy?start=<start>&end=<end>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/bandwidth-stats/users/{uuid}

**Get User Usage by Range**

**Operation ID**: `BandwidthStatsUsersController_getStatsNodesUsage`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `topNodesLimit` | query | `number` | ✅ | Limit of top nodes to return |
| `start` | query | `string` (date) | ✅ | Start date (YYYY-MM-DD) |
| `end` | query | `string` (date) | ✅ | End date (YYYY-MM-DD) |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: Stats user usage fetched successfully

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
curl -X GET "http://localhost/api/bandwidth-stats/users/<uuid>?topNodesLimit=<topNodesLimit>&start=2026-01-31&end=2026-01-01" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/bandwidth-stats/nodes

**Get Nodes Usage by Range**

**Operation ID**: `NodesUsageHistoryController_getStatsNodesUsage`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `topNodesLimit` | query | `number` | ✅ | Limit of top nodes to return |
| `start` | query | `string` (date) | ✅ | Start date (YYYY-MM-DD) |
| `end` | query | `string` (date) | ✅ | End date (YYYY-MM-DD) |

### Responses

#### Status `200`: Stats nodes usage fetched successfully

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
curl -X GET "http://localhost/api/bandwidth-stats/nodes?topNodesLimit=<topNodesLimit>&start=2026-01-31&end=2026-01-01" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
