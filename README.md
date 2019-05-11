# 共円 in Google App Engine(go)

## local development

```sh
$ dev_appserver.py app.yaml --datastore_path=`pwd`/database/db.datastore -A my-android-server --support_datastore_emulator True
```

link to server
http://localhost:8080/

link to admin console
http://localhost:8000/

## deploy to production

```sh
$ gcloud app deploy --no-promote
```

## show swagger ui

```sh
$ docker run -p 10000:8080 -v $(pwd)/docs:/usr/share/nginx/html/docs -e API_URL=http://localhost:10000/docs/specs/index.yaml swaggerapi/swagger-ui
```
