# IP Management Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`POST` /api/ip-control/fetch-ips/{uuid}](#post-api-ip-control-fetch-ips-uuid) — Request IP List for User
- [`GET` /api/ip-control/fetch-ips/result/{jobId}](#get-api-ip-control-fetch-ips-result-jobid) — Get IP List Result by Job ID
- [`POST` /api/ip-control/drop-connections](#post-api-ip-control-drop-connections) — Drop Connections for Users or IPs

---

## `POST` /api/ip-control/fetch-ips/{uuid}

**Request IP List for User**

**Operation ID**: `IpControlController_fetchUserIps`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: Return jobId for further processing

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
curl -X POST "http://localhost/api/ip-control/fetch-ips/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/ip-control/fetch-ips/result/{jobId}

**Get IP List Result by Job ID**

**Operation ID**: `IpControlController_getFetchIpsResult`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `jobId` | path | `string` | ✅ | Job ID |

### Responses

#### Status `200`: Return result or status of the job

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

#### Status `404`: Job not found

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
curl -X GET "http://localhost/api/ip-control/fetch-ips/result/<jobId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/ip-control/drop-connections

**Drop Connections for Users or IPs**

**Operation ID**: `IpControlController_dropConnections`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `dropBy` | `object` \| `object` | ✅ |  |
| `targetNodes` | `object` \| `object` | ✅ |  |

**Example Body**:

```json
{
  "dropBy": null,
  "targetNodes": null
}
```

### Responses

#### Status `200`: Event sent to background executor

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

#### Status `404`: User not found // Connected nodes not found

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
curl -X POST "http://localhost/api/ip-control/drop-connections" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "dropBy": null,
  "targetNodes": null
}' \
```


---
