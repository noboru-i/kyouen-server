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
bundle exec prmd combine --meta docs/src/meta.json docs/src/schemata/ | bundle exec prmd verify | bundle exec prmd doc --prepend docs/src/overview.md > docs/schema.md
```

## Run Vagrant

```
vagrant up
vagrant rsync-auto
vagrant ssh
cd /app
rails s -p 3000 -b '0.0.0.0'
```

## Run Docker

```
docker-compose build
docker-compose up
```
