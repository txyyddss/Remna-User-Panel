# Users Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`POST` /api/users](#post-api-users) — Create a new user
- [`PATCH` /api/users](#patch-api-users) — Update a user by UUID or username
- [`GET` /api/users](#get-api-users) — Get all users
- [`DELETE` /api/users/{uuid}](#delete-api-users-uuid) — Delete user
- [`GET` /api/users/{uuid}](#get-api-users-uuid) — Get user by UUID
- [`GET` /api/users/tags](#get-api-users-tags) — Get all existing user tags
- [`GET` /api/users/{uuid}/accessible-nodes](#get-api-users-uuid-accessible-nodes) — Get user accessible nodes
- [`GET` /api/users/{uuid}/subscription-request-history](#get-api-users-uuid-subscription-request-history) — Get user subscription request history, recent 24 records
- [`GET` /api/users/by-short-uuid/{shortUuid}](#get-api-users-by-short-uuid-shortuuid) — Get user by Short UUID
- [`GET` /api/users/by-username/{username}](#get-api-users-by-username-username) — Get user by username
- [`GET` /api/users/by-id/{id}](#get-api-users-by-id-id) — Get user by ID
- [`GET` /api/users/by-telegram-id/{telegramId}](#get-api-users-by-telegram-id-telegramid) — Get users by telegram ID
- [`GET` /api/users/by-email/{email}](#get-api-users-by-email-email) — Get users by email
- [`GET` /api/users/by-tag/{tag}](#get-api-users-by-tag-tag) — Get users by tag
- [`POST` /api/users/{uuid}/actions/revoke](#post-api-users-uuid-actions-revoke) — Revoke user subscription
- [`POST` /api/users/{uuid}/actions/disable](#post-api-users-uuid-actions-disable) — Disable user
- [`POST` /api/users/{uuid}/actions/enable](#post-api-users-uuid-actions-enable) — Enable user
- [`POST` /api/users/{uuid}/actions/reset-traffic](#post-api-users-uuid-actions-reset-traffic) — Reset user traffic

---

## `POST` /api/users

**Create a new user**

**Operation ID**: `UsersController_createUser`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `username` | `string` | ✅ | Unique username for the user. Required. Must be 3-36 characters long and contain only letters, numbers, underscores and dashes. |
| `status` | `string` | ❌ | Optional. User account status. Defaults to ACTIVE. Enum: [`ACTIVE`, `DISABLED`, `LIMITED`, `EXPIRED`] |
| `shortUuid` | `string` | ❌ | Optional. Short UUID identifier for the user. |
| `trojanPassword` | `string` | ❌ | Optional. Password for Trojan protocol. Must be 8-32 characters. |
| `vlessUuid` | `string` (uuid) | ❌ | Optional. UUID for VLESS protocol. Must be a valid UUID format. |
| `ssPassword` | `string` | ❌ | Optional. Password for Shadowsocks protocol. Must be 8-32 characters. |
| `trafficLimitBytes` | `integer` | ❌ | Optional. Traffic limit in bytes. Set to 0 for unlimited traffic. |
| `trafficLimitStrategy` | `string` | ❌ | Available reset periods Enum: [`NO_RESET`, `DAY`, `WEEK`, `MONTH`] |
| `expireAt` | `string` (date-time) | ✅ | Account expiration date. Required. Format: 2025-01-17T15:38:45.065Z |
| `createdAt` | `string` (date-time) | ❌ | Optional. Account creation date. Format: 2025-01-17T15:38:45.065Z |
| `lastTrafficResetAt` | `string` (date-time) | ❌ | Optional. Date of last traffic reset. Format: 2025-01-17T15:38:45.065Z |
| `description` | `string` | ❌ | Optional. Additional notes or description for the user account. |
| `tag` | `string` | ❌ | Optional. User tag for categorization. Max 16 characters, uppercase letters, numbers and underscores only. |
| `telegramId` | `integer` | ❌ | Optional. Telegram user ID for notifications. Must be an integer. |
| `email` | `string` (email) | ❌ | Optional. User email address. Must be a valid email format. |
| `hwidDeviceLimit` | `integer` | ❌ | Optional. Maximum number of hardware devices allowed. Must be a positive integer. |
| `activeInternalSquads` | array of `string` (uuid) | ❌ | Optional. Array of UUIDs representing enabled internal squads. |
| `uuid` | `string` (uuid) | ❌ | Optional. Pass UUID to create user with specific UUID, otherwise it will be generated automatically. |
| `externalSquadUuid` | `string` (uuid) | ❌ | Optional. External squad UUID. |

**Example Body**:

```json
{
  "username": "string",
  "status": "ACTIVE",
  "shortUuid": "string",
  "trojanPassword": "string",
  "vlessUuid": "550e8400-e29b-41d4-a716-446655440000",
  "ssPassword": "string",
  "trafficLimitBytes": 0,
  "trafficLimitStrategy": "NO_RESET",
  "expireAt": "2024-01-01T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "lastTrafficResetAt": "2024-01-01T00:00:00Z",
  "description": "string",
  "tag": "string",
  "telegramId": 0,
  "email": "string"
}
```

### Responses

#### Status `201`: User created successfully

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
curl -X POST "http://localhost/api/users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "username": "string",
  "status": "ACTIVE",
  "shortUuid": "string",
  "trojanPassword": "string",
  "vlessUuid": "550e8400-e29b-41d4-a716-446655440000",
  "ssPassword": "string",
  "trafficLimitBytes": 0,
  "trafficLimitStrategy": "NO_RESET",
  "expireAt": "2024-01-01T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "lastTrafficResetAt": "2024-01-01T00:00:00Z",
  "description": "string",
  "tag": "string",
  "telegramId": 0,
  "email": "string"
}' \
```


---

## `PATCH` /api/users

**Update a user by UUID or username**

**Operation ID**: `UsersController_updateUser`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `username` | `string` | ❌ | Username of the user |
| `uuid` | `string` (uuid) | ❌ | UUID of the user. UUID has higher priority than username, so if both are provided, username will be ignored. |
| `status` | `string` | ❌ |  Enum: [`ACTIVE`, `DISABLED`] |
| `trafficLimitBytes` | `integer` | ❌ | Traffic limit in bytes. 0 - unlimited |
| `trafficLimitStrategy` | `string` | ❌ | Available reset periods Enum: [`NO_RESET`, `DAY`, `WEEK`, `MONTH`] |
| `expireAt` | `string` (date-time) | ❌ | Expiration date: 2025-01-17T15:38:45.065Z |
| `description` | `string` | ❌ |  |
| `tag` | `string` | ❌ |  |
| `telegramId` | `integer` | ❌ |  |
| `email` | `string` (email) | ❌ |  |
| `hwidDeviceLimit` | `integer` | ❌ |  |
| `activeInternalSquads` | array of `string` (uuid) | ❌ |  |
| `externalSquadUuid` | `string` (uuid) | ❌ | Optional. External squad UUID. |

**Example Body**:

```json
{
  "username": "string",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "status": "ACTIVE",
  "trafficLimitBytes": 0,
  "trafficLimitStrategy": "NO_RESET",
  "expireAt": "2024-01-01T00:00:00Z",
  "description": "string",
  "tag": "string",
  "telegramId": 0,
  "email": "string",
  "hwidDeviceLimit": 0,
  "activeInternalSquads": [
    "string"
  ],
  "externalSquadUuid": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Responses

#### Status `200`: User updated successfully

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
curl -X PATCH "http://localhost/api/users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "username": "string",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "status": "ACTIVE",
  "trafficLimitBytes": 0,
  "trafficLimitStrategy": "NO_RESET",
  "expireAt": "2024-01-01T00:00:00Z",
  "description": "string",
  "tag": "string",
  "telegramId": 0,
  "email": "string",
  "hwidDeviceLimit": 0,
  "activeInternalSquads": [
    "..."
  ],
  "externalSquadUuid": "550e8400-e29b-41d4-a716-446655440000"
}' \
```


---

## `GET` /api/users

**Get all users**

**Operation ID**: `UsersController_getAllUsers`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `size` | query | `number` | ❌ | Page size for pagination |
| `start` | query | `number` | ❌ | Offset for pagination |

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
curl -X GET "http://localhost/api/users?size=<size>&start=<start>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /api/users/{uuid}

**Delete user**

**Operation ID**: `UsersController_deleteUser`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User deleted successfully

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
curl -X DELETE "http://localhost/api/users/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/{uuid}

**Get user by UUID**

**Operation ID**: `UsersController_getUserByUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User fetched successfully

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
curl -X GET "http://localhost/api/users/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/tags

**Get all existing user tags**

**Operation ID**: `UsersController_getAllTags`

**Authentication**: `Authorization`

### Responses

#### Status `200`: Tags fetched successfully

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
curl -X GET "http://localhost/api/users/tags" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/{uuid}/accessible-nodes

**Get user accessible nodes**

**Operation ID**: `UsersController_getUserAccessibleNodes`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User accessible nodes fetched successfully

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
curl -X GET "http://localhost/api/users/<uuid>/accessible-nodes" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/{uuid}/subscription-request-history

**Get user subscription request history, recent 24 records**

**Operation ID**: `UsersController_getUserSubscriptionRequestHistory`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User subscription request history fetched successfully

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
curl -X GET "http://localhost/api/users/<uuid>/subscription-request-history" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-short-uuid/{shortUuid}

**Get user by Short UUID**

**Operation ID**: `UsersController_getUserByShortUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `shortUuid` | path | `string` | ✅ | Short UUID of the user |

### Responses

#### Status `200`: User fetched successfully

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
curl -X GET "http://localhost/api/users/by-short-uuid/<shortUuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-username/{username}

**Get user by username**

**Operation ID**: `UsersController_getUserByUsername`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `username` | path | `string` | ✅ | Username of the user |

### Responses

#### Status `200`: User fetched successfully

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
curl -X GET "http://localhost/api/users/by-username/<username>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-id/{id}

**Get user by ID**

**Operation ID**: `UsersController_getUserById`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `id` | path | `string` | ✅ | ID of the user |

### Responses

#### Status `200`: User fetched successfully

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
curl -X GET "http://localhost/api/users/by-id/<id>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-telegram-id/{telegramId}

**Get users by telegram ID**

**Operation ID**: `UsersController_getUserByTelegramId`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `telegramId` | path | `string` | ✅ | Telegram ID of the user |

### Responses

#### Status `200`: Users fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "id": 0,
      "shortUuid": "string",
      "username": "string",
      "status": "ACTIVE",
      "trafficLimitBytes": 0,
      "trafficLimitStrategy": "NO_RESET",
      "expireAt": "2024-01-01T00:00:00Z",
      "telegramId": 0,
      "email": "string",
      "description": "string",
      "tag": "string",
      "hwidDeviceLimit": 0,
      "externalSquadUuid": "550e8400-e29b-41d4-a716-446655440000",
      "trojanPassword": "string"
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
curl -X GET "http://localhost/api/users/by-telegram-id/<telegramId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-email/{email}

**Get users by email**

**Operation ID**: `UsersController_getUsersByEmail`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `email` | path | `string` | ✅ | Email of the user |

### Responses

#### Status `200`: Users fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "id": 0,
      "shortUuid": "string",
      "username": "string",
      "status": "ACTIVE",
      "trafficLimitBytes": 0,
      "trafficLimitStrategy": "NO_RESET",
      "expireAt": "2024-01-01T00:00:00Z",
      "telegramId": 0,
      "email": "string",
      "description": "string",
      "tag": "string",
      "hwidDeviceLimit": 0,
      "externalSquadUuid": "550e8400-e29b-41d4-a716-446655440000",
      "trojanPassword": "string"
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
curl -X GET "http://localhost/api/users/by-email/<email>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/users/by-tag/{tag}

**Get users by tag**

**Operation ID**: `UsersController_getUsersByTag`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `tag` | path | `string` | ✅ | Tag of the user |

### Responses

#### Status `200`: Users fetched successfully

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `response` | array of `object` | ✅ |  |

**Example Response**:

```json
{
  "response": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "id": 0,
      "shortUuid": "string",
      "username": "string",
      "status": "ACTIVE",
      "trafficLimitBytes": 0,
      "trafficLimitStrategy": "NO_RESET",
      "expireAt": "2024-01-01T00:00:00Z",
      "telegramId": 0,
      "email": "string",
      "description": "string",
      "tag": "string",
      "hwidDeviceLimit": 0,
      "externalSquadUuid": "550e8400-e29b-41d4-a716-446655440000",
      "trojanPassword": "string"
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
curl -X GET "http://localhost/api/users/by-tag/PROMO_1" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/users/{uuid}/actions/revoke

**Revoke user subscription**

**Operation ID**: `UsersController_revokeUserSubscription`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `revokeOnlyPasswords` | `boolean` | ❌ | Optional. If true, only passwords will be revoked, without changing the short UUID (Subscription URL). |
| `shortUuid` | `string` | ❌ | Optional. If not provided, a new short UUID will be generated by Remnawave. Please note that it is strongly recommended to allow Remnawave to generate the short UUID. |

**Example Body**:

```json
{
  "revokeOnlyPasswords": false,
  "shortUuid": "string"
}
```

### Responses

#### Status `200`: User subscription revoked successfully

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
curl -X POST "http://localhost/api/users/<uuid>/actions/revoke" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "revokeOnlyPasswords": false,
  "shortUuid": "string"
}' \
```


---

## `POST` /api/users/{uuid}/actions/disable

**Disable user**

**Operation ID**: `UsersController_disableUser`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User disabled successfully

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
curl -X POST "http://localhost/api/users/<uuid>/actions/disable" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/users/{uuid}/actions/enable

**Enable user**

**Operation ID**: `UsersController_enableUser`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User enabled successfully

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
curl -X POST "http://localhost/api/users/<uuid>/actions/enable" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/users/{uuid}/actions/reset-traffic

**Reset user traffic**

**Operation ID**: `UsersController_resetUserTraffic`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the user |

### Responses

#### Status `200`: User traffic reset successfully

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
curl -X POST "http://localhost/api/users/<uuid>/actions/reset-traffic" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
