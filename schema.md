# Overview

Sample of overview message.

## <a name="resource-app"></a>App

FIXME

### Attributes

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **created_at** | *date-time* | when app was created | `"2015-01-01T12:00:00Z"` |
| **description** | *string* | discription of app<br/> **Length:** `0..140` | `"example"` |
| **id** | *integer* | unique identifier of app | `42` |
| **name** | *string* | unique name of app<br/> **pattern:** <code>^(\([0-9]{3}\))?[0-9]{3}-[0-9]{4}$</code> | `"example"` |
| **[updated_at](#resource-user)** | *date-time* | when app was updated | `"2015-01-01T12:00:00Z"` |
| **[user:created_at](#resource-user)** | *date-time* | when user was created | `"2015-01-01T12:00:00Z"` |
| **[user:id](#resource-user)** | *uuid* | unique identifier of user | `"01234567-89ab-cdef-0123-456789abcdef"` |
| **[user:name](#resource-user)** | *string* | unique name of user | `"example"` |
| **user:updated_at** | *date-time* | when user was updated | `"2015-01-01T12:00:00Z"` |

### App Create

Create a new app.

```
POST /apps
```


#### Curl Example

```bash
$ curl -n -X POST https://api.hello.com/apps \
  -d '{
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 201 Created
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "description": "example",
  "id": 42,
  "name": "example",
  "user": {
    "created_at": "2015-01-01T12:00:00Z",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "name": "example",
    "updated_at": "2015-01-01T12:00:00Z"
  },
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### App Delete

Delete an existing app.

```
DELETE /apps/{app_id}
```


#### Curl Example

```bash
$ curl -n -X DELETE https://api.hello.com/apps/$APP_ID \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "description": "example",
  "id": 42,
  "name": "example",
  "user": {
    "created_at": "2015-01-01T12:00:00Z",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "name": "example",
    "updated_at": "2015-01-01T12:00:00Z"
  },
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### App Info

Info for existing app.

```
GET /apps/{app_id}
```


#### Curl Example

```bash
$ curl -n https://api.hello.com/apps/$APP_ID
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "description": "example",
  "id": 42,
  "name": "example",
  "user": {
    "created_at": "2015-01-01T12:00:00Z",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "name": "example",
    "updated_at": "2015-01-01T12:00:00Z"
  },
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### App List

List existing apps.

```
GET /apps
```


#### Curl Example

```bash
$ curl -n https://api.hello.com/apps
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "list": [
    {
      "created_at": "2015-01-01T12:00:00Z",
      "description": "example",
      "id": 42,
      "name": "example",
      "user": {
        "created_at": "2015-01-01T12:00:00Z",
        "id": "01234567-89ab-cdef-0123-456789abcdef",
        "name": "example",
        "updated_at": "2015-01-01T12:00:00Z"
      },
      "updated_at": "2015-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "page": 42
  }
}
```

### App Update

Update an existing app.

```
PATCH /apps/{app_id}
```


#### Curl Example

```bash
$ curl -n -X PATCH https://api.hello.com/apps/$APP_ID \
  -d '{
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "description": "example",
  "id": 42,
  "name": "example",
  "user": {
    "created_at": "2015-01-01T12:00:00Z",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "name": "example",
    "updated_at": "2015-01-01T12:00:00Z"
  },
  "updated_at": "2015-01-01T12:00:00Z"
}
```


## <a name="resource-user"></a>User

FIXME

### Attributes

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **created_at** | *date-time* | when user was created | `"2015-01-01T12:00:00Z"` |
| **id** | *uuid* | unique identifier of user | `"01234567-89ab-cdef-0123-456789abcdef"` |
| **name** | *string* | unique name of user | `"example"` |
| **updated_at** | *date-time* | when user was updated | `"2015-01-01T12:00:00Z"` |

### User Create

Create a new user.

```
POST /users
```


#### Curl Example

```bash
$ curl -n -X POST https://api.hello.com/users \
  -d '{
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 201 Created
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "example",
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### User Delete

Delete an existing user.

```
DELETE /users/{user_id_or_name}
```


#### Curl Example

```bash
$ curl -n -X DELETE https://api.hello.com/users/$USER_ID_OR_NAME \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "example",
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### User Info

Info for existing user.

```
GET /users/{user_id_or_name}
```


#### Curl Example

```bash
$ curl -n https://api.hello.com/users/$USER_ID_OR_NAME
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "example",
  "updated_at": "2015-01-01T12:00:00Z"
}
```

### User List

List existing users.

```
GET /users
```


#### Curl Example

```bash
$ curl -n https://api.hello.com/users
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
[
  {
    "created_at": "2015-01-01T12:00:00Z",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "name": "example",
    "updated_at": "2015-01-01T12:00:00Z"
  }
]
```

### User Update

Update an existing user.

```
PATCH /users/{user_id_or_name}
```


#### Curl Example

```bash
$ curl -n -X PATCH https://api.hello.com/users/$USER_ID_OR_NAME \
  -d '{
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "created_at": "2015-01-01T12:00:00Z",
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "example",
  "updated_at": "2015-01-01T12:00:00Z"
}
```


