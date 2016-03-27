# How to create infra

## create S3 bucket for manage `tfstate`

```
aws s3 mb s3://kyouen-terraform-state
```

setting remote config

```
terraform remote config -backend=S3 -backend-config="bucket=kyouen-terraform-state" -backend-config="key=terraform.tfstate"
```

push `tfstate`

```
terraform remote push
```

## Generate graph

```
terraform graph | dot -Tpng > graph.png
```

You may need `brew install graphviz`.
