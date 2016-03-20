# How to create infra

## create S3 bucket

```
aws s3 mb s3://kyouen-terraform-state
```

setting remote config

```
terraform remote config -backend=S3 -backend-config="bucket=kyouen-terraform-state" -backend-config="key=terraform.tfstate"
```
