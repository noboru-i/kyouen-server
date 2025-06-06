#!/bin/bash

# Cloud Run デプロイスクリプト
set -e

PROJECT_ID="my-android-server"
SERVICE_NAME="kyouen-server"
REGION="asia-northeast1"

echo "🚀 Starting Cloud Run deployment for $SERVICE_NAME"

# プロジェクトIDが設定されているか確認
if [ -z "$PROJECT_ID" ]; then
    echo "❌ Error: PROJECT_ID is not set"
    exit 1
fi

# gcloudプロジェクトを設定
echo "📋 Setting gcloud project to $PROJECT_ID"
gcloud config set project $PROJECT_ID

# Dockerイメージをビルド
echo "🔨 Building Docker image"
docker build -t gcr.io/$PROJECT_ID/$SERVICE_NAME:latest .

# Container Registryにプッシュ
echo "📤 Pushing image to Container Registry"
docker push gcr.io/$PROJECT_ID/$SERVICE_NAME:latest

# Cloud Runにデプロイ
echo "🌐 Deploying to Cloud Run"
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME:latest \
  --region $REGION \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --set-env-vars GOOGLE_CLOUD_PROJECT=$PROJECT_ID \
  --memory 512Mi \
  --cpu 1 \
  --max-instances 10

echo "✅ Deployment completed successfully!"
echo ""
echo "🔗 Service URL:"
gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)"