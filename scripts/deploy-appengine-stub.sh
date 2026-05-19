#!/bin/bash

# App Engine スタブ デプロイスクリプト
# 旧 App Engine エンドポイントへのリクエストを https://kyouen.app/ に 301 リダイレクトするためのもの。
# 原則一度デプロイすれば更新不要。
set -e

ENVIRONMENT=${1:-dev}
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

case "$ENVIRONMENT" in
  dev)
    PROJECT_ID="api-project-732262258565"
    ;;
  prod)
    PROJECT_ID="my-android-server"
    ;;
  *)
    echo "Usage: $0 [dev|prod]"
    exit 1
    ;;
esac

echo "Deploying App Engine stub to $ENVIRONMENT (project: $PROJECT_ID) ..."
gcloud app deploy "$REPO_ROOT/appengine/app.yaml" --project="$PROJECT_ID" --quiet
echo "Done."
