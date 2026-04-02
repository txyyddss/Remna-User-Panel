# External Squads Controller

[← Back to Remnawave API Index](./README.md)

## Endpoints

- [`GET` /api/external-squads](#get-api-external-squads) — Get all external squads
- [`POST` /api/external-squads](#post-api-external-squads) — Create external squad
- [`PATCH` /api/external-squads](#patch-api-external-squads) — Update external squad
- [`GET` /api/external-squads/{uuid}](#get-api-external-squads-uuid) — Get external squad by uuid
- [`DELETE` /api/external-squads/{uuid}](#delete-api-external-squads-uuid) — Delete external squad
- [`POST` /api/external-squads/{uuid}/bulk-actions/add-users](#post-api-external-squads-uuid-bulk-actions-add-users) — Add all users to external squad
- [`DELETE` /api/external-squads/{uuid}/bulk-actions/remove-users](#delete-api-external-squads-uuid-bulk-actions-remove-users) — Delete users from external squad
- [`POST` /api/external-squads/actions/reorder](#post-api-external-squads-actions-reorder) — Reorder external squads

---

## `GET` /api/external-squads

**Get all external squads**

**Operation ID**: `ExternalSquadController_getExternalSquads`

**Authentication**: `Authorization`

### Responses

#### Status `200`: External squads retrieved successfully

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
curl -X GET "http://localhost/api/external-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/external-squads

**Create external squad**

**Operation ID**: `ExternalSquadController_createExternalSquad`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | `string` | ✅ |  |

**Example Body**:

```json
{
  "name": "string"
}
```

### Responses

#### Status `201`: External squad created successfully

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

#### Status `409`: External squad already exists

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
curl -X POST "http://localhost/api/external-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "name": "string"
}' \
```


---

## `PATCH` /api/external-squads

**Update external squad**

**Operation ID**: `ExternalSquadController_updateExternalSquad`

**Authentication**: `Authorization`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `uuid` | `string` (uuid) | ✅ |  |
| `name` | `string` | ❌ |  |
| `templates` | array of `object` | ❌ |  |
| `subscriptionSettings` | `object` | ❌ |  |
| `hostOverrides` | `object` | ❌ |  |
| `responseHeaders` | `object` | ❌ |  |
| `hwidSettings` | `object` | ❌ |  |
| `customRemarks` | `object` | ❌ |  |
| `subpageConfigUuid` | `string` (uuid) | ❌ |  |

**Example Body**:

```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "string",
  "templates": [
    {
      "templateUuid": "550e8400-e29b-41d4-a716-446655440000",
      "templateType": "XRAY_JSON"
    }
  ],
  "subscriptionSettings": null,
  "hostOverrides": null,
  "responseHeaders": null,
  "hwidSettings": null,
  "customRemarks": null,
  "subpageConfigUuid": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Responses

#### Status `200`: External squad updated successfully

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

#### Status `404`: External squad not found

#### Status `409`: External squad already exists

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
curl -X PATCH "http://localhost/api/external-squads" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "string",
  "templates": [
    "..."
  ],
  "subscriptionSettings": null,
  "hostOverrides": null,
  "responseHeaders": null,
  "hwidSettings": null,
  "customRemarks": null,
  "subpageConfigUuid": "550e8400-e29b-41d4-a716-446655440000"
}' \
```


---

## `GET` /api/external-squads/{uuid}

**Get external squad by uuid**

**Operation ID**: `ExternalSquadController_getExternalSquadByUuid`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: External squad retrieved successfully

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
curl -X GET "http://localhost/api/external-squads/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /api/external-squads/{uuid}

**Delete external squad**

**Operation ID**: `ExternalSquadController_deleteExternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: External squad deleted successfully

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

#### Status `404`: External squad not found

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
curl -X DELETE "http://localhost/api/external-squads/<uuid>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/external-squads/{uuid}/bulk-actions/add-users

**Add all users to external squad**

**Operation ID**: `ExternalSquadController_addUsersToExternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Task added to external job queue

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

#### Status `404`: External squad not found

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
curl -X POST "http://localhost/api/external-squads/<uuid>/bulk-actions/add-users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /api/external-squads/{uuid}/bulk-actions/remove-users

**Delete users from external squad**

**Operation ID**: `ExternalSquadController_removeUsersFromExternalSquad`

**Authentication**: `Authorization`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `uuid` | path | `string` | ✅ |  |

### Responses

#### Status `200`: Task added to external job queue

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

#### Status `404`: External squad not found

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
curl -X DELETE "http://localhost/api/external-squads/<uuid>/bulk-actions/remove-users" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /api/external-squads/actions/reorder

**Reorder external squads**

**Operation ID**: `ExternalSquadController_reorderExternalSquads`

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

#### Status `200`: External squads reordered successfully

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
curl -X POST "http://localhost/api/external-squads/actions/reorder" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "items": [
    "..."
  ]
}' \
```


---
