# QuickConnect

[← Back to Jellyfin API Index](./README.md)

## Endpoints

- [`POST` /QuickConnect/Authorize](#post-quickconnect-authorize) — Authorizes a pending quick connect request.
- [`GET` /QuickConnect/Connect](#get-quickconnect-connect) — Attempts to retrieve authentication information.
- [`GET` /QuickConnect/Enabled](#get-quickconnect-enabled) — Gets the current quick connect state.
- [`POST` /QuickConnect/Initiate](#post-quickconnect-initiate) — Initiate a new quick connect request.

---

## `POST` /QuickConnect/Authorize

**Authorizes a pending quick connect request.**

**Operation ID**: `AuthorizeQuickConnect`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `code` | query | `string` | ✅ | Quick connect code to authorize. |
| `userId` | query | `string` (uuid) | ❌ | The user the authorize. Access to the requested user is required. |

### Responses

#### Status `200`: Quick connect result authorized successfully.



#### Status `401`: Unauthorized

#### Status `403`: Unknown user id.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `type` | `string` | ❌ |  |
| `title` | `string` | ❌ |  |
| `status` | `integer` (int32) | ❌ |  |
| `detail` | `string` | ❌ |  |
| `instance` | `string` | ❌ |  |

**Example Response**:

```json
{
  "type": "string",
  "title": "string",
  "status": 0,
  "detail": "string",
  "instance": "string"
}
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/QuickConnect/Authorize?code=<code>&userId=<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /QuickConnect/Connect

**Attempts to retrieve authentication information.**

**Operation ID**: `GetQuickConnectState`

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `secret` | query | `string` | ✅ | Secret previously returned from the Initiate endpoint. |

### Responses

#### Status `200`: Quick connect result returned.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Authenticated` | `boolean` | ❌ | Gets or sets a value indicating whether this request is authorized. |
| `Secret` | `string` | ❌ | Gets the secret value used to uniquely identify this request. Can be used to retrieve authentication information. |
| `Code` | `string` | ❌ | Gets the user facing code used so the user can quickly differentiate this request from others. |
| `DeviceId` | `string` | ❌ | Gets the requesting device id. |
| `DeviceName` | `string` | ❌ | Gets the requesting device name. |
| `AppName` | `string` | ❌ | Gets the requesting app name. |
| `AppVersion` | `string` | ❌ | Gets the requesting app version. |
| `DateAdded` | `string` (date-time) | ❌ | Gets or sets the DateTime that this request was created. |

**Example Response**:

```json
{
  "Authenticated": false,
  "Secret": "string",
  "Code": "string",
  "DeviceId": "string",
  "DeviceName": "string",
  "AppName": "string",
  "AppVersion": "string",
  "DateAdded": "2024-01-01T00:00:00Z"
}
```

#### Status `404`: Unknown quick connect secret.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `type` | `string` | ❌ |  |
| `title` | `string` | ❌ |  |
| `status` | `integer` (int32) | ❌ |  |
| `detail` | `string` | ❌ |  |
| `instance` | `string` | ❌ |  |

**Example Response**:

```json
{
  "type": "string",
  "title": "string",
  "status": 0,
  "detail": "string",
  "instance": "string"
}
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/QuickConnect/Connect?secret=<secret>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /QuickConnect/Enabled

**Gets the current quick connect state.**

**Operation ID**: `GetQuickConnectEnabled`

### Responses

#### Status `200`: Quick connect state returned.



#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/QuickConnect/Enabled" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /QuickConnect/Initiate

**Initiate a new quick connect request.**

**Operation ID**: `InitiateQuickConnect`

### Responses

#### Status `200`: Quick connect request successfully created.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Authenticated` | `boolean` | ❌ | Gets or sets a value indicating whether this request is authorized. |
| `Secret` | `string` | ❌ | Gets the secret value used to uniquely identify this request. Can be used to retrieve authentication information. |
| `Code` | `string` | ❌ | Gets the user facing code used so the user can quickly differentiate this request from others. |
| `DeviceId` | `string` | ❌ | Gets the requesting device id. |
| `DeviceName` | `string` | ❌ | Gets the requesting device name. |
| `AppName` | `string` | ❌ | Gets the requesting app name. |
| `AppVersion` | `string` | ❌ | Gets the requesting app version. |
| `DateAdded` | `string` (date-time) | ❌ | Gets or sets the DateTime that this request was created. |

**Example Response**:

```json
{
  "Authenticated": false,
  "Secret": "string",
  "Code": "string",
  "DeviceId": "string",
  "DeviceName": "string",
  "AppName": "string",
  "AppVersion": "string",
  "DateAdded": "2024-01-01T00:00:00Z"
}
```

#### Status `401`: Quick connect is not active on this server.

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/QuickConnect/Initiate" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
