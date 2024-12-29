# Development

## Run

To run the appication you need running **docker** with **docker compose**

You also need to have local image of file-storage build and stored. The tag for image should be `maxit/file-storage`. Refer to [documentation](https://github.com/mini-maxit/file-storage?tab=readme-ov-file#build) on how to do it.

```bash
docker compose up --build -d
```

# Endpoints

Quick links:

- [Error format](#error)
- [Task](#task)
- [Session](#session)
- [Auth](#auth)

All endpoints are prefixed with `/api/v1` prefix. For example: example.com/api/v1/task

## Error

All endpoints return errors in the same format. JSON as below, and sets corresponding HTTP status code:

```json
{
  "ok": false,
  "data": {
    "code": "Bad Request",
    "message": "Invalid user ID."
  }
}
```

## Task

### 1. **Get All Tasks**

Session: The session token (required).

#### `GET /task/`

**Headers**

Session: The session token (required).

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of tasks.

```json
{
  "ok": true,
  "data": [
    {
      "id": 1,
      "title": "Example Task",
      "created_by": 123,
      "created_at": "2024-11-18T19:15:29.997499Z"
    },
    {
      "id": 2,
      "title": "Another Task",
      "created_by": 456,
      "created_at": "2024-11-18T19:15:29.997499Z"
    }
  ]
}
```

- **500 Internal Server Error**: An error occurred while retrieving the tasks.

```json
{
  "ok": false,
  "data": {
    "code": "Internal Server Error",
    "message": "Error getting tasks. Database connection failed."
  }
}
```

---

### 2. **Get Task by ID**

#### `GET /task/{id}`

Retrieves the details of a specific task by its ID.

**Request Parameters:**

- **Path Parameter**:
  `id` (required) - The ID of the task to retrieve.

**Possible Responses:**

- **200 OK**: Successfully retrieved the task details.

```json
{
  "ok": true,
  "data": {
    "id": 1,
    "title": "Example Task",
    "description_url": "http://file-storage:8888/getTaskDescription&?taskID=2", // This Url should be used to fetch the descirption file. Be aware that you can only do it on server side.
    "created_by_name": "Name",
    "created_by": 123
  }
}
```

- **400 Bad Request**: Invalid task ID.

```json
{
  "ok": false,
  "data": {
    "code": "Bad Request",
    "message": "Invalid task ID."
  }
}
```

- **500 Internal Server Error**: An error occurred while retrieving the task.

```json
{
  "ok": false,
  "data": {
    "code": "Internal Server Error",
    "message": "Error getting task. record not found"
  }
}
```

---

### 3. **Upload Task**

#### `POST /task/`

Uploads a new task.

**Request Parameters:**

- **Form Data**:
  - `taskName` (required): The name of the task.
  - `userId` (required): The ID of the user uploading the task.
  - `overwrite` (optional): Boolean flag to indicate if the task should be overwritten.
  - `archive` (required): The task file to upload (must be `.zip` or `.tar.gz`).

**Possible Responses:**

- **200 OK**: Task uploaded successfully.

```json
{
  "ok": true,
  "data": {
    "id": 6
  }
}
```

- **400 Bad Request**: Invalid request parameters or file format.

- **500 Internal Server Error**: An error occurred during the task upload process.

```json
{
  "ok": false,
  "data": {
    "code": "Internal Server Error",
    "message": "Error sending file to FileStorage service. Connection timeout."
  }
}
```

---

### 4. **WIP (NOT UPDATED)** Submit Solution

#### `POST /task/submit`

Submits a solution for a task.

**Request Parameters:**

- **Form Data**:
  - `taskID` (required): The ID of the task for which the solution is being submitted.
  - `userID` (required): The ID of the user submitting the solution.
  - `languageID` (required): The programming language ID of the solution.
  - `solution` (required): The solution file.

**Possible Responses:**

Unexpected behaviour <3

## Session

Endpoints to store, validate or delete user sessions from the database.

- [Create Session](#create-session)
- [Validate Session](#validate-session)
- [Invalidate Session](#invalidate-session)

### **Create Session**

#### `POST /session/` (DEPRECATED)

This endpoint is used to create a new session for a user. However, this is huge security issue and should not be used in production. The session is created after login and registration, and only mentioned methods should be used!

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
    "session": "asdflkjasdlfjk",
    "expires_at": "2024-12-19T16:19:07.756644Z",
    "userId": 1
  }
}
```

### Validate Session

#### `GET /session/validate`

This endpoint is used to validate an existing session.

##### Request headers:

Session: the session token (required).

##### Possible Responses:

- 200 OK: Session is valid.

```json
{
  "ok": true,
  "data": {
    "valid": true,
    "user_id": 123
  }
}
```

- 401 Unauthorized: ff the session token is empty or session is invalid

- 500 Internal Server Error: Failed to validate the session.

### Invalidate Session

#### `POST /session/invalidate`

This endpoint is used to invalidate an existing session.

##### Request Headers:

Session: The session token (required).

##### Possible Responses:

- 200 OK: Successfully invalidated the session.

```json
{
  "ok": true,
  "data": "Session invalidated"
}
```

- 401 Unauthorized: If the session token is empty.

- 500 Internal Server Error: Failed to invalidate the session.

## Auth

### **Login**

Handles user authentication by validating credentials and returning a session.

#### `POST /auth/login`

##### Request Body:

```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

- `email` _(string)_: The user's registered email.
- `password` _(string)_: The user's password.

##### Responses

- **200 OK**:

```json
{
  "ok": true,
  "data": {
    "session": "asdflkjasdlfjk",
    "expires_at": "2024-12-19T16:19:07.756644Z",
    "userId": 1
  }
}
```

Returns a session with a token and expiration information.

- **400 Bad Request**: Triggered when the request body is invalid.

- **401 Unauthorized**: Triggered when the email does not exist or password is invalid.

- **405 Method Not Allowed**: Triggered when a non-`POST` request is made.

- **500 Internal Server Error**: Triggered when an unexpected server error occurs.

---

#### Register

Registers a new user and returns a session upon successful registration.

##### `POST /auth/register`

###### Request Body

```json
{
  "name": "name",
  "surname": "surname",
  "email": "user@example.com",
  "password": "securepassword",
  "username": "User Name"
}
```

- `email` _(string)_: The user's email address.
- `password` _(string)_: The user's desired password.
- `name` _(string)_: The user's name.
- `surname` _(string)_: The user's surname.
- `username` _(string)_: The user's username.

##### Responses

- **201 Created**:

  ```json
  {
    "ok": true,
    "data": {
      "session": "asdflkjasdlfjk",
      "expires_at": "2024-12-19T16:19:07.756644Z",
      "userId": 1
    }
  }
  ```

  Returns a session with a token and expiration information.

- **400 Bad Request**: Triggered when the request body is invalid.

- **409 Conflict**: Triggered when the provided email is already registered.

- **405 Method Not Allowed**: Triggered when a non-`POST` request is made.

- **500 Internal Server Error**: Triggered when an unexpected server error occurs.
