# Firebase / GCP 設定比較：DEV vs PROD

DEVとPROD両環境の実際の設定をAPI経由で取得し、差異を整理した資料です。

> **取得日**: 2026-02-22
> **取得方法**: Firebase Admin API / Cloud Run API / Artifact Registry API（`gcloud auth print-access-token` 経由）

---

## 1. Firebase Authentication

### 1.1 基本設定

| 項目 | DEV | PROD | 差異 |
|------|-----|------|------|
| Project ID | `api-project-732262258565` | `my-android-server` | ✅ 異なる |
| API Key | `AIzaSyAh8p...` | `AIzaSyC505...` | ✅ 異なる |
| Firebase Subdomain | `api-project-732262258565` | `my-android-server` | ✅ 異なる |
| Default Hosting Site | `api-project-732262258565` | `my-android-server` | ✅ 異なる |
| Callback URI | `https://api-project-732262258565.firebaseapp.com/__/auth/action` | `https://my-android-server.firebaseapp.com/__/auth/action` | ✅ 異なる |
| MFA | DISABLED | DISABLED | ✓ 一致 |
| Email Privacy (Improved) | true | true | ✓ 一致 |

### 1.2 サインイン方法（Identity Providers）

| IdP | DEV | PROD | 差異 |
|-----|-----|------|------|
| 匿名認証 | **無効** | **有効** | ⚠️ 差異あり |
| Twitter (`twitter.com`) | **有効** (`clientId: 4oHTCBMSkiq...`) | **未設定** | ⚠️ 差異あり |
| Apple (`apple.com`) | **有効** (`clientId: hm.orz.chaos114...dev.sign`, teamId: `56W5SXE4HE`) | **未設定** | ⚠️ 差異あり |

> **注意**: DEVのApple IdPには `privateKey` が設定されています（Secret Managerへの移行を推奨）

### 1.3 許可ドメイン（Authorized Domains）

| ドメイン | DEV | PROD |
|----------|-----|------|
| `localhost` | ✅ | ✅ |
| `api-project-732262258565.firebaseapp.com` | ✅ | - |
| `api-project-732262258565.web.app` | ✅ | - |
| `my-android-server.firebaseapp.com` | - | ✅ |
| `my-android-server.web.app` | - | ✅ |

---

## 2. Firestore（Datastore モード）

| 項目 | DEV | PROD | 差異 |
|------|-----|------|------|
| データベース名 | `(default)` | `(default)` | ✓ 一致 |
| ロケーション | **`asia-northeast1`**（東京） | **`nam5`**（米国） | ⚠️ 差異あり |
| タイプ | `DATASTORE_MODE` | `DATASTORE_MODE` | ✓ 一致 |
| 並行処理モード | **`PESSIMISTIC`** | **`OPTIMISTIC`** | ⚠️ 差異あり |
| Point-in-Time Recovery | DISABLED | DISABLED | ✓ 一致 |
| Delete Protection | DISABLED | DISABLED | ✓ 一致 |
| App Engine Integration | ENABLED | ENABLED | ✓ 一致 |
| Free Tier | true | true | ✓ 一致 |
| Version Retention | 3600s | 3600s | ✓ 一致 |
| 作成日 | 2023-10-14 | 1970-01-01（旧Datastore） | ✓ 旧来の差異 |

> **重要**: PRODのFirestoreは `nam5`（米国）に配置されており、Cloud Runのリージョン（`asia-northeast1`）と異なります。ネットワーク遅延の懸念があります。

---

## 3. Cloud Run

