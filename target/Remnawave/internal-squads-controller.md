# Internal Squads Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/internal-squads](#get-api-internal-squads) — Get all internal squads
- [`POST` /api/internal-squads](#post-api-internal-squads) — Create internal squad
- [`PATCH` /api/internal-squads](#patch-api-internal-squads) — Update internal squad
- [`GET` /api/internal-squads/{uuid}](#get-api-internal-squads-uuid) — Get internal squad by uuid
- [`DELETE` /api/internal-squads/{uuid}](#delete-api-internal-squads-uuid) — Delete internal squad
- [`GET` /api/internal-squads/{uuid}/accessible-nodes](#get-api-internal-squads-uuid-accessible-nodes) — Get internal squad accessible nodes
- [`POST` /api/internal-squads/{uuid}/bulk-actions/add-users](#post-api-internal-squads-uuid-bulk-actions-add-users) — Add all users to internal squad
- [`DELETE` /api/internal-squads/{uuid}/bulk-actions/remove-users](#delete-api-internal-squads-uuid-bulk-actions-remove-users) — Delete users from internal squad
- [`POST` /api/internal-squads/actions/reorder](#post-api-internal-squads-actions-reorder) — Reorder internal squads

---

## `GET` /api/internal-squads

**Get all internal squads**

**Operation ID**: `InternalSquadController_getInternalSquads`

**Authentication**: `Authorization`

### Responses

#### Status `200`: Internal squads retrieved successfully

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
curl -X GET "http://localhost/api/internal-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/internal-squads

**Create internal squad**

**Operation ID**: `InternalSquadController_createInternalSquad`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | `string` | ✅ |  |
| `inbounds` | array of `string` (uuid) | ✅ |  |

**Example Body**:

```json
{
  "name": "string",
  "inbounds": [
    "string"
  ]
}
```

### Responses

#### Status `201`: Internal squad created successfully

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

#### Status `409`: Internal squad already exists

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
curl -X POST "http://localhost/api/internal-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "name": "string",
  "inbounds": [
    "..."
  ]
}' \
```


---

## `PATCH` /api/internal-squads

**Update internal squad**

**Operation ID**: `InternalSquadController_updateInternalSquad`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `uuid` | `string` (uuid) | ✅ |  |
| `name` | `string` | ❌ |  |
| `inbounds` | array of `string` (uuid) | ❌ |  |

**Example Body**:

```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "string",
  "inbounds": [
    "string"
  ]
}
```

### Responses

#### Status `200`: Internal squad updated successfully

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

#### Status `404`: Internal squad not found

#### Status `409`: Internal squad already exists

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
curl -X PATCH "http://localhost/api/internal-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "string",
  "inbounds": [
    "..."
  ]
}' \
```


---

## `GET` /api/internal-squads/{uuid}

**Get internal squad by uuid**

**Operation ID**: `InternalSquadController_getInternalSquadByUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Internal squad retrieved successfully

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
curl -X GET "http://localhost/api/internal-squads/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /api/internal-squads/{uuid}

**Delete internal squad**

**Operation ID**: `InternalSquadController_deleteInternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Internal squad deleted successfully

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

#### Status `404`: Internal squad not found

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
curl -X DELETE "http://localhost/api/internal-squads/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /api/internal-squads/{uuid}/accessible-nodes

**Get internal squad accessible nodes**

**Operation ID**: `InternalSquadController_getInternalSquadAccessibleNodes`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ | UUID of the internal squad |

### Responses

#### Status `200`: Internal squad accessible nodes fetched successfully

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

#### Status `404`: Internal squad not found

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
curl -X GET "http://localhost/api/internal-squads/<uuid>/accessible-nodes" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/internal-squads/{uuid}/bulk-actions/add-users

**Add all users to internal squad**

**Operation ID**: `InternalSquadController_addUsersToInternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Task added to internal job queue

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

#### Status `404`: Internal squad not found

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
curl -X POST "http://localhost/api/internal-squads/<uuid>/bulk-actions/add-users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /api/internal-squads/{uuid}/bulk-actions/remove-users

**Delete users from internal squad**

**Operation ID**: `InternalSquadController_removeUsersFromInternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Task added to internal job queue

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

#### Status `404`: Internal squad not found

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
curl -X DELETE "http://localhost/api/internal-squads/<uuid>/bulk-actions/remove-users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/internal-squads/actions/reorder

**Reorder internal squads**

**Operation ID**: `InternalSquadController_reorderInternalSquads`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `items` | array of `object` | ✅ |  |

**Example Body**:

```json
{
  "items": [
    {
      "viewPosition": 0,
      "uuid": "550e8400-e29b-41d4-a716-446655440000"
    }
  ]
}
```

### Responses

#### Status `200`: Internal squads reordered successfully

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
curl -X POST "http://localhost/api/internal-squads/actions/reorder" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "items": [
    "..."
  ]
}' \
```


---
