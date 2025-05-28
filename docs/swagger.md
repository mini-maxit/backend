


# Mini-Maxit API
  

## Informations

### Version

1.0.0

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
| POST | /api/v1./auth/refresh | [post auth refresh](#post-auth-refresh) | Refresh JWT tokens |
| POST | /api/v1./login | [post login](#post-login) | Login a user |
| POST | /api/v1./register | [post register](#post-register) | Register a user |
  


###  group

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/v1./group/{id}/users | [delete group ID users](#delete-group-id-users) | Delete users from a group |
| GET | /api/v1./group | [get group](#get-group) | Get all groups |
| GET | /api/v1./group/{id} | [get group ID](#get-group-id) | Get a group |
| GET | /api/v1./group/{id}/users | [get group ID users](#get-group-id-users) | Get users in a group |
| POST | /api/v1./group | [post group](#post-group) | Create a group |
| POST | /api/v1./group/{id}/users | [post group ID users](#post-group-id-users) | Add users to a group |
| PUT | /api/v1./group/{id} | [put group ID](#put-group-id) | Edit a group |
  


###  submission

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1./submission | [get submission](#get-submission) | Get all submissions for the current user |
| GET | /api/v1./submission/group/{id} | [get submission group ID](#get-submission-group-id) | Get all submissions for a group |
| GET | /api/v1./submission/{id} | [get submission ID](#get-submission-id) | Get a submission by ID |
| GET | /api/v1./submission/languages | [get submission languages](#get-submission-languages) | Get all available languages |
| GET | /api/v1./submission/task/{id} | [get submission task ID](#get-submission-task-id) | Get all submissions for a task |
| GET | /api/v1./submission/user/{id} | [get submission user ID](#get-submission-user-id) | Get all submissions for a user |
| GET | /api/v1./submission/user/{id}/short | [get submission user ID short](#get-submission-user-id-short) | Get all submissions for a user |
  


###  task

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/v1./task/{id} | [delete task ID](#delete-task-id) | Delete a task |
| DELETE | /api/v1./task/{id}/unassign/groups | [delete task ID unassign groups](#delete-task-id-unassign-groups) | Unassign a task from groups |
| DELETE | /api/v1./task/{id}/unassign/users | [delete task ID unassign users](#delete-task-id-unassign-users) | Unassign a task from users |
| GET | /api/v1./task | [get task](#get-task) | Get all tasks |
| GET | /api/v1./task/group/{id} | [get task group ID](#get-task-group-id) | Get all tasks for a group |
| GET | /api/v1./task/{id} | [get task ID](#get-task-id) | Get a task |
| GET | /api/v1./task/{id}/limits | [get task ID limits](#get-task-id-limits) | Gets task limits |
| PATCH | /api/v1./task/{id} | [patch task ID](#patch-task-id) | Update a task |
| POST | /api/v1./task | [post task](#post-task) | Upload a task |
| POST | /api/v1./task/{id}/assign/groups | [post task ID assign groups](#post-task-id-assign-groups) | Assign a task to groups |
| POST | /api/v1./task/{id}/assign/users | [post task ID assign users](#post-task-id-assign-users) | Assign a task to users |
| PUT | /api/v1./task/{id}/limits | [put task ID limits](#put-task-id-limits) | Updates task limits |
  


###  user

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1./user | [get user](#get-user) | Get all users |
| GET | /api/v1./user/email | [get user email](#get-user-email) | Get user by email |
| GET | /api/v1./user/{id} | [get user ID](#get-user-id) | Get user by ID |
| PATCH | /api/v1./user/{id} | [patch user ID](#patch-user-id) | Edit user |
| PATCH | /api/v1./user/{id}/password | [patch user ID password](#patch-user-id-password) | Change user password |
  


###  worker

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1./worker/status | [get worker status](#get-worker-status) | Get worker status |
  


## Paths

### <span id="delete-group-id-users"></span> Delete users from a group (*DeleteGroupIDUsers*)

```
DELETE /api/v1./group/{id}/users
```

Delete users from a group

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Group ID |
| body | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserIDs](#github-com-mini-maxit-backend-package-domain-schemas-user-i-ds) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserIDs` | | ✓ | | User IDs |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-group-id-users-200) | OK | OK |  | [schema](#delete-group-id-users-200-schema) |
| [400](#delete-group-id-users-400) | Bad Request | Bad Request |  | [schema](#delete-group-id-users-400-schema) |
| [403](#delete-group-id-users-403) | Forbidden | Forbidden |  | [schema](#delete-group-id-users-403-schema) |
| [405](#delete-group-id-users-405) | Method Not Allowed | Method Not Allowed |  | [schema](#delete-group-id-users-405-schema) |
| [500](#delete-group-id-users-500) | Internal Server Error | Internal Server Error |  | [schema](#delete-group-id-users-500-schema) |

#### Responses


##### <span id="delete-group-id-users-200"></span> 200 - OK
Status: OK

###### <span id="delete-group-id-users-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="delete-group-id-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-group-id-users-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-group-id-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-group-id-users-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-group-id-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-group-id-users-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-group-id-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-group-id-users-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="delete-task-id"></span> Delete a task (*DeleteTaskID*)

```
DELETE /api/v1./task/{id}
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="delete-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="delete-task-id-unassign-groups"></span> Unassign a task from groups (*DeleteTaskIDUnassignGroups*)

```
DELETE /api/v1./task/{id}/unassign/groups
```

Unassigns a task from groups by task ID and group IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| groupIDs | `body` | []integer | `[]int64` | | ✓ | | Group IDs |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="delete-task-id-unassign-groups-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-unassign-groups-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-groups-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-unassign-groups-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-groups-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-unassign-groups-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-groups-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-unassign-groups-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="delete-task-id-unassign-users"></span> Unassign a task from users (*DeleteTaskIDUnassignUsers*)

```
DELETE /api/v1./task/{id}/unassign/users
```

Unassigns a task from users by task ID and user IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| userIDs | `body` | []integer | `[]int64` | | ✓ | | User IDs |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="delete-task-id-unassign-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="delete-task-id-unassign-users-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="delete-task-id-unassign-users-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="delete-task-id-unassign-users-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="delete-task-id-unassign-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="delete-task-id-unassign-users-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-group"></span> Get all groups (*GetGroup*)

```
GET /api/v1./group
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasGroup](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-group)

##### <span id="get-group-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-group-id"></span> Get a group (*GetGroupID*)

```
GET /api/v1./group/{id}
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasGroup](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-group)

##### <span id="get-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-group-id-users"></span> Get users in a group (*GetGroupIDUsers*)

```
GET /api/v1./group/{id}/users
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="get-group-id-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-group-id-users-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-group-id-users-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-id-users-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-group-id-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-id-users-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission"></span> Get all submissions for the current user (*GetSubmission*)

```
GET /api/v1./submission
```

Depending on the user role, this endpoint will return all submissions for the current user.

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-group-id"></span> Get all submissions for a group (*GetSubmissionGroupID*)

```
GET /api/v1./submission/group/{id}
```

If the user is a student, it fails with 403 Forbidden.

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
| [200](#get-submission-group-id-200) | OK | OK |  | [schema](#get-submission-group-id-200-schema) |
| [400](#get-submission-group-id-400) | Bad Request | Bad Request |  | [schema](#get-submission-group-id-400-schema) |
| [403](#get-submission-group-id-403) | Forbidden | Forbidden |  | [schema](#get-submission-group-id-403-schema) |
| [500](#get-submission-group-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-group-id-500-schema) |

#### Responses


##### <span id="get-submission-group-id-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-group-id-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-group-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-group-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-group-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-id"></span> Get a submission by ID (*GetSubmissionID*)

```
GET /api/v1./submission/{id}
```

If the user is a student, the submission must belong to the user.

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-languages"></span> Get all available languages (*GetSubmissionLanguages*)

```
GET /api/v1./submission/languages
```

Get all available languages for submitting solutions.

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasLanguageConfig](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-language-config)

##### <span id="get-submission-languages-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-languages-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-task-id"></span> Get all submissions for a task (*GetSubmissionTaskID*)

```
GET /api/v1./submission/task/{id}
```

If the user is a student and has no access to this task, it fails with 403 Forbidden.

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-task-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-task-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-task-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-user-id"></span> Get all submissions for a user (*GetSubmissionUserID*)

```
GET /api/v1./submission/user/{id}
```

If the user is a student, it fails with 403 Forbidden.

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
| [200](#get-submission-user-id-200) | OK | OK |  | [schema](#get-submission-user-id-200-schema) |
| [400](#get-submission-user-id-400) | Bad Request | Bad Request |  | [schema](#get-submission-user-id-400-schema) |
| [403](#get-submission-user-id-403) | Forbidden | Forbidden |  | [schema](#get-submission-user-id-403-schema) |
| [500](#get-submission-user-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-submission-user-id-500-schema) |

#### Responses


##### <span id="get-submission-user-id-200"></span> 200 - OK
Status: OK

###### <span id="get-submission-user-id-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-user-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-user-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-user-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-user-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-submission-user-id-short"></span> Get all submissions for a user (*GetSubmissionUserIDShort*)

```
GET /api/v1./submission/user/{id}/short
```

If the user is a student, it fails with 403 Forbidden.

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission)

##### <span id="get-submission-user-id-short-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-submission-user-id-short-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-user-id-short-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-submission-user-id-short-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-submission-user-id-short-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-submission-user-id-short-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-task"></span> Get all tasks (*GetTask*)

```
GET /api/v1./task
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasTask](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-task)

##### <span id="get-task-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-task-group-id"></span> Get all tasks for a group (*GetTaskGroupID*)

```
GET /api/v1./task/group/{id}
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseArrayGithubComMiniMaxitBackendPackageDomainSchemasTask](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-task)

##### <span id="get-task-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-task-group-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-task-group-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-task-group-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-group-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-task-id"></span> Get a task (*GetTaskID*)

```
GET /api/v1./task/{id}
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasTaskDetailed](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-task-detailed)

##### <span id="get-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-task-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="get-task-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-task-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-task-id-limits"></span> Gets task limits (*GetTaskIDLimits*)

```
GET /api/v1./task/{id}/limits
```

Gets task limits by task ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-task-id-limits-200) | OK | OK |  | [schema](#get-task-id-limits-200-schema) |
| [400](#get-task-id-limits-400) | Bad Request | Bad Request |  | [schema](#get-task-id-limits-400-schema) |
| [404](#get-task-id-limits-404) | Not Found | Not Found |  | [schema](#get-task-id-limits-404-schema) |
| [500](#get-task-id-limits-500) | Internal Server Error | Internal Server Error |  | [schema](#get-task-id-limits-500-schema) |

#### Responses


##### <span id="get-task-id-limits-200"></span> 200 - OK
Status: OK

###### <span id="get-task-id-limits-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="get-task-id-limits-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-task-id-limits-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-id-limits-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-task-id-limits-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-task-id-limits-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-id-limits-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-user"></span> Get all users (*GetUser*)

```
GET /api/v1./user
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasUser](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-user)

##### <span id="get-user-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-user-email"></span> Get user by email (*GetUserEmail*)

```
GET /api/v1./user/email
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasUser](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-user)

##### <span id="get-user-email-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-user-email-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-email-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-user-email-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-email-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-email-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-email-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-email-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-user-id"></span> Get user by ID (*GetUserID*)

```
GET /api/v1./user/{id}
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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasUser](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-user)

##### <span id="get-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-user-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-user-id-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-user-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-user-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="get-worker-status"></span> Get worker status (*GetWorkerStatus*)

```
GET /api/v1./worker/status
```

Returns the current status of all worker nodes

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-worker-status-200) | OK | OK |  | [schema](#get-worker-status-200-schema) |
| [401](#get-worker-status-401) | Unauthorized | Not authorized - requires teacher or admin role |  | [schema](#get-worker-status-401-schema) |
| [500](#get-worker-status-500) | Internal Server Error | Internal server error |  | [schema](#get-worker-status-500-schema) |
| [504](#get-worker-status-504) | Gateway Timeout | Gateway timeout - worker status request timed out |  | [schema](#get-worker-status-504-schema) |

#### Responses


##### <span id="get-worker-status-200"></span> 200 - OK
Status: OK

###### <span id="get-worker-status-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasWorkerStatus](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-worker-status)

##### <span id="get-worker-status-401"></span> 401 - Not authorized - requires teacher or admin role
Status: Unauthorized

###### <span id="get-worker-status-401-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-worker-status-500"></span> 500 - Internal server error
Status: Internal Server Error

###### <span id="get-worker-status-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="get-worker-status-504"></span> 504 - Gateway timeout - worker status request timed out
Status: Gateway Timeout

###### <span id="get-worker-status-504-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="patch-task-id"></span> Update a task (*PatchTaskID*)

```
PATCH /api/v1./task/{id}
```

Updates a task by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| archive | `formData` | file | `io.ReadCloser` |  |  |  | New archive for the task |
| title | `formData` | string | `string` |  |  |  | New title for the task |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="patch-task-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="patch-task-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-task-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="patch-task-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="patch-task-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="patch-task-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="patch-user-id"></span> Edit user (*PatchUserID*)

```
PATCH /api/v1./user/{id}
```

Edit user

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | User ID |
| request | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserEdit](#github-com-mini-maxit-backend-package-domain-schemas-user-edit) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserEdit` | | ✓ | | User Edit Request |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="patch-user-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="patch-user-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="patch-user-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="patch-user-id-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="patch-user-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="patch-user-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="patch-user-id-password"></span> Change user password (*PatchUserIDPassword*)

```
PATCH /api/v1./user/{id}/password
```

Change user password

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | User ID |
| request | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserChangePassword](#github-com-mini-maxit-backend-package-domain-schemas-user-change-password) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserChangePassword` | | ✓ | | User Change Password Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-user-id-password-200) | OK | OK |  | [schema](#patch-user-id-password-200-schema) |
| [400](#patch-user-id-password-400) | Bad Request | Bad Request |  | [schema](#patch-user-id-password-400-schema) |
| [403](#patch-user-id-password-403) | Forbidden | Forbidden |  | [schema](#patch-user-id-password-403-schema) |
| [404](#patch-user-id-password-404) | Not Found | Not Found |  | [schema](#patch-user-id-password-404-schema) |
| [500](#patch-user-id-password-500) | Internal Server Error | Internal Server Error |  | [schema](#patch-user-id-password-500-schema) |

#### Responses


##### <span id="patch-user-id-password-200"></span> 200 - OK
Status: OK

###### <span id="patch-user-id-password-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="patch-user-id-password-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="patch-user-id-password-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-password-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="patch-user-id-password-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-password-404"></span> 404 - Not Found
Status: Not Found

###### <span id="patch-user-id-password-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="patch-user-id-password-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="patch-user-id-password-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-auth-refresh"></span> Refresh JWT tokens (*PostAuthRefresh*)

```
POST /api/v1./auth/refresh
```

Refreshes JWT tokens using a valid refresh token

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| request | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasRefreshTokenRequest](#github-com-mini-maxit-backend-package-domain-schemas-refresh-token-request) | `models.GithubComMiniMaxitBackendPackageDomainSchemasRefreshTokenRequest` | | ✓ | | Refresh Token Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-auth-refresh-200) | OK | OK |  | [schema](#post-auth-refresh-200-schema) |
| [400](#post-auth-refresh-400) | Bad Request | Bad Request |  | [schema](#post-auth-refresh-400-schema) |
| [401](#post-auth-refresh-401) | Unauthorized | Unauthorized |  | [schema](#post-auth-refresh-401-schema) |
| [405](#post-auth-refresh-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-auth-refresh-405-schema) |
| [500](#post-auth-refresh-500) | Internal Server Error | Internal Server Error |  | [schema](#post-auth-refresh-500-schema) |

#### Responses


##### <span id="post-auth-refresh-200"></span> 200 - OK
Status: OK

###### <span id="post-auth-refresh-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasJWTTokens](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens)

##### <span id="post-auth-refresh-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-auth-refresh-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-auth-refresh-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="post-auth-refresh-401-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-auth-refresh-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-auth-refresh-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-auth-refresh-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-auth-refresh-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-group"></span> Create a group (*PostGroup*)

```
POST /api/v1./group
```

Create a group

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasCreateGroup](#github-com-mini-maxit-backend-package-domain-schemas-create-group) | `models.GithubComMiniMaxitBackendPackageDomainSchemasCreateGroup` | | ✓ | | Create Group |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseInt64](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-int64)

##### <span id="post-group-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-group-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-group-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-group-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-group-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-group-id-users"></span> Add users to a group (*PostGroupIDUsers*)

```
POST /api/v1./group/{id}/users
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
| body | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserIDs](#github-com-mini-maxit-backend-package-domain-schemas-user-i-ds) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserIDs` | | ✓ | | User IDs |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="post-group-id-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-group-id-users-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-id-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-group-id-users-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-id-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-group-id-users-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-group-id-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-group-id-users-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-login"></span> Login a user (*PostLogin*)

```
POST /api/v1./login
```

Logs in a user with email and password, returns JWT tokens

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| request | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserLoginRequest](#github-com-mini-maxit-backend-package-domain-schemas-user-login-request) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserLoginRequest` | | ✓ | | User Login Request |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasJWTTokens](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens)

##### <span id="post-login-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-login-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-login-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="post-login-401-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-login-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-login-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-register"></span> Register a user (*PostRegister*)

```
POST /api/v1./register
```

Registers a user with name, surname, email, username and password, returns JWT tokens

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| request | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasUserRegisterRequest](#github-com-mini-maxit-backend-package-domain-schemas-user-register-request) | `models.GithubComMiniMaxitBackendPackageDomainSchemasUserRegisterRequest` | | ✓ | | User Register Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#post-register-201) | Created | Created |  | [schema](#post-register-201-schema) |
| [400](#post-register-400) | Bad Request | Bad Request |  | [schema](#post-register-400-schema) |
| [405](#post-register-405) | Method Not Allowed | Method Not Allowed |  | [schema](#post-register-405-schema) |
| [409](#post-register-409) | Conflict | Conflict |  | [schema](#post-register-409-schema) |
| [500](#post-register-500) | Internal Server Error | Internal Server Error |  | [schema](#post-register-500-schema) |

#### Responses


##### <span id="post-register-201"></span> 201 - Created
Status: Created

###### <span id="post-register-201-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasJWTTokens](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens)

##### <span id="post-register-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-register-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-register-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-register-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-register-409"></span> 409 - Conflict
Status: Conflict

###### <span id="post-register-409-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-register-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-register-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-task"></span> Upload a task (*PostTask*)

```
POST /api/v1./task
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
| title | `formData` | string | `string` |  | ✓ |  | Name of the task |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasTaskCreateResponse](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-task-create-response)

##### <span id="post-task-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-task-id-assign-groups"></span> Assign a task to groups (*PostTaskIDAssignGroups*)

```
POST /api/v1./task/{id}/assign/groups
```

Assigns a task to groups by task ID and group IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| groupIDs | `body` | []integer | `[]int64` | | ✓ | | Group IDs |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="post-task-id-assign-groups-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-id-assign-groups-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-groups-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-task-id-assign-groups-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-groups-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-id-assign-groups-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-groups-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-id-assign-groups-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="post-task-id-assign-users"></span> Assign a task to users (*PostTaskIDAssignUsers*)

```
POST /api/v1./task/{id}/assign/users
```

Assigns a task to users by task ID and user IDs

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| userIDs | `body` | []integer | `[]int64` | | ✓ | | User IDs |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="post-task-id-assign-users-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="post-task-id-assign-users-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-users-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="post-task-id-assign-users-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-users-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-task-id-assign-users-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="post-task-id-assign-users-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-task-id-assign-users-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="put-group-id"></span> Edit a group (*PutGroupID*)

```
PUT /api/v1./group/{id}
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
| body | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasEditGroup](#github-com-mini-maxit-backend-package-domain-schemas-edit-group) | `models.GithubComMiniMaxitBackendPackageDomainSchemasEditGroup` | | ✓ | | Edit Group |

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
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseGithubComMiniMaxitBackendPackageDomainSchemasGroup](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-group)

##### <span id="put-group-id-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="put-group-id-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="put-group-id-403"></span> 403 - Forbidden
Status: Forbidden

###### <span id="put-group-id-403-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="put-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="put-group-id-405-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="put-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="put-group-id-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

### <span id="put-task-id-limits"></span> Updates task limits (*PutTaskIDLimits*)

```
PUT /api/v1./task/{id}/limits
```

Updates task limits by task ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | integer | `int64` |  | ✓ |  | Task ID |
| limits | `body` | [GithubComMiniMaxitBackendPackageDomainSchemasPutInputOutputRequest](#github-com-mini-maxit-backend-package-domain-schemas-put-input-output-request) | `models.GithubComMiniMaxitBackendPackageDomainSchemasPutInputOutputRequest` | | ✓ | | Task limits |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#put-task-id-limits-200) | OK | OK |  | [schema](#put-task-id-limits-200-schema) |
| [400](#put-task-id-limits-400) | Bad Request | Bad Request |  | [schema](#put-task-id-limits-400-schema) |
| [404](#put-task-id-limits-404) | Not Found | Not Found |  | [schema](#put-task-id-limits-404-schema) |
| [500](#put-task-id-limits-500) | Internal Server Error | Internal Server Error |  | [schema](#put-task-id-limits-500-schema) |

#### Responses


##### <span id="put-task-id-limits-200"></span> 200 - OK
Status: OK

###### <span id="put-task-id-limits-200-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIResponseString](#github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string)

##### <span id="put-task-id-limits-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="put-task-id-limits-400-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="put-task-id-limits-404"></span> 404 - Not Found
Status: Not Found

###### <span id="put-task-id-limits-404-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

##### <span id="put-task-id-limits-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="put-task-id-limits-500-schema"></span> Schema
   
  

[GithubComMiniMaxitBackendInternalAPIHTTPHttputilsAPIError](#github-com-mini-maxit-backend-internal-api-http-httputils-api-error)

## Models

### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-error"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIError


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendInternalAPIHTTPHttputilsErrorStruct](#github-com-mini-maxit-backend-internal-api-http-httputils-error-struct)| `GithubComMiniMaxitBackendInternalAPIHTTPHttputilsErrorStruct` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-group"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-array_github_com_mini-maxit_backend_package_domain_schemas_Group


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][GithubComMiniMaxitBackendPackageDomainSchemasGroup](#github-com-mini-maxit-backend-package-domain-schemas-group)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasGroup` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-language-config"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-array_github_com_mini-maxit_backend_package_domain_schemas_LanguageConfig


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][GithubComMiniMaxitBackendPackageDomainSchemasLanguageConfig](#github-com-mini-maxit-backend-package-domain-schemas-language-config)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasLanguageConfig` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-submission"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-array_github_com_mini-maxit_backend_package_domain_schemas_Submission


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][GithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-package-domain-schemas-submission)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasSubmission` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-array-github-com-mini-maxit-backend-package-domain-schemas-task"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-array_github_com_mini-maxit_backend_package_domain_schemas_Task


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][GithubComMiniMaxitBackendPackageDomainSchemasTask](#github-com-mini-maxit-backend-package-domain-schemas-task)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasTask` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-group"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_Group


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasGroup](#github-com-mini-maxit-backend-package-domain-schemas-group)| `GithubComMiniMaxitBackendPackageDomainSchemasGroup` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_JWTTokens


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasJWTTokens](#github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens)| `GithubComMiniMaxitBackendPackageDomainSchemasJWTTokens` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-submission"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_Submission


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasSubmission](#github-com-mini-maxit-backend-package-domain-schemas-submission)| `GithubComMiniMaxitBackendPackageDomainSchemasSubmission` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-submit-response"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_SubmitResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasSubmitResponse](#github-com-mini-maxit-backend-package-domain-schemas-submit-response)| `GithubComMiniMaxitBackendPackageDomainSchemasSubmitResponse` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-task-create-response"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_TaskCreateResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasTaskCreateResponse](#github-com-mini-maxit-backend-package-domain-schemas-task-create-response)| `GithubComMiniMaxitBackendPackageDomainSchemasTaskCreateResponse` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-task-detailed"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_TaskDetailed


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasTaskDetailed](#github-com-mini-maxit-backend-package-domain-schemas-task-detailed)| `GithubComMiniMaxitBackendPackageDomainSchemasTaskDetailed` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-user"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_User


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasUser](#github-com-mini-maxit-backend-package-domain-schemas-user)| `GithubComMiniMaxitBackendPackageDomainSchemasUser` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-github-com-mini-maxit-backend-package-domain-schemas-worker-status"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-github_com_mini-maxit_backend_package_domain_schemas_WorkerStatus


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [GithubComMiniMaxitBackendPackageDomainSchemasWorkerStatus](#github-com-mini-maxit-backend-package-domain-schemas-worker-status)| `GithubComMiniMaxitBackendPackageDomainSchemasWorkerStatus` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-int64"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-int64


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | integer| `int64` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-api-response-string"></span> github_com_mini-maxit_backend_internal_api_http_httputils.APIResponse-string


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | string| `string` |  | |  |  |
| ok | boolean| `bool` |  | |  |  |



### <span id="github-com-mini-maxit-backend-internal-api-http-httputils-error-struct"></span> github_com_mini-maxit_backend_internal_api_http_httputils.errorStruct


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | string| `string` |  | |  |  |
| message | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-create-group"></span> github_com_mini-maxit_backend_package_domain_schemas.CreateGroup


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| name | string| `string` | ✓ | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-edit-group"></span> github_com_mini-maxit_backend_package_domain_schemas.EditGroup


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| name | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-group"></span> github_com_mini-maxit_backend_package_domain_schemas.Group


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| createdAt | string| `string` |  | |  |  |
| createdBy | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| updatedAt | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-j-w-t-tokens"></span> github_com_mini-maxit_backend_package_domain_schemas.JWTTokens


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| accessToken | string| `string` |  | |  |  |
| expiresAt | string| `string` |  | |  |  |
| refreshToken | string| `string` |  | |  |  |
| tokenType | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-language-config"></span> github_com_mini-maxit_backend_package_domain_schemas.LanguageConfig


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| fileExtension | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| language | string| `string` |  | |  |  |
| version | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-put-input-output"></span> github_com_mini-maxit_backend_package_domain_schemas.PutInputOutput


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| memoryLimit | integer| `int64` |  | |  |  |
| order | integer| `int64` |  | |  |  |
| timeLimit | integer| `int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-put-input-output-request"></span> github_com_mini-maxit_backend_package_domain_schemas.PutInputOutputRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| limits | [][GithubComMiniMaxitBackendPackageDomainSchemasPutInputOutput](#github-com-mini-maxit-backend-package-domain-schemas-put-input-output)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasPutInputOutput` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-refresh-token-request"></span> github_com_mini-maxit_backend_package_domain_schemas.RefreshTokenRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| refreshToken | string| `string` | ✓ | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-submission"></span> github_com_mini-maxit_backend_package_domain_schemas.Submission


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| checkedAt | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| language | [GithubComMiniMaxitBackendPackageDomainSchemasLanguageConfig](#github-com-mini-maxit-backend-package-domain-schemas-language-config)| `GithubComMiniMaxitBackendPackageDomainSchemasLanguageConfig` |  | |  |  |
| languageId | integer| `int64` |  | |  |  |
| order | integer| `int64` |  | |  |  |
| result | [GithubComMiniMaxitBackendPackageDomainSchemasSubmissionResult](#github-com-mini-maxit-backend-package-domain-schemas-submission-result)| `GithubComMiniMaxitBackendPackageDomainSchemasSubmissionResult` |  | |  |  |
| status | string| `string` |  | |  |  |
| statusMessage | string| `string` |  | |  |  |
| submittedAt | string| `string` |  | |  |  |
| task | [GithubComMiniMaxitBackendPackageDomainSchemasTask](#github-com-mini-maxit-backend-package-domain-schemas-task)| `GithubComMiniMaxitBackendPackageDomainSchemasTask` |  | |  |  |
| taskId | integer| `int64` |  | |  |  |
| user | [GithubComMiniMaxitBackendPackageDomainSchemasUser](#github-com-mini-maxit-backend-package-domain-schemas-user)| `GithubComMiniMaxitBackendPackageDomainSchemasUser` |  | |  |  |
| userId | integer| `int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-submission-result"></span> github_com_mini-maxit_backend_package_domain_schemas.SubmissionResult


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | string| `string` |  | |  |  |
| createdAt | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| message | string| `string` |  | |  |  |
| submissionId | integer| `int64` |  | |  |  |
| testResults | [][GithubComMiniMaxitBackendPackageDomainSchemasTestResult](#github-com-mini-maxit-backend-package-domain-schemas-test-result)| `[]*GithubComMiniMaxitBackendPackageDomainSchemasTestResult` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-submit-response"></span> github_com_mini-maxit_backend_package_domain_schemas.SubmitResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| message | string| `string` |  | |  |  |
| submissionNumber | integer| `int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-task"></span> github_com_mini-maxit_backend_package_domain_schemas.Task


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| createdAt | string| `string` |  | |  |  |
| createdBy | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |
| updatedAt | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-task-create-response"></span> github_com_mini-maxit_backend_package_domain_schemas.TaskCreateResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | integer| `int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-task-detailed"></span> github_com_mini-maxit_backend_package_domain_schemas.TaskDetailed


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| createdAt | string| `string` |  | |  |  |
| createdBy | integer| `int64` |  | |  |  |
| createdByName | string| `string` |  | |  |  |
| descriptionUrl | string| `string` |  | |  |  |
| groupIds | []integer| `[]int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-test-result"></span> github_com_mini-maxit_backend_package_domain_schemas.TestResult


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | integer| `int64` |  | |  |  |
| inputOutputId | integer| `int64` |  | |  |  |
| passed | boolean| `bool` |  | |  |  |
| submissionResultId | integer| `int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user"></span> github_com_mini-maxit_backend_package_domain_schemas.User


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| role | [TypesUserRole](#types-user-role)| `TypesUserRole` |  | |  |  |
| surname | string| `string` |  | |  |  |
| username | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user-change-password"></span> github_com_mini-maxit_backend_package_domain_schemas.UserChangePassword


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| newPassword | string| `string` | ✓ | |  |  |
| newPasswordConfirm | string| `string` | ✓ | |  |  |
| oldPassword | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user-edit"></span> github_com_mini-maxit_backend_package_domain_schemas.UserEdit


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |
| role | [TypesUserRole](#types-user-role)| `TypesUserRole` |  | |  |  |
| surname | string| `string` |  | |  |  |
| username | string| `string` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user-i-ds"></span> github_com_mini-maxit_backend_package_domain_schemas.UserIDs


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| userIDs | []integer| `[]int64` |  | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user-login-request"></span> github_com_mini-maxit_backend_package_domain_schemas.UserLoginRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-user-register-request"></span> github_com_mini-maxit_backend_package_domain_schemas.UserRegisterRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| confirmPassword | string| `string` | ✓ | |  |  |
| email | string| `string` | ✓ | |  |  |
| name | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |
| surname | string| `string` | ✓ | |  |  |
| username | string| `string` | ✓ | |  |  |



### <span id="github-com-mini-maxit-backend-package-domain-schemas-worker-status"></span> github_com_mini-maxit_backend_package_domain_schemas.WorkerStatus


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| busyWorkers | integer| `int64` |  | |  |  |
| statusTime | string| `string` |  | |  |  |
| totalWorkers | integer| `int64` |  | |  |  |
| workerStatus | map of string| `map[string]string` |  | |  |  |



### <span id="types-user-role"></span> types.UserRole


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| types.UserRole | string| string | |  |  |


