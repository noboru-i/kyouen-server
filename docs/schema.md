## <a name="resource-kyouen"></a>Kyouen

kyouen stages

### Attributes

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **created_at** | *date-time* | when kyouen was created | `"2015-01-01T12:00:00Z"` |
| **creator** | *string* | creator name. | `"example"` |
| **id** | *integer* | unique identifier of kyouen. | `42` |
| **size** | *integer* | size of kyouen.<br/> **one of:**`6` or `9` | `6` |
| **stage** | *string* | stage of kyouen.<br/> **pattern:** <code>^[0&#124;1]*$</code> | `"000000010000001100001100000000001000"` |

### Kyouen Create

Create a new kyouen.

```
POST /kyouen
```

#### Optional Parameters

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **creator** | *string* | creator name. | `"example"` |
| **size** | *integer* | size of kyouen.<br/> **one of:**`6` or `9` | `6` |
| **stage** | *string* | stage of kyouen.<br/> **pattern:** <code>^[0&#124;1]*$</code> | `"000000010000001100001100000000001000"` |


#### Curl Example

```bash
$ curl -n -X POST https://api.hello.com/kyouen \
  -d '{
  "size": 6,
  "stage": "000000010000001100001100000000001000",
  "creator": "example"
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 201 Created
```

```json
{
  "id": 42,
  "size": 6,
  "stage": "000000010000001100001100000000001000",
  "creator": "example",
  "created_at": "2015-01-01T12:00:00Z"
}
```

### Kyouen List

List existing kyouen.

```
GET /kyouen
```

#### Optional Parameters

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **count** | *integer* | Specifies the number of kyouen.<br/> **Range:** `value <= 200` | `42` |
| **start_stage_no** | *integer* | Returns results with an ID greater than the specified ID. | `42` |


#### Curl Example

```bash
$ curl -n https://api.hello.com/kyouen
 -G \
  -d start_stage_no=42 \
  -d count=42
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
[
  {
    "id": 42,
    "size": 6,
    "stage": "000000010000001100001100000000001000",
    "creator": "example",
    "created_at": "2015-01-01T12:00:00Z"
  }
]
```

### Kyouen clear kyouen

Post cleared kyouen.

```
PUT /kyouen/{kyouen_id}/clear
```

#### Optional Parameters

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **stage** | *string* | check is kyouen<br/> **pattern:** <code>^[0&#124;1&#124;2]*$</code> | `"000000010000002200002200000000001000"` |


#### Curl Example

```bash
$ curl -n -X PUT https://api.hello.com/kyouen/$KYOUEN_ID/clear \
  -d '{
  "stage": "000000010000002200002200000000001000"
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "id": 42,
  "size": 6,
  "stage": "000000010000001100001100000000001000",
  "creator": "example",
  "created_at": "2015-01-01T12:00:00Z"
}
```


## <a name="resource-user"></a>User

logged in user

### Attributes

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **access_secret** | *string* | token secret of twitter. | `"example"` |
| **access_token** | *string* | token of twitter. | `"example"` |
| **clear_stage_count** | *integer* | count of cleared stage. | `42` |
| **id** | *integer* | unique identifier of user. | `42` |
| **image_path** | *string* | when user was created | `"http://my-android-server.appspot.com/image/icon.png"` |
| **screen_name** | *string* | screen name of twitter | `"twitter_name"` |

### User Login

Login user.

```
POST /user/login
```

#### Optional Parameters

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **access_secret** | *string* | token secret of twitter. | `"example"` |
| **access_token** | *string* | token of twitter. | `"example"` |


#### Curl Example

```bash
$ curl -n -X POST https://api.hello.com/user/login \
  -d '{
  "access_token": "example",
  "access_secret": "example"
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 201 Created
```

```json
{
  "id": 42,
  "screen_name": "twitter_name",
  "access_token": "example",
  "access_secret": "example",
  "clear_stage_count": 42,
  "image_path": "http://my-android-server.appspot.com/image/icon.png"
}
```

### User Sync

Sync user.

```
POST /user/sync
```

#### Optional Parameters

| Name | Type | Description | Example |
| ------- | ------- | ------- | ------- |
| **cleared/cleared_at** | *date-time* |  | `"2015-01-01T12:00:00Z"` |
| **cleared/id** | *integer* | unique identifier of kyouen. | `42` |


#### Curl Example

```bash
$ curl -n -X POST https://api.hello.com/user/sync \
  -d '{
  "cleared": [
    {
      "id": 42,
      "cleared_at": "2015-01-01T12:00:00Z"
    }
  ]
}' \
  -H "Content-Type: application/json"
```


#### Response Example

```
HTTP/1.1 200 OK
```

```json
{
  "cleared": [
    {
      "id": 42,
      "cleared_at": "2015-01-01T12:00:00Z"
    }
  ]
}
```


