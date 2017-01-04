## README

This README would normally document whatever steps are necessary to get the
application up and running.

Things you may want to cover:

* Ruby version

* System dependencies

* Configuration

* Database creation

* Database initialization

* How to run the test suite

* Services (job queues, cache servers, search engines, etc.)

* Deployment instructions

* ...

## How to create prmd doc

```
prmd combine --meta docs/src/meta.json docs/src/schemata/ | prmd verify | prmd doc --prepend docs/src/overview.md > docs/schema.md
```

## Run local

```
heroku local
```

access to http://localhost:5000/
