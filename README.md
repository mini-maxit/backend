# Development

## Run

To run the appication you need running **docker** with **docker compose**

You also need to have local image of file-storage build and stored. The tag for image should be `maxit/file-storage`. Refer to [documentation](https://github.com/mini-maxit/file-storage?tab=readme-ov-file#build) on how to do it.

```bash
docker compose up --build -d
```

# Endpoints

Quick links:

- [Task](#task)
- [Session](#session)
- [Auth](#auth)
- [Submission](#submission)
- [User](#user)

All endpoints are prefixed with `/api/v1` prefix. For example: example.com/api/v1/task

## Task

### 1. **Get All Tasks**

#### `GET /task`

Retrieves a list of all tasks.

**Request Parameters:**
- **Query Parameters:**
  - **Filtering Options:**
    - `id` (optional) - Filter tasks by ID.
    - `title` (optional) - Filter tasks by title.
    - `created_by` (optional) - Filter tasks by user ID.
    - `created_at` (optional) - Filter tasks by creation date.
    - `order` (optional) - Sort tasks by order.
  - **Pagination Options:**
    - `offset` (optional) - The number of tasks to skip. Default is 0.
    - `limit` (optional) - The number of tasks per page. Default is 10.
  - **Sorting Options:**
    - `sort` (optional) - The field to sort by. Default is `created_at desc`.

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
  "error": "Error getting tasks. Database connection failed."
}
```

---

### 2. **Get Task by ID**

#### `GET /task/{id}`

Retrieves the details of a specific task by its ID.

**Request Parameters:**

- **Path Parameter**:
  `id` (required) - The ID of the task to retrieve.
- **Query Parameters:**
  - **Filtering Options:**
    - `id` (optional) - Filter tasks by ID.
    - `title` (optional) - Filter tasks by title.
    - `created_by` (optional) - Filter tasks by user ID.
    - `created_at` (optional) - Filter tasks by creation date.
    - `order` (optional) - Sort tasks by order.
  - **Pagination Options:**
    - `offset` (optional) - The number of tasks to skip. Default is 0.
    - `limit` (optional) - The number of tasks per page. Default is 10.
  - **Sorting Options:**
    - `sort` (optional) - The field to sort by. Default is `created_at desc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the task details.

```json
{
  "ok": true,
  "data": {
    "id": 1,
    "title": "Example Task",
    "description_url": "http://file-storage:8888/getTaskDescription&?taskID=2", // This Url should be used to fetch the descirption file. Be aware that you can only do it on server side.
    "created_by": 123
  }
}
```

- **400 Bad Request**: Invalid or missing task ID.

```json
{
  "ok": false,
  "error": "Task ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the task.

```json
{
  "ok": false,
  "error": "Error getting task. Task not found."
}
```

---

### 3. **Upload Task**

#### `POST /task/upload`

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
  "data": "Task uploaded successfully."
}
```

- **400 Bad Request**: Invalid request parameters or file format.

```json
{
  "ok": false,
  "error": "Invalid file format. Only .zip and .tar.gz files are allowed."
}
```

- **500 Internal Server Error**: An error occurred during the task upload process.

```json
{
  "ok": false,
  "error": "Error sending file to FileStorage service. Connection timeout."
}
```

---

### 4. **Submit Solution**

#### `POST /task/submit`

Submits a solution for a task.

**Request Parameters:**

- **Form Data**:
  - `taskID` (required): The ID of the task for which the solution is being submitted.
  - `userID` (required): The ID of the user submitting the solution.
  - `languageID` (required): The programming language ID of the solution.
  - `solution` (required): The solution file.

**Possible Responses:**

- **200 OK**: Solution submitted successfully.

```json
{
  "ok": true,
  "data": "Solution submitted successfully."
}
```

- **400 Bad Request**: Invalid request parameters.

```json
{
  "ok": false,
  "error": "Task ID is required."
}
```

- **500 Internal Server Error**: An error occurred during the submission process.

```json
{
  "ok": false,
  "error": "Error publishing submission to the queue. RabbitMQ not available."
}
```

---

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

#### `GET /session/validate`

This endpoint is used to validate an existing session.

##### Request Headers:

session: The session token (required).

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

#### `POST /session/invalidate`

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
    "session": "asdflkjasdlfjk",
    "expires_at": "2024-12-19T16:19:07.756644Z",
    "userId": 1
  }
  ```

  Returns a session with a token and expiration information.

- **400 Bad Request**:

  ```json
  {
    "error": "Invalid request body. [details]"
  }
  ```

  Triggered when the request body is invalid.

- **401 Unauthorized**:

  ```json
  {
    "error": "Given email does not exist. Verify your email and try again."
  }
  ```

  Triggered when the email does not exist.

  ```json
  {
    "error": "Invalid credentials. Verify your email and password and try again."
  }
  ```

  Triggered when the credentials are incorrect.

- **405 Method Not Allowed**:

  ```json
  {
    "error": "Method not allowed"
  }
  ```

  Triggered when a non-`POST` request is made.

- **500 Internal Server Error**:
  ```json
  {
    "error": "Failed to login. [details]"
  }
  ```
  Triggered when an unexpected server error occurs.

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
    "session": "asdflkjasdlfjk",
    "expires_at": "2024-12-19T16:19:07.756644Z",
    "userId": 1
  }
  ```

  Returns a session with a token and expiration information.

- **400 Bad Request**:

  ```json
  {
    "error": "Invalid request body. [details]"
  }
  ```

  Triggered when the request body is invalid.

- **409 Conflict**:

  ```json
  {
    "error": "Email already exists. Please login."
  }
  ```

  Triggered when the provided email is already registered.

- **405 Method Not Allowed**:

  ```json
  {
    "error": "Method not allowed"
  }
  ```

  Triggered when a non-`POST` request is made.

- **500 Internal Server Error**:
  ```json
  {
    "error": "Failed to register. [details]"
  }
  ```
  Triggered when an unexpected server error occurs.


## Submission

### **Get All Submissions**

#### `GET /submission`

Retrieves a list of all submissions.

**Role-based Access:**

- **Admin**: Retrieves all submissions.
- **Teacher**: Retrieves submissions for tasks created by the teacher.
- **Student**: Retrieves only the student's own submissions.

**Request Parameters:**

- **Query Parameters:**
  - **Filtering Options:**
    - `task_id` (optional) - Filter submissions by task ID.
    - `user_id` (optional) - Filter submissions by user ID.
    - `order` (optional) - Sort submissions by order.
    - `language_id` (optional) - Filter submissions by language ID.
    - `status` (optional) - Filter submissions by status.
    - `submitted_at` (optional) - Filter submissions by submission date.
    - `check_at` (optional) - Filter submissions by check date.

  - **Pagination Options:**
    - `offset` (optional) - The number of submissions to skip. Default is 0.
    - `limit` (optional) - The number of submissions per page. Default is 10.

  - **Sorting Options:**
    - `sort` (optional) - The field to sort by. Default is `submitted_at desc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of submissions.

```json
{
    "ok": true,
    "data": [
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        }
    ]
}
```

- **500 Internal Server Error**: An error occurred while retrieving the submissions.

```json
{
  "ok": false,
  "error": "Error getting submissions. Database connection failed."
}
```

---

### **Get Submission by ID**

#### `GET /submission/{id}`

Retrieves the details of a specific submission by its ID.

**Role-based Access:**

- **Admin**: Can view all submissions.
- **Teacher**: Can view submissions for tasks created by the teacher.
- **Student**: Can view only the student's own submissions.

- **Path Parameter**:
  `id` (required) - The ID of the submission to retrieve.

**Possible Responses:**

- **200 OK**: Successfully retrieved the submission details.

```json
{
    "ok": true,
    "data": [
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        }
    ]
}
```

- **400 Bad Request**: Invalid or missing submission ID.

```json
{
  "ok": false,
  "error": "Submission ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the submission.

```json
{
  "ok": false,
  "error": "Error getting submission. Submission not found."
}
```

---

### **Get All Submissions for User**

#### `GET /submission/user/{id}`

Retrieves a list of all submissions for a specific user.

**Role-based Access:**

- **Admin**: Can view all submissions for the specified user.
- **Teacher**: Can view submissions for tasks created by the teacher.
- **Student**: Can view only the student's own submissions.

**Request Parameters:**

  - **Path Parameter**:
    `id` (required) - The ID of the user whose submissions to retrieve.

  - **Query Parameters:**
    - **Filtering Options:**
      - `task_id` (optional) - Filter submissions by task ID.
      - `user_id` (optional) - Filter submissions by user ID.
      - `order` (optional) - Sort submissions by order.
      - `language_id` (optional) - Filter submissions by language ID.
      - `status` (optional) - Filter submissions by status.
      - `submitted_at` (optional) - Filter submissions by submission date.
      - `check_at` (optional) - Filter submissions by check date.

    - **Pagination Options:**
      - `offset` (optional) - The number of submissions to skip. Default is 0.
      - `limit` (optional) - The number of submissions per page. Default is 10.

    - **Sorting Options:**
      - `sort` (optional) - The field to sort by. Default is `submitted_at desc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of submissions.

```json
{
    "ok": true,
    "data": [
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        },
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        }
    ]
}
```

- **400 Bad Request**: Invalid or missing user ID.

```json
{
  "ok": false,
  "error": "User ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the submissions.

```json
{
  "ok": false,
  "error": "Error getting submissions. Database connection failed."
}
```

---

### **Get All Submissions for Group**

#### `GET /submission/group/{id}`

Retrieves a list of all submissions for a specific group.

**Role-based Access:**

- **Admin**: Can view all submissions for the specified group.
- **Teacher**: Can view submissions for tasks created by the teacher.
- **Student**: Cannot view group submissions.

**Request Parameters:**

  - **Path Parameter**:
    `id` (required) - The ID of the group whose submissions to retrieve.

  - **Query Parameters:**
    - **Filtering Options:**
      - `task_id` (optional) - Filter submissions by task ID.
      - `user_id` (optional) - Filter submissions by user ID.
      - `order` (optional) - Sort submissions by order.
      - `language_id` (optional) - Filter submissions by language ID.
      - `status` (optional) - Filter submissions by status.
      - `submitted_at` (optional) - Filter submissions by submission date.
      - `check_at` (optional) - Filter submissions by check date.

    - **Pagination Options:**
      - `offset` (optional) - The number of submissions to skip. Default is 0.
      - `limit` (optional) - The number of submissions per page. Default is 10.

    - **Sorting Options:**
      - `sort` (optional) - The field to sort by. Default is `submitted_at desc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of submissions.

```json
{
    "ok": true,
    "data": [
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        },
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        }
    ]
}
```

- **400 Bad Request**: Invalid or missing group ID.

```json
{
  "ok": false,
  "error": "Group ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the submissions.

```json
{
  "ok": false,
  "error": "Error getting submissions. Database connection failed."
}
```

---

### **Get All Submissions for Task**

#### `GET /submission/task/{id}`

Retrieves a list of all submissions for a specific task.

**Role-based Access:**

- **Admin**: Can view all submissions for the specified task.
- **Teacher**: Can view submissions for tasks created by the teacher.
- **Student**: Can view only the student's own submissions.

**Request Parameters:**

**Request Parameters:**

  - **Path Parameter**:
    `id` (required) - The ID of the task whose submissions to retrieve.

  - **Query Parameters:**
    - **Filtering Options:**
      - `task_id` (optional) - Filter submissions by task ID.
      - `user_id` (optional) - Filter submissions by user ID.
      - `order` (optional) - Sort submissions by order.
      - `language_id` (optional) - Filter submissions by language ID.
      - `status` (optional) - Filter submissions by status.
      - `submitted_at` (optional) - Filter submissions by submission date.
      - `check_at` (optional) - Filter submissions by check date.

    - **Pagination Options:**
      - `offset` (optional) - The number of submissions to skip. Default is 0.
      - `limit` (optional) - The number of submissions per page. Default is 10.

    - **Sorting Options:**
      - `sort` (optional) - The field to sort by. Default is `submitted_at desc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of submissions.

```json
{
    "ok": true,
    "data": [
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        },
        {
            "id": 3,
            "task_id": 1,
            "user_id": 1,
            "order": 1,
            "language_id": 1,
            "status": "afsdf",
            "status_message": "",
            "submitted_at": "0001-01-01T00:00:00Z",
            "checked_at": null,
            "language": {
                "language": "CPP",
                "version": "20"
            },
            "task": {
                "id": 1,
                "title": "dasf",
                "created_by": 1,
                "created_at": "2025-01-05T15:00:00Z"
            },
            "user": {
                "id": 1,
                "name": "name",
                "surname": "surname",
                "email": "adfadfa@email.com",
                "username": "userfadfadname",
                "role": "student"
            }
        }
    ]
}
```

- **400 Bad Request**: Invalid or missing task ID.

```json
{
  "ok": false,
  "error": "Task ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the submissions.

```json
{
  "ok": false,
  "error": "Error getting submissions. Database connection failed."
}
```

---

### **Submit Solution**

#### `POST /submission/submit`

Submits a solution for a task.

**Request Parameters:**

- **Form Data**:
  - `taskID` (required): The ID of the task for which the solution is being submitted.
  - `userID` (required): The ID of the user submitting the solution.
  - `languageID` (required): The programming language ID of the solution.
  - `solution` (required): The solution file.

**Possible Responses:**

- **200 OK**: Solution submitted successfully.

```json
{
  "ok": true,
  "data": "Solution submitted successfully."
}
```

- **400 Bad Request**: Invalid request parameters.

```json
{
  "ok": false,
  "error": "Task ID is required."
}
```

- **500 Internal Server Error**: An error occurred during the submission process.

```json
{
  "ok": false,
  "error": "Error publishing submission to the queue. RabbitMQ not available."
}
```

## User

### **Get All Users**

#### `GET /user`

Retrieves a list of all users.

**Request Parameters:**

- **Query Parameters:**
  - **Filtering Options:**
    - `id` (optional) - Filter users by ID.
    - `name` (optional) - Filter users by name.
    - `surname` (optional) - Filter users by surname.
    - `email` (optional) - Filter users by email.
    - `username` (optional) - Filter users by username.
    - `role` (optional) - Filter users by role.
  - **Pagination Options:**
    - `offset` (optional) - The number of users to skip. Default is 0.
    - `limit` (optional) - The number of users per page. Default is 10.
  - **Sorting Options:**
    - `sort` (optional) - The field to sort by. Default is `id asc`.

**Possible Responses:**

- **200 OK**: Successfully retrieved the list of users.

```json
{
  "ok": true,
  "data": [
    {
      "id": 1,
      "name": "John",
      "surname": "Doe",
      "email": "john.doe@example.com",
      "username": "johndoe",
      "role": "student"
    },
    {
      "id": 2,
      "name": "Jane",
      "surname": "Smith",
      "email": "jane.smith@example.com",
      "username": "janesmith",
      "role": "teacher"
    }
  ]
}
```

- **500 Internal Server Error**: An error occurred while retrieving the users.

```json
{
  "ok": false,
  "error": "Error getting users. Database connection failed."
}
```

---

### **Get User by ID**

#### `GET /user/{id}`

Retrieves the details of a specific user by their ID.

**Request Parameters:**

- **Path Parameter**:
  `id` (required) - The ID of the user to retrieve.

**Possible Responses:**

- **200 OK**: Successfully retrieved the user details.

```json
{
  "ok": true,
  "data": {
    "id": 1,
    "name": "John",
    "surname": "Doe",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "role": "student"
  }
}
```

- **400 Bad Request**: Invalid or missing user ID.

```json
{
  "ok": false,
  "error": "User ID is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the user.

```json
{
  "ok": false,
  "error": "Error getting user. User not found."
}
```

---

### **Get User by Email**

#### `GET /user/email`

Retrieves the details of a specific user by their email.

**Request Parameters:**

- **Query Parameter**:
  `email` (required) - The email of the user to retrieve.

**Possible Responses:**

- **200 OK**: Successfully retrieved the user details.

```json
{
  "ok": true,
  "data": {
    "id": 1,
    "name": "John",
    "surname": "Doe",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "role": "student"
  }
}
```

- **400 Bad Request**: Invalid or missing email.

```json
{
  "ok": false,
  "error": "Email is required."
}
```

- **500 Internal Server Error**: An error occurred while retrieving the user.

```json
{
  "ok": false,
  "error": "Error getting user. User not found."
}
```

---

### **Edit User**

#### `PUT /user/{id}`

Edits the details of a specific user.

**Request Parameters:**

- **Path Parameter**:
  `id` (required) - The ID of the user to edit.

- **Request Body**:
  - `name` (optional): The new name of the user.
  - `surname` (optional): The new surname of the user.
  - `email` (optional): The new email of the user.
  - `username` (optional): The new username of the user.

**Possible Responses:**

- **200 OK**: Successfully edited the user.

```json
{
  "ok": true,
  "data": "Update successful"
}
```

- **400 Bad Request**: Invalid request parameters.

```json
{
  "ok": false,
  "error": "Invalid request body."
}
```

- **500 Internal Server Error**: An error occurred while editing the user.

```json
{
  "ok": false,
  "error": "Error occurred during editing."
}
```