| 項目 | DEV | PROD | 差異 |
|------|-----|------|------|
| サービス名 | `kyouen-server-dev` | `kyouen-server-prod` | ✅ 異なる |
| リージョン | `asia-northeast1` | `asia-northeast1` | ✓ 一致 |
| CPU | `1` | `1` | ✓ 一致 |
| メモリ | `512Mi` | `512Mi` | ✓ 一致 |
| 最大インスタンス数 | `10` | `10` | ✓ 一致 |
| 最小インスタンス数 | 未設定（0） | 未設定（0） | ✓ 一致 |
| 同時実行数 | `80` | `80` | ✓ 一致 |
| タイムアウト | `300s` | `300s` | ✓ 一致 |
| Ingress | `INGRESS_TRAFFIC_ALL` | `INGRESS_TRAFFIC_ALL` | ✓ 一致 |
| 認証 | 未認証許可 | 未認証許可 | ✓ 一致 |
| `GOOGLE_CLOUD_PROJECT` env | `api-project-732262258565` | `my-android-server` | ✅ 異なる |
| `ENVIRONMENT` env | `dev` | `prod` | ✅ 異なる |
| サービスアカウント | `732262258565-compute@developer.gserviceaccount.com` | `495839147593-compute@developer.gserviceaccount.com` | ✅ 異なる（デフォルトSA） |
| URL | `https://kyouen-server-dev-jrvczq7a7a-an.a.run.app` | `https://kyouen-server-prod-pmf3yplfzq-an.a.run.app` | ✅ 異なる |

---

## 4. Artifact Registry

| 項目 | DEV | PROD | 差異 |
|------|-----|------|------|
| リポジトリ名 | `kyouen-repo` | `kyouen-repo` | ✓ 一致 |
| フォーマット | `DOCKER` | `DOCKER` | ✓ 一致 |
| ロケーション | `asia-northeast1` | `asia-northeast1` | ✓ 一致 |
| モード | `STANDARD_REPOSITORY` | `STANDARD_REPOSITORY` | ✓ 一致 |
| 脆弱性スキャン | **`SCANNING_DISABLED`** | **`SCANNING_ACTIVE`** | ⚠️ 差異あり |
| サイズ | ~700MB | ~49MB | ✅ 異なる（DEVに多くのイメージ） |

---

## 5. 差異サマリーと改善提案

### 重要な差異

| # | 項目 | DEV | PROD | 優先度 |
|---|------|-----|------|--------|
| 1 | **Firestore ロケーション** | asia-northeast1 | nam5 | 🔴 高 |
| 2 | **Firestore 並行処理モード** | PESSIMISTIC | OPTIMISTIC | 🟡 中 |
| 3 | **匿名認証** | 無効 | 有効 | 🟡 中 |
| 4 | **Twitter/Apple IdP** | 有効 | 未設定 | 🔴 高 |
| 5 | **脆弱性スキャン** | 無効 | 有効 | 🟡 中 |

### 改善提案

1. **Firestore ロケーション**: PRODが`nam5`（米国）のため、Cloud Run（東京）との通信に遅延が発生している可能性があります。PRODのFirestoreを移行するか、Cloud Runをnam5に対応するリージョン（us-central1等）に移動することを検討してください。ただし、既存データの移行は困難なため、現状維持でもよいか要判断。

2. **Firebase Auth IdP**: DEVにはTwitter・Apple認証が設定されていますが、PRODには設定されていません。意図的であれば問題ありませんが、PRODへの展開が必要な場合はIaCで管理することを推奨します。

3. **脆弱性スキャン**: DEVでも有効化を推奨します（Container Analysis APIを有効にする必要があります）。

4. **サービスアカウント**: 現在はデフォルトのComputeサービスアカウントを使用しています。最小権限の原則に基づき、専用のサービスアカウントを作成することを推奨します（IaCで管理可能）。

---

## 6. 関連ファイル

- [`.firebaserc`](../.firebaserc) - Firebase プロジェクトエイリアス定義
- [`firebase.json`](../firebase.json) - Firebase設定（エミュレーター設定）
- [`terraform/`](../terraform/) - IaC（Terraform）定義
- [`docs/environments.md`](./environments.md) - 環境設定差異（デプロイ観点）
