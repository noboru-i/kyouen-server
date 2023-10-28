# 共円 in Google App Engine(go)

## local development

```sh
$ dev_appserver.py app.yaml --datastore_path=`pwd`/database/db.datastore -A my-android-server --support_datastore_emulator True --enable_host_checking=false
```

link to server
http://localhost:8080/

link to admin console
http://localhost:8000/

## setup env

For dev.

```sh
gcloud config set project api-project-732262258565
```

## deploy to production

```sh
$ gcloud app deploy --no-promote
```

### deploy dispatch.yaml

```sh
$ gcloud app deploy dispatch.yaml
```

## OpenAPI(Swagger)

### show Swagger UI

```sh
$ docker run -p 10000:8080 -v $(pwd)/docs:/usr/share/nginx/html/docs -e API_URL=http://localhost:10000/docs/specs/index.yaml swaggerapi/swagger-ui
```

### generate struct for go

#### golang

```sh
$ openapi-generator generate -i docs/specs/index.yaml -g go-server -o ./tmp
$ cp tmp/go/model_*.go openapi
$ rm -rf tmp
```

#### Android client

```sh
$ openapi-generator generate -i docs/specs/index.yaml -g kotlin -o ./tmp --additional-properties="packageName=hm.orz.chaos114.android.tumekyouen.network"
$ cp -r tmp/src/main/kotlin/hm/orz/chaos114/android/tumekyouen/network/models ../kyouen-android/app/src/main/java/hm/orz/chaos114/android/tumekyouen/network
$ rm -rf tmp
```
