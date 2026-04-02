# Devices

[← Back to Jellyfin API Index](./README.md)

## Endpoints

- [`GET` /Devices](#get-devices) — Get Devices.
- [`DELETE` /Devices](#delete-devices) — Deletes a device.
- [`GET` /Devices/Info](#get-devices-info) — Get info for a device.
- [`GET` /Devices/Options](#get-devices-options) — Get options for a device.
- [`POST` /Devices/Options](#post-devices-options) — Update device options.

---

## `GET` /Devices

**Get Devices.**

**Operation ID**: `GetDevices`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | query | `string` (uuid) | ❌ | Gets or sets the user identifier. |

### Responses

#### Status `200`: Devices retrieved.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Items` | array of `DeviceInfoDto` | ❌ | Gets or sets the items. |
| `TotalRecordCount` | `integer` (int32) | ❌ | Gets or sets the total number of records available. |
| `StartIndex` | `integer` (int32) | ❌ | Gets or sets the index of the first record in Items. |

**Example Response**:

```json
{
  "Items": [
    "..."
  ],
  "TotalRecordCount": 0,
  "StartIndex": 0
}
```

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/Devices?userId=<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /Devices

**Deletes a device.**

**Operation ID**: `DeleteDevice`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `id` | query | `string` | ✅ | Device Id. |

### Responses

#### Status `204`: Device deleted.

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `404`: Device not found.

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
curl -X DELETE "http://localhost/Devices?id=<id>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /Devices/Info

**Get info for a device.**

**Operation ID**: `GetDeviceInfo`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `id` | query | `string` | ✅ | Device Id. |

### Responses

#### Status `200`: Device info retrieved.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `CustomName` | `string` | ❌ | Gets or sets the custom name. |
| `AccessToken` | `string` | ❌ | Gets or sets the access token. |
| `Id` | `string` | ❌ | Gets or sets the identifier. |
| `LastUserName` | `string` | ❌ | Gets or sets the last name of the user. |
| `AppName` | `string` | ❌ | Gets or sets the name of the application. |
| `AppVersion` | `string` | ❌ | Gets or sets the application version. |
| `LastUserId` | `string` (uuid) | ❌ | Gets or sets the last user identifier. |
| `DateLastActivity` | `string` (date-time) | ❌ | Gets or sets the date last modified. |
| `Capabilities` | `ClientCapabilitiesDto` | ❌ | Client capabilities dto. |
| `IconUrl` | `string` | ❌ | Gets or sets the icon URL. |

**Example Response**:

```json
{
  "Name": "string",
  "CustomName": "string",
  "AccessToken": "string",
  "Id": "string",
  "LastUserName": "string",
  "AppName": "string",
  "AppVersion": "string",
  "LastUserId": "550e8400-e29b-41d4-a716-446655440000",
  "DateLastActivity": "2024-01-01T00:00:00Z",
  "Capabilities": null,
  "IconUrl": "string"
}
```

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `404`: Device not found.

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
curl -X GET "http://localhost/Devices/Info?id=<id>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `GET` /Devices/Options

**Get options for a device.**

**Operation ID**: `GetDeviceOptions`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `id` | query | `string` | ✅ | Device Id. |

### Responses

#### Status `200`: Device options retrieved.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Id` | `integer` (int32) | ❌ | Gets or sets the id. |
| `DeviceId` | `string` | ❌ | Gets or sets the device id. |
| `CustomName` | `string` | ❌ | Gets or sets the custom name. |

**Example Response**:

```json
{
  "Id": 0,
  "DeviceId": "string",
  "CustomName": "string"
}
```

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `404`: Device not found.

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
curl -X GET "http://localhost/Devices/Options?id=<id>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /Devices/Options

**Update device options.**

**Operation ID**: `UpdateDeviceOptions`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `id` | query | `string` | ✅ | Device Id. |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Id` | `integer` (int32) | ❌ | Gets or sets the id. |
| `DeviceId` | `string` | ❌ | Gets or sets the device id. |
| `CustomName` | `string` | ❌ | Gets or sets the custom name. |

**Example Body**:

```json
{
  "Id": 0,
  "DeviceId": "string",
  "CustomName": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Id` | `integer` (int32) | ❌ | Gets or sets the id. |
| `DeviceId` | `string` | ❌ | Gets or sets the device id. |
| `CustomName` | `string` | ❌ | Gets or sets the custom name. |

**Example Body**:

```json
{
  "Id": 0,
  "DeviceId": "string",
  "CustomName": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Id` | `integer` (int32) | ❌ | Gets or sets the id. |
| `DeviceId` | `string` | ❌ | Gets or sets the device id. |
| `CustomName` | `string` | ❌ | Gets or sets the custom name. |

**Example Body**:

```json
{
  "Id": 0,
  "DeviceId": "string",
  "CustomName": "string"
}
```

### Responses

#### Status `204`: Device options updated.

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/Devices/Options?id=<id>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Id": 0,
  "DeviceId": "string",
  "CustomName": "string"
}' \
```


---
