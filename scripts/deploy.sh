#!/bin/bash

# Cloud Run デプロイスクリプト - GitHub Actions ラッパー
set -e

ENVIRONMENT=${1:-dev}

# GitHub CLIがインストールされているかチェック
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) が必要です。以下でインストールしてください："
    echo "  macOS: brew install gh"
    echo "  Ubuntu: sudo apt install gh"
    echo "  詳細: https://cli.github.com/"
    exit 1
fi

# GitHub認証チェック
if ! gh auth status &> /dev/null; then
    echo "❌ GitHub認証が必要です。以下のコマンドを実行してください："
    echo "  gh auth login"
    exit 1
fi

echo "🚀 Starting deployment via GitHub Actions"
echo "📋 Environment: $ENVIRONMENT"

if [ "$ENVIRONMENT" = "dev" ]; then
    echo "🧪 Triggering DEV environment deployment..."
    gh workflow run deploy-dev.yml
    echo "✅ DEV deployment workflow triggered!"
    echo "📊 進捗確認: gh run watch"
    echo "🔗 GitHub Actions: https://github.com/$(gh repo view --json owner,name -q '.owner.login + "/" + .name')/actions"
    
elif [ "$ENVIRONMENT" = "prod" ]; then
    echo "🚨 本番環境へのデプロイを開始します"
    echo "⚠️  この操作は本番環境に影響します"
    echo ""
    read -p "本当に本番環境にデプロイしますか？ 'deploy' と入力してください: " confirm
    
    if [ "$confirm" = "deploy" ]; then
        echo "🚀 Triggering PRODUCTION deployment..."
        gh workflow run deploy-prod.yml -f confirm=deploy
        echo "✅ Production deployment workflow triggered!"
        echo "📊 進捗確認: gh run watch"
        echo "🔗 GitHub Actions: https://github.com/$(gh repo view --json owner,name -q '.owner.login + "/" + .name')/actions"
    else
        echo "❌ デプロイをキャンセルしました"
        echo "   ('deploy' と正確に入力する必要があります)"
        exit 1
    fi
    
else
    echo "❌ Error: Invalid environment '$ENVIRONMENT'"
    echo "Usage: $0 [dev|prod]"
    echo ""
    echo "Examples:"
    echo "  $0 dev   # DEV環境にデプロイ"
    echo "  $0 prod  # 本番環境にデプロイ（確認付き）"
    exit 1
fi

echo ""
echo "💡 Tips:"
echo "  - ワークフロー一覧: gh run list"
echo "  - ログ確認: gh run view --log"
echo "  - リアルタイム監視: gh run watch"