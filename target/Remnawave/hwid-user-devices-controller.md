# HWID User Devices Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/hwid/devices](#get-api-hwid-devices) — Get all HWID devices
- [`POST` /api/hwid/devices](#post-api-hwid-devices) — Create a user HWID device
- [`POST` /api/hwid/devices/delete](#post-api-hwid-devices-delete) — Delete a user HWID device
- [`POST` /api/hwid/devices/delete-all](#post-api-hwid-devices-delete-all) — Delete all user HWID devices
- [`GET` /api/hwid/devices/stats](#get-api-hwid-devices-stats) — Get HWID devices stats
- [`GET` /api/hwid/devices/top-users](#get-api-hwid-devices-top-users) — Get top users by HWID devices
- [`GET` /api/hwid/devices/{userUuid}](#get-api-hwid-devices-useruuid) — Get user HWID devices

---

## `GET` /api/hwid/devices

**Get all HWID devices**

**Operation ID**: `HwidUserDevicesController_getAllUsers`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `size` | query | `number` | ❌ | Page size for pagination |
| `start` | query | `number` | ❌ | Offset for pagination |

### Responses

#### Status `200`: Hwid devices fetched successfully

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
curl -X GET "http://localhost/api/hwid/devices?size=<size>&start=<start>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/hwid/devices

**Create a user HWID device**

**Operation ID**: `HwidUserDevicesController_createUserHwidDevice`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `hwid` | `string` | ✅ |  |
| `userUuid` | `string` (uuid) | ✅ |  |
| `platform` | `string` | ❌ |  |
| `osVersion` | `string` | ❌ |  |
| `deviceModel` | `string` | ❌ |  |
| `userAgent` | `string` | ❌ |  |

**Example Body**:

```json
{
  "hwid": "string",
  "userUuid": "550e8400-e29b-41d4-a716-446655440000",
  "platform": "string",
  "osVersion": "string",
  "deviceModel": "string",
  "userAgent": "string"
}
```

### Responses

#### Status `200`: User HWID device created successfully

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

#### Status `404`: One of requested resources not found

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
curl -X POST "http://localhost/api/hwid/devices" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "hwid": "string",
  "userUuid": "550e8400-e29b-41d4-a716-446655440000",
  "platform": "string",
  "osVersion": "string",
  "deviceModel": "string",
  "userAgent": "string"
}' \
```


---

## `POST` /api/hwid/devices/delete

**Delete a user HWID device**

**Operation ID**: `HwidUserDevicesController_deleteUserHwidDevice`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `userUuid` | `string` (uuid) | ✅ |  |
| `hwid` | `string` | ✅ |  |

**Example Body**:

```json
{
  "userUuid": "550e8400-e29b-41d4-a716-446655440000",
  "hwid": "string"
}
```

### Responses

#### Status `200`: User HWID device deleted successfully

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

#### Status `404`: One of requested resources not found

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
curl -X POST "http://localhost/api/hwid/devices/delete" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "userUuid": "550e8400-e29b-41d4-a716-446655440000",
  "hwid": "string"
}' \
```


---

## `POST` /api/hwid/devices/delete-all

**Delete all user HWID devices**

**Operation ID**: `HwidUserDevicesController_deleteAllUserHwidDevices`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `userUuid` | `string` (uuid) | ✅ |  |

**Example Body**:

```json
{
  "userUuid": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Responses

#### Status `200`: User HWID devices deleted successfully

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

#### Status `404`: One of requested resources not found

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
curl -X POST "http://localhost/api/hwid/devices/delete-all" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "userUuid": "550e8400-e29b-41d4-a716-446655440000"
}' \
```


---

## `GET` /api/hwid/devices/stats

**Get HWID devices stats**

**Operation ID**: `HwidUserDevicesController_getHwidDevicesStats`

**Authentication**: `Authorization`

### Responses

#### Status `200`: Hwid devices stats fetched successfully

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
curl -X GET "http://localhost/api/hwid/devices/stats" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/hwid/devices/top-users

**Get top users by HWID devices**

**Operation ID**: `HwidUserDevicesController_getTopUsersByHwidDevices`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `size` | query | `number` | ❌ | Page size for pagination |
| `start` | query | `number` | ❌ | Offset for pagination |

### Responses

#### Status `200`: Top users by HWID devices fetched successfully

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
curl -X GET "http://localhost/api/hwid/devices/top-users?size=<size>&start=<start>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/hwid/devices/{userUuid}

**Get user HWID devices**

**Operation ID**: `HwidUserDevicesController_getUserHwidDevices`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userUuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User HWID devices fetched successfully

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

#### Status `404`: One of requested resources not found

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
curl -X GET "http://localhost/api/hwid/devices/<userUuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
