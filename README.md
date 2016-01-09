# kyouen-server

Recreate from https://github.com/noboru-i/kyouen-python

## How to create prmd doc

```
bundle exec prmd combine --meta meta.json schemata/ | bundle exec prmd verify | bundle exec prmd doc --prepend overview.md > schema.md
```
