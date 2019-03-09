# 共円 in Google App Engine(go)

## local development

```sh
$ dev_appserver.py app.yaml --datastore_path=`pwd`/database/db.datastore -A my-android-server
```

## deploy to production

```sh
$ gcloud app deploy --no-promote
```
