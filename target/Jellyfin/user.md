# User

[← Back to Jellyfin API Index](./README.md)

## Endpoints

- [`GET` /Users](#get-users) — Gets a list of users.
- [`POST` /Users](#post-users) — Updates a user.
- [`GET` /Users/{userId}](#get-users-userid) — Gets a user by Id.
- [`DELETE` /Users/{userId}](#delete-users-userid) — Deletes a user.
- [`POST` /Users/{userId}/Policy](#post-users-userid-policy) — Updates a user policy.
- [`POST` /Users/AuthenticateByName](#post-users-authenticatebyname) — Authenticates a user by name.
- [`POST` /Users/AuthenticateWithQuickConnect](#post-users-authenticatewithquickconnect) — Authenticates a user with quick connect.
- [`POST` /Users/Configuration](#post-users-configuration) — Updates a user configuration.
- [`POST` /Users/ForgotPassword](#post-users-forgotpassword) — Initiates the forgot password process for a local user.
- [`POST` /Users/ForgotPassword/Pin](#post-users-forgotpassword-pin) — Redeems a forgot password pin.
- [`GET` /Users/Me](#get-users-me) — Gets the user based on auth token.
- [`POST` /Users/New](#post-users-new) — Creates a user.
- [`POST` /Users/Password](#post-users-password) — Updates a user's password.
- [`GET` /Users/Public](#get-users-public) — Gets a list of publicly visible users for display on a login screen.

---

## `GET` /Users

**Gets a list of users.**

**Operation ID**: `GetUsers`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `isHidden` | query | `boolean` | ❌ | Optional filter by IsHidden=true or false. |
| `isDisabled` | query | `boolean` | ❌ | Optional filter by IsDisabled=true or false. |

### Responses

#### Status `200`: Users returned.



**Example Response**:

```json
[
  {
    "Name": "string",
    "ServerId": "string",
    "ServerName": "string",
    "Id": "550e8400-e29b-41d4-a716-446655440000",
    "PrimaryImageTag": "string",
    "HasPassword": false,
    "HasConfiguredPassword": false,
    "HasConfiguredEasyPassword": false,
    "EnableAutoLogin": false,
    "LastLoginDate": "2024-01-01T00:00:00Z",
    "LastActivityDate": "2024-01-01T00:00:00Z",
    "Configuration": null,
    "Policy": null,
    "PrimaryImageAspectRatio": 0
  }
]
```

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/Users?isHidden=<isHidden>&isDisabled=<isDisabled>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /Users

**Updates a user.**

**Operation ID**: `UpdateUser`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | query | `string` (uuid) | ❌ | The user id. |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Body**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Body**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Body**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}
```

### Responses

#### Status `204`: User updated.

#### Status `400`: User information was not supplied.

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

#### Status `401`: Unauthorized

#### Status `403`: User update forbidden.

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
curl -X POST "http://localhost/Users?userId=<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}' \
```


---

## `GET` /Users/{userId}

**Gets a user by Id.**

**Operation ID**: `GetUserById`

**Authentication**: `CustomAuthentication` (scopes: IgnoreParentalControl, DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | path | `string` (uuid) | ✅ | The user id. |

### Responses

#### Status `200`: User returned.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Response**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}
```

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `404`: User not found.

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
curl -X GET "http://localhost/Users/<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `DELETE` /Users/{userId}

**Deletes a user.**

**Operation ID**: `DeleteUser`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | path | `string` (uuid) | ✅ | The user id. |

### Responses

#### Status `204`: User deleted.

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `404`: User not found.

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
curl -X DELETE "http://localhost/Users/<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /Users/{userId}/Policy

**Updates a user policy.**

**Operation ID**: `UpdateUserPolicy`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | path | `string` (uuid) | ✅ | The user id. |

### Request Body

**Note**: MaxParentalSubRating have value 0~21,1000 for XXX, and 1001 for Banned

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IsAdministrator` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is administrator. |
| `IsHidden` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is hidden. |
| `EnableCollectionManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage collections. |
| `EnableSubtitleManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage subtitles. |
| `EnableLyricManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this user can manage lyrics. |
| `IsDisabled` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is disabled. |
| `MaxParentalRating` | `integer` (int32) | ❌ | Gets or sets the max parental rating. |
| `MaxParentalSubRating` | `integer` (int32) | ❌ |  |
| `BlockedTags` | array of `string` | ❌ |  |
| `AllowedTags` | array of `string` | ❌ |  |
| `EnableUserPreferenceAccess` | `boolean` | ❌ |  |
| `AccessSchedules` | array of `AccessSchedule` | ❌ |  |
| `BlockUnratedItems` | array of `UnratedItem` | ❌ |  |
| `EnableRemoteControlOfOtherUsers` | `boolean` | ❌ |  |
| `EnableSharedDeviceControl` | `boolean` | ❌ |  |
| `EnableRemoteAccess` | `boolean` | ❌ |  |
| `EnableLiveTvManagement` | `boolean` | ❌ |  |
| `EnableLiveTvAccess` | `boolean` | ❌ |  |
| `EnableMediaPlayback` | `boolean` | ❌ |  |
| `EnableAudioPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnableVideoPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnablePlaybackRemuxing` | `boolean` | ❌ |  |
| `ForceRemoteSourceTranscoding` | `boolean` | ❌ |  |
| `EnableContentDeletion` | `boolean` | ❌ |  |
| `EnableContentDeletionFromFolders` | array of `string` | ❌ |  |
| `EnableContentDownloading` | `boolean` | ❌ |  |
| `EnableSyncTranscoding` | `boolean` | ❌ | Gets or sets a value indicating whether [enable synchronize]. |
| `EnableMediaConversion` | `boolean` | ❌ |  |
| `EnabledDevices` | array of `string` | ❌ |  |
| `EnableAllDevices` | `boolean` | ❌ |  |
| `EnabledChannels` | array of `string` (uuid) | ❌ |  |
| `EnableAllChannels` | `boolean` | ❌ |  |
| `EnabledFolders` | array of `string` (uuid) | ❌ |  |
| `EnableAllFolders` | `boolean` | ❌ |  |
| `InvalidLoginAttemptCount` | `integer` (int32) | ❌ |  |
| `LoginAttemptsBeforeLockout` | `integer` (int32) | ❌ |  |
| `MaxActiveSessions` | `integer` (int32) | ❌ |  |
| `EnablePublicSharing` | `boolean` | ❌ |  |
| `BlockedMediaFolders` | array of `string` (uuid) | ❌ |  |
| `BlockedChannels` | array of `string` (uuid) | ❌ |  |
| `RemoteClientBitrateLimit` | `integer` (int32) | ❌ |  |
| `AuthenticationProviderId` | `string` | ✅ |  |
| `PasswordResetProviderId` | `string` | ✅ |  |
| `SyncPlayAccess` | `SyncPlayUserAccessType` | ❌ | Enum SyncPlayUserAccessType. Enum: [`CreateAndJoinGroups`, `JoinGroups`, `None`] |

**Example Body**:

```json
{
  "IsAdministrator": false,
  "IsHidden": false,
  "EnableCollectionManagement": false,
  "EnableSubtitleManagement": false,
  "EnableLyricManagement": false,
  "IsDisabled": false,
  "MaxParentalRating": 0,
  "MaxParentalSubRating": 0,
  "BlockedTags": [
    "string"
  ],
  "AllowedTags": [
    "string"
  ],
  "EnableUserPreferenceAccess": false,
  "AccessSchedules": [
    {
      "Id": 0,
      "UserId": "550e8400-e29b-41d4-a716-446655440000",
      "DayOfWeek": null,
      "StartHour": 0,
      "EndHour": 0
    }
  ],
  "BlockUnratedItems": [
    "Movie"
  ],
  "EnableRemoteControlOfOtherUsers": false,
  "EnableSharedDeviceControl": false
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IsAdministrator` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is administrator. |
| `IsHidden` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is hidden. |
| `EnableCollectionManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage collections. |
| `EnableSubtitleManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage subtitles. |
| `EnableLyricManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this user can manage lyrics. |
| `IsDisabled` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is disabled. |
| `MaxParentalRating` | `integer` (int32) | ❌ | Gets or sets the max parental rating. |
| `MaxParentalSubRating` | `integer` (int32) | ❌ |  |
| `BlockedTags` | array of `string` | ❌ |  |
| `AllowedTags` | array of `string` | ❌ |  |
| `EnableUserPreferenceAccess` | `boolean` | ❌ |  |
| `AccessSchedules` | array of `AccessSchedule` | ❌ |  |
| `BlockUnratedItems` | array of `UnratedItem` | ❌ |  |
| `EnableRemoteControlOfOtherUsers` | `boolean` | ❌ |  |
| `EnableSharedDeviceControl` | `boolean` | ❌ |  |
| `EnableRemoteAccess` | `boolean` | ❌ |  |
| `EnableLiveTvManagement` | `boolean` | ❌ |  |
| `EnableLiveTvAccess` | `boolean` | ❌ |  |
| `EnableMediaPlayback` | `boolean` | ❌ |  |
| `EnableAudioPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnableVideoPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnablePlaybackRemuxing` | `boolean` | ❌ |  |
| `ForceRemoteSourceTranscoding` | `boolean` | ❌ |  |
| `EnableContentDeletion` | `boolean` | ❌ |  |
| `EnableContentDeletionFromFolders` | array of `string` | ❌ |  |
| `EnableContentDownloading` | `boolean` | ❌ |  |
| `EnableSyncTranscoding` | `boolean` | ❌ | Gets or sets a value indicating whether [enable synchronize]. |
| `EnableMediaConversion` | `boolean` | ❌ |  |
| `EnabledDevices` | array of `string` | ❌ |  |
| `EnableAllDevices` | `boolean` | ❌ |  |
| `EnabledChannels` | array of `string` (uuid) | ❌ |  |
| `EnableAllChannels` | `boolean` | ❌ |  |
| `EnabledFolders` | array of `string` (uuid) | ❌ |  |
| `EnableAllFolders` | `boolean` | ❌ |  |
| `InvalidLoginAttemptCount` | `integer` (int32) | ❌ |  |
| `LoginAttemptsBeforeLockout` | `integer` (int32) | ❌ |  |
| `MaxActiveSessions` | `integer` (int32) | ❌ |  |
| `EnablePublicSharing` | `boolean` | ❌ |  |
| `BlockedMediaFolders` | array of `string` (uuid) | ❌ |  |
| `BlockedChannels` | array of `string` (uuid) | ❌ |  |
| `RemoteClientBitrateLimit` | `integer` (int32) | ❌ |  |
| `AuthenticationProviderId` | `string` | ✅ |  |
| `PasswordResetProviderId` | `string` | ✅ |  |
| `SyncPlayAccess` | `SyncPlayUserAccessType` | ❌ | Enum SyncPlayUserAccessType. Enum: [`CreateAndJoinGroups`, `JoinGroups`, `None`] |

**Example Body**:

```json
{
  "IsAdministrator": false,
  "IsHidden": false,
  "EnableCollectionManagement": false,
  "EnableSubtitleManagement": false,
  "EnableLyricManagement": false,
  "IsDisabled": false,
  "MaxParentalRating": 0,
  "MaxParentalSubRating": 0,
  "BlockedTags": [
    "string"
  ],
  "AllowedTags": [
    "string"
  ],
  "EnableUserPreferenceAccess": false,
  "AccessSchedules": [
    {
      "Id": 0,
      "UserId": "550e8400-e29b-41d4-a716-446655440000",
      "DayOfWeek": null,
      "StartHour": 0,
      "EndHour": 0
    }
  ],
  "BlockUnratedItems": [
    "Movie"
  ],
  "EnableRemoteControlOfOtherUsers": false,
  "EnableSharedDeviceControl": false
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IsAdministrator` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is administrator. |
| `IsHidden` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is hidden. |
| `EnableCollectionManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage collections. |
| `EnableSubtitleManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this instance can manage subtitles. |
| `EnableLyricManagement` | `boolean` | ❌ | Gets or sets a value indicating whether this user can manage lyrics. |
| `IsDisabled` | `boolean` | ❌ | Gets or sets a value indicating whether this instance is disabled. |
| `MaxParentalRating` | `integer` (int32) | ❌ | Gets or sets the max parental rating. |
| `MaxParentalSubRating` | `integer` (int32) | ❌ |  |
| `BlockedTags` | array of `string` | ❌ |  |
| `AllowedTags` | array of `string` | ❌ |  |
| `EnableUserPreferenceAccess` | `boolean` | ❌ |  |
| `AccessSchedules` | array of `AccessSchedule` | ❌ |  |
| `BlockUnratedItems` | array of `UnratedItem` | ❌ |  |
| `EnableRemoteControlOfOtherUsers` | `boolean` | ❌ |  |
| `EnableSharedDeviceControl` | `boolean` | ❌ |  |
| `EnableRemoteAccess` | `boolean` | ❌ |  |
| `EnableLiveTvManagement` | `boolean` | ❌ |  |
| `EnableLiveTvAccess` | `boolean` | ❌ |  |
| `EnableMediaPlayback` | `boolean` | ❌ |  |
| `EnableAudioPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnableVideoPlaybackTranscoding` | `boolean` | ❌ |  |
| `EnablePlaybackRemuxing` | `boolean` | ❌ |  |
| `ForceRemoteSourceTranscoding` | `boolean` | ❌ |  |
| `EnableContentDeletion` | `boolean` | ❌ |  |
| `EnableContentDeletionFromFolders` | array of `string` | ❌ |  |
| `EnableContentDownloading` | `boolean` | ❌ |  |
| `EnableSyncTranscoding` | `boolean` | ❌ | Gets or sets a value indicating whether [enable synchronize]. |
| `EnableMediaConversion` | `boolean` | ❌ |  |
| `EnabledDevices` | array of `string` | ❌ |  |
| `EnableAllDevices` | `boolean` | ❌ |  |
| `EnabledChannels` | array of `string` (uuid) | ❌ |  |
| `EnableAllChannels` | `boolean` | ❌ |  |
| `EnabledFolders` | array of `string` (uuid) | ❌ |  |
| `EnableAllFolders` | `boolean` | ❌ |  |
| `InvalidLoginAttemptCount` | `integer` (int32) | ❌ |  |
| `LoginAttemptsBeforeLockout` | `integer` (int32) | ❌ |  |
| `MaxActiveSessions` | `integer` (int32) | ❌ |  |
| `EnablePublicSharing` | `boolean` | ❌ |  |
| `BlockedMediaFolders` | array of `string` (uuid) | ❌ |  |
| `BlockedChannels` | array of `string` (uuid) | ❌ |  |
| `RemoteClientBitrateLimit` | `integer` (int32) | ❌ |  |
| `AuthenticationProviderId` | `string` | ✅ |  |
| `PasswordResetProviderId` | `string` | ✅ |  |
| `SyncPlayAccess` | `SyncPlayUserAccessType` | ❌ | Enum SyncPlayUserAccessType. Enum: [`CreateAndJoinGroups`, `JoinGroups`, `None`] |

**Example Body**:

```json
{
  "IsAdministrator": false,
  "IsHidden": false,
  "EnableCollectionManagement": false,
  "EnableSubtitleManagement": false,
  "EnableLyricManagement": false,
  "IsDisabled": false,
  "MaxParentalRating": 0,
  "MaxParentalSubRating": 0,
  "BlockedTags": [
    "string"
  ],
  "AllowedTags": [
    "string"
  ],
  "EnableUserPreferenceAccess": false,
  "AccessSchedules": [
    {
      "Id": 0,
      "UserId": "550e8400-e29b-41d4-a716-446655440000",
      "DayOfWeek": null,
      "StartHour": 0,
      "EndHour": 0
    }
  ],
  "BlockUnratedItems": [
    "Movie"
  ],
  "EnableRemoteControlOfOtherUsers": false,
  "EnableSharedDeviceControl": false
}
```

### Responses

#### Status `204`: User policy updated.

#### Status `400`: User policy was not supplied.

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

#### Status `401`: Unauthorized

#### Status `403`: User policy update forbidden.

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
curl -X POST "http://localhost/Users/<userId>/Policy" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "IsAdministrator": false,
  "IsHidden": false,
  "EnableCollectionManagement": false,
  "EnableSubtitleManagement": false,
  "EnableLyricManagement": false,
  "IsDisabled": false,
  "MaxParentalRating": 0,
  "MaxParentalSubRating": 0,
  "BlockedTags": [
    "string"
  ],
  "AllowedTags": [
    "string"
  ],
  "EnableUserPreferenceAccess": false,
  "AccessSchedules": [
    "..."
  ],
  "BlockUnratedItems": [
    "..."
  ],
  "EnableRemoteControlOfOtherUsers": false,
  "EnableSharedDeviceControl": false
}' \
```


---

## `POST` /Users/AuthenticateByName

**Authenticates a user by name.**

**Operation ID**: `AuthenticateUserByName`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Username` | `string` | ❌ | Gets or sets the username. |
| `Pw` | `string` | ❌ | Gets or sets the plain text password. |

**Example Body**:

```json
{
  "Username": "string",
  "Pw": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Username` | `string` | ❌ | Gets or sets the username. |
| `Pw` | `string` | ❌ | Gets or sets the plain text password. |

**Example Body**:

```json
{
  "Username": "string",
  "Pw": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Username` | `string` | ❌ | Gets or sets the username. |
| `Pw` | `string` | ❌ | Gets or sets the plain text password. |

**Example Body**:

```json
{
  "Username": "string",
  "Pw": "string"
}
```

### Responses

#### Status `200`: User authenticated.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `User` | `UserDto` | ❌ | Class UserDto. |
| `SessionInfo` | `SessionInfoDto` | ❌ | Session info DTO. |
| `AccessToken` | `string` | ❌ | Gets or sets the access token. |
| `ServerId` | `string` | ❌ | Gets or sets the server id. |

**Example Response**:

```json
{
  "User": null,
  "SessionInfo": null,
  "AccessToken": "string",
  "ServerId": "string"
}
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/Users/AuthenticateByName" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Username": "string",
  "Pw": "string"
}' \
```


---

## `POST` /Users/AuthenticateWithQuickConnect

**Authenticates a user with quick connect.**

**Operation ID**: `AuthenticateWithQuickConnect`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Secret` | `string` | ✅ | Gets or sets the quick connect secret. |

**Example Body**:

```json
{
  "Secret": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Secret` | `string` | ✅ | Gets or sets the quick connect secret. |

**Example Body**:

```json
{
  "Secret": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Secret` | `string` | ✅ | Gets or sets the quick connect secret. |

**Example Body**:

```json
{
  "Secret": "string"
}
```

### Responses

#### Status `200`: User authenticated.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `User` | `UserDto` | ❌ | Class UserDto. |
| `SessionInfo` | `SessionInfoDto` | ❌ | Session info DTO. |
| `AccessToken` | `string` | ❌ | Gets or sets the access token. |
| `ServerId` | `string` | ❌ | Gets or sets the server id. |

**Example Response**:

```json
{
  "User": null,
  "SessionInfo": null,
  "AccessToken": "string",
  "ServerId": "string"
}
```

#### Status `400`: Missing token.

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/Users/AuthenticateWithQuickConnect" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Secret": "string"
}' \
```


---

## `POST` /Users/Configuration

**Updates a user configuration.**

**Operation ID**: `UpdateUserConfiguration`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | query | `string` (uuid) | ❌ | The user id. |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `AudioLanguagePreference` | `string` | ❌ | Gets or sets the audio language preference. |
| `PlayDefaultAudioTrack` | `boolean` | ❌ | Gets or sets a value indicating whether [play default audio track]. |
| `SubtitleLanguagePreference` | `string` | ❌ | Gets or sets the subtitle language preference. |
| `DisplayMissingEpisodes` | `boolean` | ❌ |  |
| `GroupedFolders` | array of `string` (uuid) | ❌ |  |
| `SubtitleMode` | `SubtitlePlaybackMode` | ❌ | An enum representing a subtitle playback mode. Enum: [`Default`, `Always`, `OnlyForced`, `None`, `Smart`] |
| `DisplayCollectionsView` | `boolean` | ❌ |  |
| `EnableLocalPassword` | `boolean` | ❌ |  |
| `OrderedViews` | array of `string` (uuid) | ❌ |  |
| `LatestItemsExcludes` | array of `string` (uuid) | ❌ |  |
| `MyMediaExcludes` | array of `string` (uuid) | ❌ |  |
| `HidePlayedInLatest` | `boolean` | ❌ |  |
| `RememberAudioSelections` | `boolean` | ❌ |  |
| `RememberSubtitleSelections` | `boolean` | ❌ |  |
| `EnableNextEpisodeAutoPlay` | `boolean` | ❌ |  |
| `CastReceiverId` | `string` | ❌ | Gets or sets the id of the selected cast receiver. |

**Example Body**:

```json
{
  "AudioLanguagePreference": "string",
  "PlayDefaultAudioTrack": false,
  "SubtitleLanguagePreference": "string",
  "DisplayMissingEpisodes": false,
  "GroupedFolders": [
    "string"
  ],
  "SubtitleMode": null,
  "DisplayCollectionsView": false,
  "EnableLocalPassword": false,
  "OrderedViews": [
    "string"
  ],
  "LatestItemsExcludes": [
    "string"
  ],
  "MyMediaExcludes": [
    "string"
  ],
  "HidePlayedInLatest": false,
  "RememberAudioSelections": false,
  "RememberSubtitleSelections": false,
  "EnableNextEpisodeAutoPlay": false
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `AudioLanguagePreference` | `string` | ❌ | Gets or sets the audio language preference. |
| `PlayDefaultAudioTrack` | `boolean` | ❌ | Gets or sets a value indicating whether [play default audio track]. |
| `SubtitleLanguagePreference` | `string` | ❌ | Gets or sets the subtitle language preference. |
| `DisplayMissingEpisodes` | `boolean` | ❌ |  |
| `GroupedFolders` | array of `string` (uuid) | ❌ |  |
| `SubtitleMode` | `SubtitlePlaybackMode` | ❌ | An enum representing a subtitle playback mode. Enum: [`Default`, `Always`, `OnlyForced`, `None`, `Smart`] |
| `DisplayCollectionsView` | `boolean` | ❌ |  |
| `EnableLocalPassword` | `boolean` | ❌ |  |
| `OrderedViews` | array of `string` (uuid) | ❌ |  |
| `LatestItemsExcludes` | array of `string` (uuid) | ❌ |  |
| `MyMediaExcludes` | array of `string` (uuid) | ❌ |  |
| `HidePlayedInLatest` | `boolean` | ❌ |  |
| `RememberAudioSelections` | `boolean` | ❌ |  |
| `RememberSubtitleSelections` | `boolean` | ❌ |  |
| `EnableNextEpisodeAutoPlay` | `boolean` | ❌ |  |
| `CastReceiverId` | `string` | ❌ | Gets or sets the id of the selected cast receiver. |

**Example Body**:

```json
{
  "AudioLanguagePreference": "string",
  "PlayDefaultAudioTrack": false,
  "SubtitleLanguagePreference": "string",
  "DisplayMissingEpisodes": false,
  "GroupedFolders": [
    "string"
  ],
  "SubtitleMode": null,
  "DisplayCollectionsView": false,
  "EnableLocalPassword": false,
  "OrderedViews": [
    "string"
  ],
  "LatestItemsExcludes": [
    "string"
  ],
  "MyMediaExcludes": [
    "string"
  ],
  "HidePlayedInLatest": false,
  "RememberAudioSelections": false,
  "RememberSubtitleSelections": false,
  "EnableNextEpisodeAutoPlay": false
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `AudioLanguagePreference` | `string` | ❌ | Gets or sets the audio language preference. |
| `PlayDefaultAudioTrack` | `boolean` | ❌ | Gets or sets a value indicating whether [play default audio track]. |
| `SubtitleLanguagePreference` | `string` | ❌ | Gets or sets the subtitle language preference. |
| `DisplayMissingEpisodes` | `boolean` | ❌ |  |
| `GroupedFolders` | array of `string` (uuid) | ❌ |  |
| `SubtitleMode` | `SubtitlePlaybackMode` | ❌ | An enum representing a subtitle playback mode. Enum: [`Default`, `Always`, `OnlyForced`, `None`, `Smart`] |
| `DisplayCollectionsView` | `boolean` | ❌ |  |
| `EnableLocalPassword` | `boolean` | ❌ |  |
| `OrderedViews` | array of `string` (uuid) | ❌ |  |
| `LatestItemsExcludes` | array of `string` (uuid) | ❌ |  |
| `MyMediaExcludes` | array of `string` (uuid) | ❌ |  |
| `HidePlayedInLatest` | `boolean` | ❌ |  |
| `RememberAudioSelections` | `boolean` | ❌ |  |
| `RememberSubtitleSelections` | `boolean` | ❌ |  |
| `EnableNextEpisodeAutoPlay` | `boolean` | ❌ |  |
| `CastReceiverId` | `string` | ❌ | Gets or sets the id of the selected cast receiver. |

**Example Body**:

```json
{
  "AudioLanguagePreference": "string",
  "PlayDefaultAudioTrack": false,
  "SubtitleLanguagePreference": "string",
  "DisplayMissingEpisodes": false,
  "GroupedFolders": [
    "string"
  ],
  "SubtitleMode": null,
  "DisplayCollectionsView": false,
  "EnableLocalPassword": false,
  "OrderedViews": [
    "string"
  ],
  "LatestItemsExcludes": [
    "string"
  ],
  "MyMediaExcludes": [
    "string"
  ],
  "HidePlayedInLatest": false,
  "RememberAudioSelections": false,
  "RememberSubtitleSelections": false,
  "EnableNextEpisodeAutoPlay": false
}
```

### Responses

#### Status `204`: User configuration updated.

#### Status `401`: Unauthorized

#### Status `403`: User configuration update forbidden.

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
curl -X POST "http://localhost/Users/Configuration?userId=<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "AudioLanguagePreference": "string",
  "PlayDefaultAudioTrack": false,
  "SubtitleLanguagePreference": "string",
  "DisplayMissingEpisodes": false,
  "GroupedFolders": [
    "string"
  ],
  "SubtitleMode": null,
  "DisplayCollectionsView": false,
  "EnableLocalPassword": false,
  "OrderedViews": [
    "string"
  ],
  "LatestItemsExcludes": [
    "string"
  ],
  "MyMediaExcludes": [
    "string"
  ],
  "HidePlayedInLatest": false,
  "RememberAudioSelections": false,
  "RememberSubtitleSelections": false,
  "EnableNextEpisodeAutoPlay": false
}' \
```


---

## `POST` /Users/ForgotPassword

**Initiates the forgot password process for a local user.**

**Operation ID**: `ForgotPassword`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `EnteredUsername` | `string` | ✅ | Gets or sets the entered username to have its password reset. |

**Example Body**:

```json
{
  "EnteredUsername": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `EnteredUsername` | `string` | ✅ | Gets or sets the entered username to have its password reset. |

**Example Body**:

```json
{
  "EnteredUsername": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `EnteredUsername` | `string` | ✅ | Gets or sets the entered username to have its password reset. |

**Example Body**:

```json
{
  "EnteredUsername": "string"
}
```

### Responses

#### Status `200`: Password reset process started.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Action` | `ForgotPasswordAction` | ❌ | Gets or sets the action. Enum: [`ContactAdmin`, `PinCode`, `InNetworkRequired`] |
| `PinFile` | `string` | ❌ | Gets or sets the pin file. |
| `PinExpirationDate` | `string` (date-time) | ❌ | Gets or sets the pin expiration date. |

**Example Response**:

```json
{
  "Action": null,
  "PinFile": "string",
  "PinExpirationDate": "2024-01-01T00:00:00Z"
}
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/Users/ForgotPassword" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "EnteredUsername": "string"
}' \
```


---

## `POST` /Users/ForgotPassword/Pin

**Redeems a forgot password pin.**

**Operation ID**: `ForgotPasswordPin`

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Pin` | `string` | ✅ | Gets or sets the entered pin to have the password reset. |

**Example Body**:

```json
{
  "Pin": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Pin` | `string` | ✅ | Gets or sets the entered pin to have the password reset. |

**Example Body**:

```json
{
  "Pin": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Pin` | `string` | ✅ | Gets or sets the entered pin to have the password reset. |

**Example Body**:

```json
{
  "Pin": "string"
}
```

### Responses

#### Status `200`: Pin reset process started.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Success` | `boolean` | ❌ | Gets or sets a value indicating whether this MediaBrowser.Model.Users.PinRedeemResult is success. |
| `UsersReset` | array of `string` | ❌ | Gets or sets the users reset. |

**Example Response**:

```json
{
  "Success": false,
  "UsersReset": [
    "string"
  ]
}
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X POST "http://localhost/Users/ForgotPassword/Pin" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Pin": "string"
}' \
```


---

## `GET` /Users/Me

**Gets the user based on auth token.**

**Operation ID**: `GetCurrentUser`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Responses

#### Status `200`: User returned.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Response**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
}
```

#### Status `400`: Token is not owned by a user.

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

#### Status `401`: Unauthorized

#### Status `403`: Forbidden

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/Users/Me" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---

## `POST` /Users/New

**Creates a user.**

**Operation ID**: `CreateUserByName`

**Authentication**: `CustomAuthentication` (scopes: RequiresElevation)

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ✅ | Gets or sets the username. |
| `Password` | `string` | ❌ | Gets or sets the password. |

**Example Body**:

```json
{
  "Name": "string",
  "Password": "string"
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ✅ | Gets or sets the username. |
| `Password` | `string` | ❌ | Gets or sets the password. |

**Example Body**:

```json
{
  "Name": "string",
  "Password": "string"
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ✅ | Gets or sets the username. |
| `Password` | `string` | ❌ | Gets or sets the password. |

**Example Body**:

```json
{
  "Name": "string",
  "Password": "string"
}
```

### Responses

#### Status `200`: User created.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | ❌ | Gets or sets the name. |
| `ServerId` | `string` | ❌ | Gets or sets the server identifier. |
| `ServerName` | `string` | ❌ | Gets or sets the name of the server.

This is not used by the server and is for client-side usage only. |
| `Id` | `string` (uuid) | ❌ | Gets or sets the id. |
| `PrimaryImageTag` | `string` | ❌ | Gets or sets the primary image tag. |
| `HasPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has password. |
| `HasConfiguredPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured password. |
| `HasConfiguredEasyPassword` | `boolean` | ❌ | Gets or sets a value indicating whether this instance has configured easy password. |
| `EnableAutoLogin` | `boolean` | ❌ | Gets or sets whether async login is enabled or not. |
| `LastLoginDate` | `string` (date-time) | ❌ | Gets or sets the last login date. |
| `LastActivityDate` | `string` (date-time) | ❌ | Gets or sets the last activity date. |
| `Configuration` | `UserConfiguration` | ❌ | Class UserConfiguration. |
| `Policy` | `UserPolicy` | ❌ | Gets or sets the policy. |
| `PrimaryImageAspectRatio` | `number` (double) | ❌ | Gets or sets the primary image aspect ratio. |

**Example Response**:

```json
{
  "Name": "string",
  "ServerId": "string",
  "ServerName": "string",
  "Id": "550e8400-e29b-41d4-a716-446655440000",
  "PrimaryImageTag": "string",
  "HasPassword": false,
  "HasConfiguredPassword": false,
  "HasConfiguredEasyPassword": false,
  "EnableAutoLogin": false,
  "LastLoginDate": "2024-01-01T00:00:00Z",
  "LastActivityDate": "2024-01-01T00:00:00Z",
  "Configuration": null,
  "Policy": null,
  "PrimaryImageAspectRatio": 0
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
curl -X POST "http://localhost/Users/New" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "Name": "string",
  "Password": "string"
}' \
```


---

## `POST` /Users/Password

**Updates a user's password.**

**Operation ID**: `UpdateUserPassword`

**Authentication**: `CustomAuthentication` (scopes: DefaultAuthorization)

### Request Parameters

| Name | In | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `userId` | query | `string` (uuid) | ❌ | The user id. |

### Request Body

**Required**: Yes

**Content-Type**: `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `CurrentPassword` | `string` | ❌ | Gets or sets the current sha1-hashed password. |
| `CurrentPw` | `string` | ❌ | Gets or sets the current plain text password. |
| `NewPw` | `string` | ❌ | Gets or sets the new plain text password. |
| `ResetPassword` | `boolean` | ❌ | Gets or sets a value indicating whether to reset the password. |

**Example Body**:

```json
{
  "CurrentPassword": "string",
  "CurrentPw": "string",
  "NewPw": "string",
  "ResetPassword": false
}
```

**Content-Type**: `text/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `CurrentPassword` | `string` | ❌ | Gets or sets the current sha1-hashed password. |
| `CurrentPw` | `string` | ❌ | Gets or sets the current plain text password. |
| `NewPw` | `string` | ❌ | Gets or sets the new plain text password. |
| `ResetPassword` | `boolean` | ❌ | Gets or sets a value indicating whether to reset the password. |

**Example Body**:

```json
{
  "CurrentPassword": "string",
  "CurrentPw": "string",
  "NewPw": "string",
  "ResetPassword": false
}
```

**Content-Type**: `application/*+json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `CurrentPassword` | `string` | ❌ | Gets or sets the current sha1-hashed password. |
| `CurrentPw` | `string` | ❌ | Gets or sets the current plain text password. |
| `NewPw` | `string` | ❌ | Gets or sets the new plain text password. |
| `ResetPassword` | `boolean` | ❌ | Gets or sets a value indicating whether to reset the password. |

**Example Body**:

```json
{
  "CurrentPassword": "string",
  "CurrentPw": "string",
  "NewPw": "string",
  "ResetPassword": false
}
```

### Responses

#### Status `204`: Password successfully reset.

#### Status `401`: Unauthorized

#### Status `403`: User is not allowed to update the password.

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

#### Status `404`: User not found.

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
curl -X POST "http://localhost/Users/Password?userId=<userId>" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
  "CurrentPassword": "string",
  "CurrentPw": "string",
  "NewPw": "string",
  "ResetPassword": false
}' \
```


---

## `GET` /Users/Public

**Gets a list of publicly visible users for display on a login screen.**

**Operation ID**: `GetPublicUsers`

### Responses

#### Status `200`: Public users returned.



**Example Response**:

```json
[
  {
    "Name": "string",
    "ServerId": "string",
    "ServerName": "string",
    "Id": "550e8400-e29b-41d4-a716-446655440000",
    "PrimaryImageTag": "string",
    "HasPassword": false,
    "HasConfiguredPassword": false,
    "HasConfiguredEasyPassword": false,
    "EnableAutoLogin": false,
    "LastLoginDate": "2024-01-01T00:00:00Z",
    "LastActivityDate": "2024-01-01T00:00:00Z",
    "Configuration": null,
    "Policy": null,
    "PrimaryImageAspectRatio": 0
  }
]
```

#### Status `503`: The server is currently starting or is temporarily not available.

**Response Headers**:

- `Retry-After`: A hint for when to retry the operation in full seconds.
- `Message`: A short plain-text reason why the server is not available.

### Usage Example

```bash
curl -X GET "http://localhost/Users/Public" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
```


---
