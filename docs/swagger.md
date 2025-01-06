


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
| POST | /api/v1/group | [post group](#post-group) | Create a group |
| PUT | /api/v1/group/{id} | [put group ID](#put-group-id) | Edit a group |
  


###  session

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/session/validate | [get session validate](#get-session-validate) | Validate a session |
| POST | /api/v1/session/invalidate | [post session invalidate](#post-session-invalidate) | Invalidate a session |
  


###  task

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/task | [get task](#get-task) | Get all tasks |
| GET | /api/v1/task/{id} | [get task ID](#get-task-id) | Get a task |
| POST | /api/v1/task | [post task](#post-task) | Upload a task |
  


## Paths

### <span id="get-group"></span> Get all groups (*GetGroup*)

```
GET /api/v1/group
```

Get all groups

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| limit | `query` | string | `string` |  |  |  | Limit |
| offset | `query` | string | `string` |  |  |  | Offset |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-group-200) | OK | OK |  | [schema](#get-group-200-schema) |
| [400](#get-group-400) | Bad Request | Bad Request |  | [schema](#get-group-400-schema) |
| [401](#get-group-401) | Unauthorized | Unauthorized |  | [schema](#get-group-401-schema) |
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

##### <span id="get-group-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="get-group-401-schema"></span> Schema
   
  

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
| [401](#get-group-id-401) | Unauthorized | Unauthorized |  | [schema](#get-group-id-401-schema) |
| [404](#get-group-id-404) | Not Found | Not Found |  | [schema](#get-group-id-404-schema) |
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

##### <span id="get-group-id-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="get-group-id-401-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="get-group-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-group-id-404-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="get-group-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-group-id-405-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="get-group-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-group-id-500-schema"></span> Schema
   
  

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

##### <span id="get-task-id-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="get-task-id-405-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="get-task-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-task-id-500-schema"></span> Schema
   
  

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
| [401](#post-group-401) | Unauthorized | Unauthorized |  | [schema](#post-group-401-schema) |
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

##### <span id="post-group-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="post-group-401-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="post-group-405"></span> 405 - Method Not Allowed
Status: Method Not Allowed

###### <span id="post-group-405-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="post-group-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-group-500-schema"></span> Schema
   
  

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
| [401](#put-group-id-401) | Unauthorized | Unauthorized |  | [schema](#put-group-id-401-schema) |
| [404](#put-group-id-404) | Not Found | Not Found |  | [schema](#put-group-id-404-schema) |
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

##### <span id="put-group-id-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="put-group-id-401-schema"></span> Schema
   
  

[APIError](#api-error)

##### <span id="put-group-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="put-group-id-404-schema"></span> Schema
   
  

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



### <span id="api-response-array-group"></span> ApiResponse-array_Group


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [][Group](#group)| `[]*Group` |  | |  |  |
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



### <span id="group"></span> Group


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | string| `string` |  | |  |  |
| created_by | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| updated_at | string| `string` |  | |  |  |



### <span id="session"></span> Session


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| expires_at | string| `string` |  | |  |  |
| session | string| `string` |  | |  |  |
| user_id | integer| `int64` |  | |  |  |
| user_role | string| `string` |  | |  |  |



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


