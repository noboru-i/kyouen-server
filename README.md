# kyouen-server

Recreate from https://github.com/noboru-i/kyouen-python

## How to create prmd doc

```
bundle exec prmd combine --meta docs/src/meta.json docs/src/schemata/ | bundle exec prmd verify | bundle exec prmd doc --prepend docs/src/overview.md > docs/schema.md
```
