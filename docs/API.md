# Factorio Server Manager API Documentation

## Overview

The Factorio Server Manager provides a REST API for managing Factorio game servers. All API endpoints are prefixed with `/api` and require authentication unless otherwise noted.

## Authentication

### Session-Based Authentication

The API uses session-based authentication with secure cookies. Users must log in to obtain a session before accessing protected endpoints.

### CSRF Protection

All state-changing requests (POST, PUT, DELETE) require a valid CSRF token. Obtain a token from the `/api/csrf-token` endpoint and include it in the `X-CSRF-Token` header.

### Rate Limiting

The login endpoint is rate-limited to prevent brute force attacks.

---

## Endpoints

### Authentication

#### Login
```
POST /api/login
```

Authenticates a user and creates a session.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "username": "string",
  "role": "string",
  "email": "string"
}
```

**Note:** This endpoint is rate-limited and does not require authentication.

---

#### Logout
```
GET /api/logout
```

Terminates the current user session.

**Response:**
```json
"User logged out successfully."
```

---

#### Get CSRF Token
```
GET /api/csrf-token
```

Returns a CSRF token for use in subsequent requests.

**Note:** This endpoint does not require authentication.

---

#### Get Current User
```
GET /api/user/status
```

Returns information about the currently logged-in user.

**Response:**
```json
{
  "username": "string",
  "role": "string",
  "email": "string"
}
```

---

### User Management

#### List Users
```
GET /api/user/list
```

Returns a list of all registered users.

**Response:**
```json
[
  {
    "username": "string",
    "role": "string",
    "email": "string"
  }
]
```

---

#### Add User
```
POST /api/user/add
```

Creates a new user account.

**Request Body:**
```json
{
  "username": "string",
  "password": "string",
  "role": "string",
  "email": "string"
}
```

---

#### Remove User
```
POST /api/user/remove
```

Deletes a user account.

**Request Body:**
```json
{
  "username": "string"
}
```

---

#### Change Password
```
POST /api/user/password
```

Changes the password for the currently logged-in user.

**Request Body:**
```json
{
  "old_password": "string",
  "new_password": "string",
  "new_password_confirmation": "string"
}
```

---

### Server Control

#### Start Server
```
POST /api/server/start
```

Starts the Factorio game server.

**Requires:** Server must be stopped.

**Request Body:**
```json
{
  "savefile": "string",
  "bindip": "string",
  "port": 34197
}
```

**Response:**
```json
"Factorio server with save: {savefile} started on port: {port}"
```

---

#### Stop Server
```
GET /api/server/stop
```

Gracefully stops the Factorio game server.

---

#### Kill Server
```
GET /api/server/kill
```

Forcefully terminates the Factorio game server process.

---

#### Check Server Status
```
GET /api/server/status
```

Returns the current status of the Factorio server.

**Response (running):**
```json
{
  "status": "running",
  "port": "34197",
  "savefile": "my-save.zip",
  "address": "0.0.0.0"
}
```

**Response (stopped):**
```json
{
  "status": "stopped"
}
```

---

#### Get Factorio Version
```
GET /api/server/facVersion
```

Returns the installed Factorio version.

**Response:**
```json
{
  "version": "1.1.0.0",
  "base_mod_version": "1.1.0"
}
```

---

### Save Management

#### List Saves
```
GET /api/saves/list
```

Returns a list of all save files.

**Response:**
```json
[
  {
    "name": "my-save.zip",
    "last_modified": "2024-01-15T10:30:00Z",
    "size": 1234567
  }
]
```

---

#### Download Save
```
GET /api/saves/dl/{save}
```

Downloads a specific save file.

**Parameters:**
- `save` (path) - Name of the save file

**Response:** Binary file download

---

#### Upload Save
```
POST /api/saves/upload
```

Uploads a new save file.

**Content-Type:** `multipart/form-data`

**Parameters:**
- `savefile` - The save file (.zip)

---

#### Remove Save
```
GET /api/saves/rm/{save}
```

Deletes a specific save file.

**Parameters:**
- `save` (path) - Name of the save file to delete

---

#### Create Save
```
GET /api/saves/create/{save}
```

Creates a new empty save file.

**Requires:** Server must be stopped.

**Parameters:**
- `save` (path) - Name for the new save file

---

#### Load Mods from Save
```
POST /api/saves/mods
```

Installs mods required by a specific save file.

**Requires:** Server must be stopped.

**Request Body:**
```json
{
  "save_file": "string"
}
```

---

### Server Settings

#### Get Server Settings
```
GET /api/settings
```

Returns the current server settings (server-settings.json).

---

#### Update Server Settings
```
POST /api/settings/update
```

Updates the server settings.

**Request Body:** Server settings JSON object

---

### Configuration

#### Load Config
```
GET /api/config
```

Returns the current Factorio config.ini contents.

---

### Logs

#### Get Log Tail
```
GET /api/log/tail
```

Returns the last lines of the Factorio server log.

---

### Mod Management

#### List Installed Mods
```
GET /api/mods/list
```

Returns a list of all installed mods.

**Response:**
```json
{
  "mods": [
    {
      "name": "mod-name",
      "version": "1.0.0",
      "title": "Mod Title",
      "author": "Author Name",
      "file_name": "mod-name_1.0.0.zip",
      "factorio_version": "1.1",
      "enabled": true,
      "compatibility": true
    }
  ]
}
```

---

#### Toggle Mod
```
POST /api/mods/toggle
```

Enables or disables a mod.

**Requires:** Server must be stopped.

**Request Body:**
```json
{
  "name": "mod-name"
}
```

---

#### Delete Mod
```
POST /api/mods/delete
```

Removes a mod from the server.

**Requires:** Server must be stopped.

**Request Body:**
```json
{
  "name": "mod-name"
}
```

---

#### Delete All Mods
```
POST /api/mods/delete/all
```

Removes all installed mods.

**Requires:** Server must be stopped.

---

#### Update Mod
```
POST /api/mods/update
```

Updates a mod to a newer version.

**Requires:** Server must be stopped.

**Request Body:**
```json
{
  "name": "mod-name",
  "download_url": "string",
  "file_name": "string"
}
```

---

#### Upload Mod
```
POST /api/mods/upload
```

Uploads and installs a mod file.

**Requires:** Server must be stopped.

**Content-Type:** `multipart/form-data`

**Parameters:**
- `mod_file` - The mod file (.zip)

---

#### Download Mods
```
GET /api/mods/download
```

Downloads all installed mods as a zip archive.

---

### Mod Portal

#### List Portal Mods
```
GET /api/mods/portal/list
```

Returns a list of all mods available on the Factorio mod portal.

---

#### Get Portal Mod Info
```
GET /api/mods/portal/info/{mod}
```

Returns detailed information about a mod from the portal.

**Parameters:**
- `mod` (path) - Mod ID/name

---

#### Install from Portal
```
POST /api/mods/portal/install
```

Installs a mod from the Factorio mod portal.

**Requires:** Server must be stopped. Factorio portal login required.

**Request Body:**
```json
{
  "name": "mod-name",
  "download_url": "string",
  "file_name": "string"
}
```

---

#### Install Multiple from Portal
```
POST /api/mods/portal/install/multiple
```

Installs multiple mods from the portal at once.

**Requires:** Server must be stopped. Factorio portal login required.

**Request Body:**
```json
{
  "mods": [
    {
      "name": "mod-name",
      "download_url": "string",
      "file_name": "string"
    }
  ]
}
```

---

#### Portal Login
```
POST /api/mods/portal/login
```

Authenticates with the Factorio mod portal.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

---

#### Portal Login Status
```
GET /api/mods/portal/loginstatus
```

Checks if the user is logged into the mod portal.

---

#### Portal Logout
```
GET /api/mods/portal/logout
```

Logs out from the Factorio mod portal.

---

### Mod Packs

Mod packs allow you to save and switch between different mod configurations.

#### List Mod Packs
```
GET /api/mods/packs/list
```

Returns a list of all saved mod packs.

---

#### Create Mod Pack
```
POST /api/mods/packs/create
```

Creates a new mod pack from the currently installed mods.

**Request Body:**
```json
{
  "name": "pack-name"
}
```

---

#### Delete Mod Pack
```
POST /api/mods/packs/{modpack}/delete
```

Deletes a mod pack.

**Parameters:**
- `modpack` (path) - Name of the mod pack

---

#### Download Mod Pack
```
GET /api/mods/packs/{modpack}/download
```

Downloads a mod pack as a zip archive.

**Parameters:**
- `modpack` (path) - Name of the mod pack

---

#### Load Mod Pack
```
POST /api/mods/packs/{modpack}/load
```

Loads a mod pack, replacing the current mods.

**Requires:** Server must be stopped.

**Parameters:**
- `modpack` (path) - Name of the mod pack to load

---

### Mod Pack Contents

#### List Mods in Pack
```
GET /api/mods/packs/{modpack}/list
```

Returns the list of mods in a specific mod pack.

---

#### Toggle Mod in Pack
```
POST /api/mods/packs/{modpack}/mod/toggle
```

Enables or disables a mod within a mod pack.

---

#### Delete Mod from Pack
```
POST /api/mods/packs/{modpack}/mod/delete
```

Removes a mod from a mod pack.

---

#### Delete All Mods from Pack
```
POST /api/mods/packs/{modpack}/mod/delete/all
```

Removes all mods from a mod pack.

---

#### Update Mod in Pack
```
POST /api/mods/packs/{modpack}/mod/update
```

Updates a mod within a mod pack.

---

#### Upload Mod to Pack
```
POST /api/mods/packs/{modpack}/mod/upload
```

Uploads a mod directly to a mod pack.

---

#### Install from Portal to Pack
```
POST /api/mods/packs/{modpack}/portal/install
```

Installs a mod from the portal directly into a mod pack.

---

#### Install Multiple from Portal to Pack
```
POST /api/mods/packs/{modpack}/portal/install/multiple
```

Installs multiple mods from the portal into a mod pack.

---

## WebSocket

### Connection
```
GET /ws
```

Establishes a WebSocket connection for real-time updates.

**Note:** Requires authentication.

### Message Format

Messages are JSON objects with the following structure:

```json
{
  "room_name": "string",
  "message": "any",
  "controls": {
    "type": "string",
    "value": "string"
  }
}
```

### Control Types

- `subscribe` - Subscribe to a room for updates
- `unsubscribe` - Unsubscribe from a room
- `command` - Send a command to the Factorio server console

### Rooms

- `gamelog` - Real-time game console output
- `server_status` - Server status change notifications
- `server_errors` - Server error notifications

---

## Error Responses

All endpoints return error responses in the following format:

**HTTP Status Codes:**
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (not logged in)
- `403` - Forbidden (CSRF token invalid)
- `404` - Not Found
- `409` - Conflict (e.g., server already running)
- `415` - Unsupported Media Type (invalid file format)
- `423` - Locked (operation requires server to be stopped)
- `429` - Too Many Requests (rate limited)
- `500` - Internal Server Error

**Error Response Body:**
```json
"Error message describing what went wrong"
```

---

## Security Notes

1. Always use HTTPS in production
2. CSRF tokens are required for all POST requests
3. Session cookies are HttpOnly and Secure (in production)
4. File uploads are validated for correct extensions
5. Path traversal attacks are prevented through input validation
6. Login endpoint is rate-limited to prevent brute force attacks
