# Endpoints

Quick links:

- [Session](#session)

## Session

Endpoints to store, validate or delete user sessions from the database.

- [Create Session](#create-session)
- [Validate Session](#validate-session)
- [Invalidate Session](#invalidate-session)

### **Create Session**

#### `POST /sessions`

This endpoint is used to create a new session for a user.

##### Request Body:

The body should be a JSON object containing the following field:

- `user_id` (required): The ID of the user who is creating the session. It should be an integer.

```json
{
  "user_id": 12345
}
```

##### Possible Responses:

- 200 OK: Successfully created a session. The response contains the session details.

```json
{
  "ok": true,
  "data": {
    "session_token": "abc123"
  }
}
```

- 400 Bad Request: Invalid request body.

```json
{
  "ok": false,
  "data": "Invalid request body. <error_message>"
}
```

- 405 Method Not Allowed: If the HTTP method is not POST.

```json
{
  "ok": false,
  "data": "Method not allowed"
}
```

- 500 Internal Server Error: Failed to create the session.

```json
{
  "ok": false,
  "data": "Failed to create session. <error_message>"
}
```

### Validate Session

#### `GET /sessions/validate`

This endpoint is used to validate an existing session.

##### Request Headers:

session: The session token (required).

##### Possible Responses:

- 200 OK: Session is valid.

```json
{
  "ok": true,
  "data": "Session is valid"
}
```

- 401 Unauthorized:

If the session token is empty:

```json
{
  "ok": false,
  "data": "Session token is empty"
}
```

If the session is not found:

```json
{
  "ok": false,
  "data": "Session not found"
}
```

If the session is expired:

```json
{
  "ok": false,
  "data": "Session expired"
}
```

If the user associated with the session is not found:

```json
{
  "ok": false,
  "data": "User associated with session not found"
}
```

- 500 Internal Server Error: Failed to validate the session.

```json
{
  "ok": false,
  "data": "Failed to validate session. <error_message>"
}
```

### Invalidate Session

#### `POST /sessions/invalidate`

This endpoint is used to invalidate an existing session.

##### Request Headers:

session: The session token (required).

##### Possible Responses:

- 200 OK: Successfully invalidated the session.

```json
{
  "ok": true,
  "data": "Session invalidated"
}
```

- 401 Unauthorized: If the session token is empty.

```json
{
  "ok": false,
  "data": "Session token is empty"
}
```

- 500 Internal Server Error: Failed to invalidate the session.

```json
{
  "ok": false,
  "data": "Failed to invalidate session. <error_message>"
}
```
