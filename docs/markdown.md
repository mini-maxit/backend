


# Mini Maxit API Documentation testing the workflow
This is the API documentation for the Mini Maxit API.


## Informations

### Version

1.0

### Contact



## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json
  * multipart/form-data

### Produces
  * application/json

## All endpoints

###  auth

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/login | [post login](#post-login) | Login a user |
| POST | /api/v1/register | [post register](#post-register) | Register a user |



###  group

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/group | [get group](#get-group) | Get all groups |
| GET | /api/v1/group/{id} | [get group ID](#get-group-id) | Get a group |
| GET | /api/v1/group/{id}/users | [get group ID users](#get-group-id-users) | Get users in a group |
| POST | /api/v1/group | [post group](#post-group) | Create a group |
| POST | /api/v1/group/{id}/users | [post group ID users](#post-group-id-users) | Add users to a group |
| PUT | /api/v1/group/{id} | [put group ID](#put-group-id) | Edit a group |



###  session

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/session/validate | [get session validate](#get-session-validate) | Validate a session |
| POST | /api/v1/session/invalidate | [post session invalidate](#post-session-invalidate) | Invalidate a session |



###  submission

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/submission | [get submission](#get-submission) | Get all submissions for the current user |
| GET | /api/v1/submission/{id} | [get submission ID](#get-submission-id) | Get a submission by ID |
| GET | /api/v1/submission/languages | [get submission languages](#get-submission-languages) | Get all available languages |
| GET | /api/v1/submission/task/{id} | [get submission task ID](#get-submission-task-id) | Get all submissions for a task |
| GET | /api/v1/submission/user/{id} | [get submission user ID](#get-submission-user-id) | Get all submissions for a group |
| GET | /api/v1/submission/user/{id}/short | [get submission user ID short](#get-submission-user-id-short) | Get all submissions for a user |



###  task

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/v1/task/{id} | [delete task ID](#delete-task-id) | Delete a task |
| DELETE | /api/v1/task/{id}/unassign/groups | [delete task ID unassign groups](#delete-task-id-unassign-groups) | Unassign a task from groups |
| DELETE | /api/v1/task/{id}/unassign/users | [delete task ID unassign users](#delete-task-id-unassign-users) | Unassign a task from users |
| GET | /api/v1/task | [get task](#get-task) | Get all tasks |
| GET | /api/v1/task/group/{id} | [get task group ID](#get-task-group-id) | Get all tasks for a group |
| GET | /api/v1/task/{id} | [get task ID](#get-task-id) | Get a task |
| PATCH | /api/v1/task/{id} | [patch task ID](#patch-task-id) | Update a task |
| POST | /api/v1/task | [post task](#post-task) | Upload a task |
| POST | /api/v1/task/{id}/assign/groups | [post task ID assign groups](#post-task-id-assign-groups) | Assign a task to groups |
| POST | /api/v1/task/{id}/assign/users | [post task ID assign users](#post-task-id-assign-users) | Assign a task to users |



###  user

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/user | [get user](#get-user) | Get all users |
| GET | /api/v1/user/email | [get user email](#get-user-email) | Get user by email |
| GET | /api/v1/user/{id} | [get user ID](#get-user-id) | Get user by ID |
| PATCH | /api/v1/user/{id} | [patch user ID](#patch-user-id) | Edit user |



## Paths

### <span id="delete-task-id"></span> Delete a task (*DeleteTaskID*)

```
DELETE /api/v1/task/{id}
```

Deletes a task by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-task-id-200) | OK | OK |  | [schema](#delete-task-id-200-schema) |
| [400](#delete-task-id-400) | Bad Request | Bad Request |  | [schema](#delete-task-id-400-schema) |
| [403](#delete-task-id-403) | Forbidden | Forbidden |  | [schema](#delete-task-id-403-schema) |
| [405](#delete-task-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#delete-task-id-405-schema) |
| [500](#delete-task-id-500) | Internal Server Error | Internal Server Error |  | [schema](#delete-task-id-500-schema) |

#### Responses


##### <span id="delete-task-id-200"></span> 200 - OK
Status: OK

###### <span id="delete-task-id-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="delete-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="delete-task-id-unassign-groups"></span> Unassign a task from groups (*DeleteTaskIDUnassignGroups*)

```
DELETE /api/v1/task/{id}/unassign/groups
```

Unassigns a task from groups by task ID and group IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| groupIds | `body` | []integer | `[]int64` | | ✓ | | Group IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-task-id-unassign-groups-200) | OK | OK |  | [schema](#delete-task-id-unassign-groups-200-schema) |
| [400](#delete-task-id-unassign-groups-400) | Bad Request | Bad Request |  | [schema](#delete-task-id-unassign-groups-400-schema) |
| [403](#delete-task-id-unassign-groups-403) | Forbidden | Forbidden |  | [schema](#delete-task-id-unassign-groups-403-schema) |
| [405](#delete-task-id-unassign-groups-405) | Method Not Allowed | Method Not Allowed |  | [schema](#delete-task-id-unassign-groups-405-schema) |
| [500](#delete-task-id-unassign-groups-500) | Internal Server Error | Internal Server Error |  | [schema](#delete-task-id-unassign-groups-500-schema) |

#### Responses


##### <span id="delete-task-id-unassign-groups-200"></span> 200 - OK
Status: OK

###### <span id="delete-task-id-unassign-groups-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="delete-task-id-unassign-groups-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-unassign-groups-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-groups-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-unassign-groups-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-groups-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-unassign-groups-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-groups-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-unassign-groups-500-schema"></span> Schema



[APIError](#api-error)

### <span id="delete-task-id-unassign-users"></span> Unassign a task from users (*DeleteTaskIDUnassignUsers*)

```
DELETE /api/v1/task/{id}/unassign/users
```

Unassigns a task from users by task ID and user IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| userIds | `body` | []integer | `[]int64` | | ✓ | | User IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-task-id-unassign-users-200) | OK | OK |  | [schema](#delete-task-id-unassign-users-200-schema) |
| [400](#delete-task-id-unassign-users-400) | Bad Request | Bad Request |  | [schema](#delete-task-id-unassign-users-400-schema) |
| [403](#delete-task-id-unassign-users-403) | Forbidden | Forbidden |  | [schema](#delete-task-id-unassign-users-403-schema) |
| [405](#delete-task-id-unassign-users-405) | Method Not Allowed | Method Not Allowed |  | [schema](#delete-task-id-unassign-users-405-schema) |
| [500](#delete-task-id-unassign-users-500) | Internal Server Error | Internal Server Error |  | [schema](#delete-task-id-unassign-users-500-schema) |

#### Responses


##### <span id="delete-task-id-unassign-users-200"></span> 200 - OK
Status: OK

###### <span id="delete-task-id-unassign-users-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="delete-task-id-unassign-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-unassign-users-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-unassign-users-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-unassign-users-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="delete-task-id-unassign-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-unassign-users-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-group"></span> Get all groups (*GetGroup*)

```
GET /api/v1/group
```

Get all groups

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-group-200) | OK | OK |  | [schema](#get-group-200-schema) |
| [400](#get-group-400) | Bad Request | Bad Request |  | [schema](#get-group-400-schema) |
| [403](#get-group-403) | Forbidden | Forbidden |  | [schema](#get-group-403-schema) |
| [405](#get-group-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-group-405-schema) |
| [500](#get-group-500) | Internal Server Error | Internal Server Error |  | [schema](#get-group-500-schema) |

#### Responses


##### <span id="get-group-200"></span> 200 - OK
Status: OK

###### <span id="get-group-200-schema"></span> Schema



[APIResponseArrayGroup](#api-response-array-group)

##### <span id="get-group-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-group-id"></span> Get a group (*GetGroupID*)

```
GET /api/v1/group/{id}
```

Get a group

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-group-id-200) | OK | OK |  | [schema](#get-group-id-200-schema) |
| [400](#get-group-id-400) | Bad Request | Bad Request |  | [schema](#get-group-id-400-schema) |
| [403](#get-group-id-403) | Forbidden | Forbidden |  | [schema](#get-group-id-403-schema) |
| [405](#get-group-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-group-id-405-schema) |
| [500](#get-group-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-group-id-500-schema) |

#### Responses


##### <span id="get-group-id-200"></span> 200 - OK
Status: OK

###### <span id="get-group-id-200-schema"></span> Schema



[APIResponseGroup](#api-response-group)

##### <span id="get-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-group-id-users"></span> Get users in a group (*GetGroupIDUsers*)

```
GET /api/v1/group/{id}/users
```

Get users in a group

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-group-id-users-200) | OK | OK |  | [schema](#get-group-id-users-200-schema) |
| [400](#get-group-id-users-400) | Bad Request | Bad Request |  | [schema](#get-group-id-users-400-schema) |
| [403](#get-group-id-users-403) | Forbidden | Forbidden |  | [schema](#get-group-id-users-403-schema) |
| [405](#get-group-id-users-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-group-id-users-405-schema) |
| [500](#get-group-id-users-500) | Internal Server Error | Internal Server Error |  | [schema](#get-group-id-users-500-schema) |

#### Responses


##### <span id="get-group-id-users-200"></span> 200 - OK
Status: OK

###### <span id="get-group-id-users-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="get-group-id-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-id-users-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-id-users-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-id-users-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-group-id-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-id-users-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-session-validate"></span> Validate a session (*GetSessionValidate*)

```
GET /api/v1/session/validate
```

Validates a session token

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Session | `header` | string | `string` |  | ✓ |  | Session Token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-session-validate-200) | OK | OK |  | [schema](#get-session-validate-200-schema) |
| [400](#get-session-validate-400) | Bad Request | Bad Request |  | [schema](#get-session-validate-400-schema) |
| [401](#get-session-validate-401) | Unauthorized | Unauthorized |  | [schema](#get-session-validate-401-schema) |
| [500](#get-session-validate-500) | Internal Server Error | Internal Server Error |  | [schema](#get-session-validate-500-schema) |

#### Responses


##### <span id="get-session-validate-200"></span> 200 - OK
Status: OK

###### <span id="get-session-validate-200-schema"></span> Schema



[APIResponseArrayTask](#api-response-array-task)

##### <span id="get-session-validate-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-session-validate-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-session-validate-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="get-session-validate-401-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-session-validate-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-session-validate-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission"></span> Get all submissions for the current user (*GetSubmission*)

```
GET /api/v1/submission
```

Depending on the user role, this endpoint will return all submissions for the current user if user is student, all submissions to owned tasks if user is teacher, and all submissions in database if user is admin

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Session | `header` | string | `string` |  | ✓ |  | Session Token |
| limit | `query` | integer | `int64` |  |  |  | Limit the number of returned submissions |
| offset | `query` | integer | `int64` |  |  |  | Offset the returned submissions |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-200) | OK | OK |  | [schema](#get-submission-200-schema) |
| [400](#get-submission-400) | Bad Request | Bad Request |  | [schema](#get-submission-400-schema) |
| [500](#get-submission-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-500-schema) |

#### Responses


##### <span id="get-submission-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-200-schema"></span> Schema



[APIResponseArraySubmission](#api-response-array-submission)

##### <span id="get-submission-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission-id"></span> Get a submission by ID (*GetSubmissionID*)

```
GET /api/v1/submission/{id}
```

Get a submission by its ID, if the user is a student, the submission must belong to the user, if the user is a teacher, the submission must belong to a task owned by the teacher, if the user is an admin, the submission can be any submission

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Submission ID |
| Session | `header` | string | `string` |  | ✓ |  | Session Token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-id-200) | OK | OK |  | [schema](#get-submission-id-200-schema) |
| [400](#get-submission-id-400) | Bad Request | Bad Request |  | [schema](#get-submission-id-400-schema) |
| [500](#get-submission-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-id-500-schema) |

#### Responses


##### <span id="get-submission-id-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-id-200-schema"></span> Schema



[APIResponseSubmission](#api-response-submission)

##### <span id="get-submission-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission-languages"></span> Get all available languages (*GetSubmissionLanguages*)

```
GET /api/v1/submission/languages
```

Get all available languages for submitting solutions. Temporary solution, while all tasks have same languages

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-languages-200) | OK | OK |  | [schema](#get-submission-languages-200-schema) |
| [500](#get-submission-languages-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-languages-500-schema) |

#### Responses


##### <span id="get-submission-languages-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-languages-200-schema"></span> Schema



[APIResponseArrayLanguageConfig](#api-response-array-language-config)

##### <span id="get-submission-languages-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-languages-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission-task-id"></span> Get all submissions for a task (*GetSubmissionTaskID*)

```
GET /api/v1/submission/task/{id}
```

Gets all submissions for specific task. If the user is a student and has no access to this task, it fails with 403 Forbidden. For teacher it returns all submissions for this task if he created it. For admin it returns all submissions for specific task.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| Session | `header` | string | `string` |  | ✓ |  | Session Token |
| limit | `query` | integer | `int64` |  |  |  | Limit the number of returned submissions |
| offset | `query` | integer | `int64` |  |  |  | Offset the returned submissions |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-task-id-200) | OK | OK |  | [schema](#get-submission-task-id-200-schema) |
| [400](#get-submission-task-id-400) | Bad Request | Bad Request |  | [schema](#get-submission-task-id-400-schema) |
| [403](#get-submission-task-id-403) | Forbidden | Forbidden |  | [schema](#get-submission-task-id-403-schema) |
| [500](#get-submission-task-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-task-id-500-schema) |

#### Responses


##### <span id="get-submission-task-id-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-task-id-200-schema"></span> Schema



[APIResponseArraySubmission](#api-response-array-submission)

##### <span id="get-submission-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-task-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-task-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-task-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission-user-id"></span> Get all submissions for a group (*GetSubmissionUserID*)

```
GET /api/v1/submission/user/{id}
```

Gets all submissions for specific group. If the user is a student, it fails with 403 Forbidden. For teacher it returns all submissions from this group for tasks he created. For admin it returns all submissions for specific group.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |
| Session | `header` | string | `string` |  | ✓ |  | Session Token |
| limit | `query` | integer | `int64` |  |  |  | Limit the number of returned submissions |
| offset | `query` | integer | `int64` |  |  |  | Offset the returned submissions |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-user-id-200) | OK | OK |  | [schema](#get-submission-user-id-200-schema) |
| [400](#get-submission-user-id-400) | Bad Request | Bad Request |  | [schema](#get-submission-user-id-400-schema) |
| [403](#get-submission-user-id-403) | Forbidden | Forbidden |  | [schema](#get-submission-user-id-403-schema) |
| [500](#get-submission-user-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-user-id-500-schema) |

#### Responses


##### <span id="get-submission-user-id-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-user-id-200-schema"></span> Schema



[APIResponseArraySubmission](#api-response-array-submission)

##### <span id="get-submission-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-user-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-user-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-user-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-user-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-submission-user-id-short"></span> Get all submissions for a user (*GetSubmissionUserIDShort*)

```
GET /api/v1/submission/user/{id}/short
```

Gets all submissions for specific user. If the user is a student, it fails with 403 Forbidden. For teacher it returns all submissions from this user for tasks owned by the teacher. For admin it returns all submissions for specific user.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | User ID |
| Session | `header` | string | `string` |  | ✓ |  | Session Token |
| limit | `query` | integer | `int64` |  |  |  | Limit the number of returned submissions |
| offset | `query` | integer | `int64` |  |  |  | Offset the returned submissions |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-submission-user-id-short-200) | OK | OK |  | [schema](#get-submission-user-id-short-200-schema) |
| [400](#get-submission-user-id-short-400) | Bad Request | Bad Request |  | [schema](#get-submission-user-id-short-400-schema) |
| [403](#get-submission-user-id-short-403) | Forbidden | Forbidden |  | [schema](#get-submission-user-id-short-403-schema) |
| [500](#get-submission-user-id-short-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-user-id-short-500-schema) |

#### Responses


##### <span id="get-submission-user-id-short-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-user-id-short-200-schema"></span> Schema



[APIResponseArraySubmission](#api-response-array-submission)

##### <span id="get-submission-user-id-short-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-user-id-short-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-user-id-short-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-user-id-short-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-submission-user-id-short-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-user-id-short-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-task"></span> Get all tasks (*GetTask*)

```
GET /api/v1/task
```

Returns all tasks

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-task-200) | OK | OK |  | [schema](#get-task-200-schema) |
| [500](#get-task-500) | Internal Server Error | Internal Server Error |  | [schema](#get-task-500-schema) |

#### Responses


##### <span id="get-task-200"></span> 200 - OK
Status: OK

###### <span id="get-task-200-schema"></span> Schema



[APIResponseArrayTask](#api-response-array-task)

##### <span id="get-task-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-task-group-id"></span> Get all tasks for a group (*GetTaskGroupID*)

```
GET /api/v1/task/group/{id}
```

Returns all tasks for a group by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-task-group-id-200) | OK | OK |  | [schema](#get-task-group-id-200-schema) |
| [400](#get-task-group-id-400) | Bad Request | Bad Request |  | [schema](#get-task-group-id-400-schema) |
| [403](#get-task-group-id-403) | Forbidden | Forbidden |  | [schema](#get-task-group-id-403-schema) |
| [405](#get-task-group-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-task-group-id-405-schema) |
| [500](#get-task-group-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-task-group-id-500-schema) |

#### Responses


##### <span id="get-task-group-id-200"></span> 200 - OK
Status: OK

###### <span id="get-task-group-id-200-schema"></span> Schema



[APIResponseArrayTask](#api-response-array-task)

##### <span id="get-task-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-task-group-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-task-group-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-task-group-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-group-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-task-id"></span> Get a task (*GetTaskID*)

```
GET /api/v1/task/{id}
```

Returns a task by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-task-id-200) | OK | OK |  | [schema](#get-task-id-200-schema) |
| [400](#get-task-id-400) | Bad Request | Bad Request |  | [schema](#get-task-id-400-schema) |
| [403](#get-task-id-403) | Forbidden | Forbidden |  | [schema](#get-task-id-403-schema) |
| [405](#get-task-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-task-id-405-schema) |
| [500](#get-task-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-task-id-500-schema) |

#### Responses


##### <span id="get-task-id-200"></span> 200 - OK
Status: OK

###### <span id="get-task-id-200-schema"></span> Schema



[APIResponseTaskDetailed](#api-response-task-detailed)

##### <span id="get-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-task-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-task-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-task-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-user"></span> Get all users (*GetUser*)

```
GET /api/v1/user
```

Get all users

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| limit | `query` | integer | `int64` |  |  |  | Limit |
| offset | `query` | integer | `int64` |  |  |  | Offset |
| sort | `query` | string | `string` |  |  |  | Sort |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-200) | OK | OK |  | [schema](#get-user-200-schema) |
| [405](#get-user-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-user-405-schema) |
| [500](#get-user-500) | Internal Server Error | Internal Server Error |  | [schema](#get-user-500-schema) |

#### Responses


##### <span id="get-user-200"></span> 200 - OK
Status: OK

###### <span id="get-user-200-schema"></span> Schema



[APIResponseUser](#api-response-user)

##### <span id="get-user-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-user-email"></span> Get user by email (*GetUserEmail*)

```
GET /api/v1/user/email
```

Get user by email

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| email | `query` | string | `string` |  | ✓ |  | User email |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-email-200) | OK | OK |  | [schema](#get-user-email-200-schema) |
| [400](#get-user-email-400) | Bad Request | Bad Request |  | [schema](#get-user-email-400-schema) |
| [404](#get-user-email-404) | Not Found | Not Found |  | [schema](#get-user-email-404-schema) |
| [405](#get-user-email-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-user-email-405-schema) |
| [500](#get-user-email-500) | Internal Server Error | Internal Server Error |  | [schema](#get-user-email-500-schema) |

#### Responses


##### <span id="get-user-email-200"></span> 200 - OK
Status: OK

###### <span id="get-user-email-200-schema"></span> Schema



[APIResponseUser](#api-response-user)

##### <span id="get-user-email-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-user-email-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-email-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-user-email-404-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-email-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-email-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-email-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-email-500-schema"></span> Schema



[APIError](#api-error)

### <span id="get-user-id"></span> Get user by ID (*GetUserID*)

```
GET /api/v1/user/{id}
```

Get user by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | User ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-id-200) | OK | OK |  | [schema](#get-user-id-200-schema) |
| [400](#get-user-id-400) | Bad Request | Bad Request |  | [schema](#get-user-id-400-schema) |
| [404](#get-user-id-404) | Not Found | Not Found |  | [schema](#get-user-id-404-schema) |
| [405](#get-user-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#get-user-id-405-schema) |
| [500](#get-user-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-user-id-500-schema) |

#### Responses


##### <span id="get-user-id-200"></span> 200 - OK
Status: OK

###### <span id="get-user-id-200-schema"></span> Schema



[APIResponseUser](#api-response-user)

##### <span id="get-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-user-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-user-id-404-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="get-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="patch-task-id"></span> Update a task (*PatchTaskID*)

```
PATCH /api/v1/task/{id}
```

Updates a task by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| body | `body` | [EditTask](#edit-task) | `models.EditTask` | | ✓ | | Task object |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-task-id-200) | OK | OK |  | [schema](#patch-task-id-200-schema) |
| [400](#patch-task-id-400) | Bad Request | Bad Request |  | [schema](#patch-task-id-400-schema) |
| [403](#patch-task-id-403) | Forbidden | Forbidden |  | [schema](#patch-task-id-403-schema) |
| [405](#patch-task-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#patch-task-id-405-schema) |
| [500](#patch-task-id-500) | Internal Server Error | Internal Server Error |  | [schema](#patch-task-id-500-schema) |

#### Responses


##### <span id="patch-task-id-200"></span> 200 - OK
Status: OK

###### <span id="patch-task-id-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="patch-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="patch-task-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="patch-task-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="patch-task-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="patch-task-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="patch-user-id"></span> Edit user (*PatchUserID*)

```
PATCH /api/v1/user/{id}
```

Edit user

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | User ID |
| body | `body` | [UserEdit](#user-edit) | `models.UserEdit` | | ✓ | | User edit object |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-user-id-200) | OK | OK |  | [schema](#patch-user-id-200-schema) |
| [400](#patch-user-id-400) | Bad Request | Bad Request |  | [schema](#patch-user-id-400-schema) |
| [403](#patch-user-id-403) | Forbidden | Forbidden |  | [schema](#patch-user-id-403-schema) |
| [404](#patch-user-id-404) | Not Found | Not Found |  | [schema](#patch-user-id-404-schema) |
| [405](#patch-user-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#patch-user-id-405-schema) |
| [500](#patch-user-id-500) | Internal Server Error | Internal Server Error |  | [schema](#patch-user-id-500-schema) |

#### Responses


##### <span id="patch-user-id-200"></span> 200 - OK
Status: OK

###### <span id="patch-user-id-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="patch-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="patch-user-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-user-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="patch-user-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-user-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="patch-user-id-404-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-user-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="patch-user-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="patch-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="patch-user-id-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-group"></span> Create a group (*PostGroup*)

```
POST /api/v1/group
```

Create a group

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [CreateGroup](#create-group) | `models.CreateGroup` | | ✓ | | Create Group |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-group-200) | OK | OK |  | [schema](#post-group-200-schema) |
| [400](#post-group-400) | Bad Request | Bad Request |  | [schema](#post-group-400-schema) |
| [403](#post-group-403) | Forbidden | Forbidden |  | [schema](#post-group-403-schema) |
| [405](#post-group-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-group-405-schema) |
| [500](#post-group-500) | Internal Server Error | Internal Server Error |  | [schema](#post-group-500-schema) |

#### Responses


##### <span id="post-group-200"></span> 200 - OK
Status: OK

###### <span id="post-group-200-schema"></span> Schema



[APIResponseInt64](#api-response-int64)

##### <span id="post-group-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-group-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-group-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-group-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-group-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-group-id-users"></span> Add users to a group (*PostGroupIDUsers*)

```
POST /api/v1/group/{id}/users
```

Add users to a group

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |
| body | `body` | [UserIds](#user-ids) | `models.UserIds` | | ✓ | | User IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-group-id-users-200) | OK | OK |  | [schema](#post-group-id-users-200-schema) |
| [400](#post-group-id-users-400) | Bad Request | Bad Request |  | [schema](#post-group-id-users-400-schema) |
| [403](#post-group-id-users-403) | Forbidden | Forbidden |  | [schema](#post-group-id-users-403-schema) |
| [405](#post-group-id-users-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-group-id-users-405-schema) |
| [500](#post-group-id-users-500) | Internal Server Error | Internal Server Error |  | [schema](#post-group-id-users-500-schema) |

#### Responses


##### <span id="post-group-id-users-200"></span> 200 - OK
Status: OK

###### <span id="post-group-id-users-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="post-group-id-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-group-id-users-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-id-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-group-id-users-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-id-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-group-id-users-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-group-id-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-group-id-users-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-login"></span> Login a user (*PostLogin*)

```
POST /api/v1/login
```

Logs in a user with email and password

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| request | `body` | [UserLoginRequest](#user-login-request) | `models.UserLoginRequest` | | ✓ | | User Login Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-login-200) | OK | OK |  | [schema](#post-login-200-schema) |
| [400](#post-login-400) | Bad Request | Bad Request |  | [schema](#post-login-400-schema) |
| [401](#post-login-401) | Unauthorized | Unauthorized |  | [schema](#post-login-401-schema) |
| [500](#post-login-500) | Internal Server Error | Internal Server Error |  | [schema](#post-login-500-schema) |

#### Responses


##### <span id="post-login-200"></span> 200 - OK
Status: OK

###### <span id="post-login-200-schema"></span> Schema



[APIResponseSession](#api-response-session)

##### <span id="post-login-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-login-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-login-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="post-login-401-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-login-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-login-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-register"></span> Register a user (*PostRegister*)

```
POST /api/v1/register
```

Registers a user with name, surname, email, username and password

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| request | `body` | [UserRegisterRequest](#user-register-request) | `models.UserRegisterRequest` | | ✓ | | User Register Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#post-register-201) | Created | Created |  | [schema](#post-register-201-schema) |
| [400](#post-register-400) | Bad Request | Bad Request |  | [schema](#post-register-400-schema) |
| [405](#post-register-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-register-405-schema) |
| [500](#post-register-500) | Internal Server Error | Internal Server Error |  | [schema](#post-register-500-schema) |

#### Responses


##### <span id="post-register-201"></span> 201 - Created
Status: Created

###### <span id="post-register-201-schema"></span> Schema



[APIResponseSession](#api-response-session)

##### <span id="post-register-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-register-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-register-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-register-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-register-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-register-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-session-invalidate"></span> Invalidate a session (*PostSessionInvalidate*)

```
POST /api/v1/session/invalidate
```

Invalidates a session token

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Session | `header` | string | `string` |  | ✓ |  | Session Token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-session-invalidate-200) | OK | OK |  | [schema](#post-session-invalidate-200-schema) |
| [400](#post-session-invalidate-400) | Bad Request | Bad Request |  | [schema](#post-session-invalidate-400-schema) |
| [401](#post-session-invalidate-401) | Unauthorized | Unauthorized |  | [schema](#post-session-invalidate-401-schema) |
| [405](#post-session-invalidate-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-session-invalidate-405-schema) |
| [500](#post-session-invalidate-500) | Internal Server Error | Internal Server Error |  | [schema](#post-session-invalidate-500-schema) |

#### Responses


##### <span id="post-session-invalidate-200"></span> 200 - OK
Status: OK

###### <span id="post-session-invalidate-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="post-session-invalidate-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-session-invalidate-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-session-invalidate-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="post-session-invalidate-401-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-session-invalidate-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-session-invalidate-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-session-invalidate-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-session-invalidate-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-task"></span> Upload a task (*PostTask*)

```
POST /api/v1/task
```

Uploads a task to the FileStorage service

#### Consumes
  * multipart/form-data

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| archive | `formData` | file | `io.ReadCloser` |  | ✓ |  | Task archive |
| overwrite | `formData` | boolean | `bool` |  |  |  | Overwrite flag |
| taskName | `formData` | string | `string` |  | ✓ |  | Name of the task |
| userId | `formData` | integer | `int64` |  | ✓ |  | ID of the author |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-task-200) | OK | OK |  | [schema](#post-task-200-schema) |
| [400](#post-task-400) | Bad Request | Bad Request |  | [schema](#post-task-400-schema) |
| [405](#post-task-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-task-405-schema) |
| [500](#post-task-500) | Internal Server Error | Internal Server Error |  | [schema](#post-task-500-schema) |

#### Responses


##### <span id="post-task-200"></span> 200 - OK
Status: OK

###### <span id="post-task-200-schema"></span> Schema



[APIResponseTaskCreateResponse](#api-response-task-create-response)

##### <span id="post-task-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-task-id-assign-groups"></span> Assign a task to groups (*PostTaskIDAssignGroups*)

```
POST /api/v1/task/{id}/assign/groups
```

Assigns a task to groups by task ID and group IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| groupIds | `body` | []integer | `[]int64` | | ✓ | | Group IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-task-id-assign-groups-200) | OK | OK |  | [schema](#post-task-id-assign-groups-200-schema) |
| [400](#post-task-id-assign-groups-400) | Bad Request | Bad Request |  | [schema](#post-task-id-assign-groups-400-schema) |
| [403](#post-task-id-assign-groups-403) | Forbidden | Forbidden |  | [schema](#post-task-id-assign-groups-403-schema) |
| [405](#post-task-id-assign-groups-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-task-id-assign-groups-405-schema) |
| [500](#post-task-id-assign-groups-500) | Internal Server Error | Internal Server Error |  | [schema](#post-task-id-assign-groups-500-schema) |

#### Responses


##### <span id="post-task-id-assign-groups-200"></span> 200 - OK
Status: OK

###### <span id="post-task-id-assign-groups-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="post-task-id-assign-groups-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-id-assign-groups-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-groups-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-task-id-assign-groups-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-groups-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-id-assign-groups-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-groups-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-id-assign-groups-500-schema"></span> Schema



[APIError](#api-error)

### <span id="post-task-id-assign-users"></span> Assign a task to users (*PostTaskIDAssignUsers*)

```
POST /api/v1/task/{id}/assign/users
```

Assigns a task to users by task ID and user IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| userIds | `body` | []integer | `[]int64` | | ✓ | | User IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-task-id-assign-users-200) | OK | OK |  | [schema](#post-task-id-assign-users-200-schema) |
| [400](#post-task-id-assign-users-400) | Bad Request | Bad Request |  | [schema](#post-task-id-assign-users-400-schema) |
| [403](#post-task-id-assign-users-403) | Forbidden | Forbidden |  | [schema](#post-task-id-assign-users-403-schema) |
| [405](#post-task-id-assign-users-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-task-id-assign-users-405-schema) |
| [500](#post-task-id-assign-users-500) | Internal Server Error | Internal Server Error |  | [schema](#post-task-id-assign-users-500-schema) |

#### Responses


##### <span id="post-task-id-assign-users-200"></span> 200 - OK
Status: OK

###### <span id="post-task-id-assign-users-200-schema"></span> Schema



[APIResponseString](#api-response-string)

##### <span id="post-task-id-assign-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-id-assign-users-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-task-id-assign-users-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-id-assign-users-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="post-task-id-assign-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-id-assign-users-500-schema"></span> Schema



[APIError](#api-error)

### <span id="put-group-id"></span> Edit a group (*PutGroupID*)

```
PUT /api/v1/group/{id}
```

Edit a group

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |
| body | `body` | [EditGroup](#edit-group) | `models.EditGroup` | | ✓ | | Edit Group |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#put-group-id-200) | OK | OK |  | [schema](#put-group-id-200-schema) |
| [400](#put-group-id-400) | Bad Request | Bad Request |  | [schema](#put-group-id-400-schema) |
| [403](#put-group-id-403) | Forbidden | Forbidden |  | [schema](#put-group-id-403-schema) |
| [405](#put-group-id-405) | Method Not Allowed | Method Not Allowed |  | [schema](#put-group-id-405-schema) |
| [500](#put-group-id-500) | Internal Server Error | Internal Server Error |  | [schema](#put-group-id-500-schema) |

#### Responses


##### <span id="put-group-id-200"></span> 200 - OK
Status: OK

###### <span id="put-group-id-200-schema"></span> Schema



[APIResponseGroup](#api-response-group)

##### <span id="put-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="put-group-id-400-schema"></span> Schema



[APIError](#api-error)

##### <span id="put-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="put-group-id-403-schema"></span> Schema



[APIError](#api-error)

##### <span id="put-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="put-group-id-405-schema"></span> Schema



[APIError](#api-error)

##### <span id="put-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="put-group-id-500-schema"></span> Schema



[APIError](#api-error)

## Models

### <span id="api-error"></span> ApiError






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [ErrorStruct](#error-struct)| `ErrorStruct` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-group"></span> ApiResponse-Group






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [Group](#group)| `Group` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-session"></span> ApiResponse-Session






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [Session](#session)| `Session` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-submission"></span> ApiResponse-Submission






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [Submission](#submission)| `Submission` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-submit-response"></span> ApiResponse-SubmitResponse






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [SubmitResponse](#submit-response)| `SubmitResponse` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-task-create-response"></span> ApiResponse-TaskCreateResponse






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [TaskCreateResponse](#task-create-response)| `TaskCreateResponse` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-task-detailed"></span> ApiResponse-TaskDetailed






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [TaskDetailed](#task-detailed)| `TaskDetailed` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-user"></span> ApiResponse-User






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [User](#user)| `User` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-array-group"></span> ApiResponse-array_Group






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][Group](#group)| `[]*Group` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-array-language-config"></span> ApiResponse-array_LanguageConfig






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][LanguageConfig](#language-config)| `[]*LanguageConfig` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-array-submission"></span> ApiResponse-array_Submission






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][Submission](#submission)| `[]*Submission` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-array-task"></span> ApiResponse-array_Task






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][Task](#task)| `[]*Task` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-int64"></span> ApiResponse-int64






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | integer| `int64` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="api-response-string"></span> ApiResponse-string






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | string| `string` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="create-group"></span> CreateGroup






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| name | string| `string` | ✓ | |  |  |



### <span id="edit-group"></span> EditGroup






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| name | string| `string` |  | |  |  |



### <span id="edit-task"></span> EditTask






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| title | string| `string` |  | |  |  |



### <span id="group"></span> Group






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | string| `string` |  | |  |  |
| created_by | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| updated_at | string| `string` |  | |  |  |



### <span id="language-config"></span> LanguageConfig






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| file_extension | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| language | [ModelsLanguageType](#models-language-type)| `ModelsLanguageType` |  | |  |  |
| version | string| `string` |  | |  |  |



### <span id="session"></span> Session






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| expires_at | string| `string` |  | |  |  |
| session | string| `string` |  | |  |  |
| user_id | integer| `int64` |  | |  |  |
| user_role | string| `string` |  | |  |  |



### <span id="submission"></span> Submission






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| checked_at | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| language | [LanguageConfig](#language-config)| `LanguageConfig` |  | |  |  |
| language_id | integer| `int64` |  | |  |  |
| order | integer| `int64` |  | |  |  |
| result | [SubmissionResult](#submission-result)| `SubmissionResult` |  | |  |  |
| status | string| `string` |  | |  |  |
| status_message | string| `string` |  | |  |  |
| submitted_at | string| `string` |  | |  |  |
| task | [Task](#task)| `Task` |  | |  |  |
| task_id | integer| `int64` |  | |  |  |
| user | [User](#user)| `User` |  | |  |  |
| user_id | integer| `int64` |  | |  |  |



### <span id="submission-result"></span> SubmissionResult






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | string| `string` |  | |  |  |
| created_at | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| message | string| `string` |  | |  |  |
| submission_id | integer| `int64` |  | |  |  |
| test_results | [][TestResult](#test-result)| `[]*TestResult` |  | |  |  |



### <span id="submit-response"></span> SubmitResponse






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| message | string| `string` |  | |  |  |
| submissionNumber | integer| `int64` |  | |  |  |



### <span id="task"></span> Task






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | string| `string` |  | |  |  |
| created_by | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |



### <span id="task-create-response"></span> TaskCreateResponse






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | integer| `int64` |  | |  |  |



### <span id="task-detailed"></span> TaskDetailed






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | string| `string` |  | |  |  |
| created_by | integer| `int64` |  | |  |  |
| created_by_name | string| `string` |  | |  |  |
| description_url | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |



### <span id="test-result"></span> TestResult






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | integer| `int64` |  | |  |  |
| input_output_id | integer| `int64` |  | |  |  |
| passed | boolean| `bool` |  | |  |  |
| submission_result_id | integer| `int64` |  | |  |  |



### <span id="user"></span> User






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| role | [TypesUserRole](#types-user-role)| `TypesUserRole` |  | |  |  |
| surname | string| `string` |  | |  |  |
| username | string| `string` |  | |  |  |



### <span id="user-edit"></span> UserEdit






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |
| role | [TypesUserRole](#types-user-role)| `TypesUserRole` |  | |  |  |
| surname | string| `string` |  | |  |  |
| username | string| `string` |  | |  |  |



### <span id="user-ids"></span> UserIds






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| user_ids | []integer| `[]int64` |  | |  |  |



### <span id="user-login-request"></span> UserLoginRequest






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |



### <span id="user-register-request"></span> UserRegisterRequest






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` | ✓ | |  |  |
| name | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |
| surname | string| `string` | ✓ | |  |  |
| username | string| `string` | ✓ | |  |  |



### <span id="error-struct"></span> errorStruct






**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | string| `string` |  | |  |  |
| message | string| `string` |  | |  |  |



### <span id="models-language-type"></span> models.LanguageType




| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| models.LanguageType | string| string | |  |  |



### <span id="types-user-role"></span> types.UserRole




| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| types.UserRole | string| string | |  |  |
