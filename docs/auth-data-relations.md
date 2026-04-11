# 認証情報の関連図

Firebase Auth ベースの認証と Datastore の User エンティティのデータ項目の関係を示す図です。

## データ項目マッピング図

```mermaid
flowchart LR
    subgraph FirebaseAuth["Firebase Auth"]
        subgraph FABase["UserRecord"]
            FA_uid["uid\n(Firebase UID)"]
            FA_displayName["displayName"]
            FA_photoURL["photoURL"]
            FA_email["email"]
        end
        subgraph FAProvider["ProviderUserInfo\n(ProviderID: twitter.com)"]
            FP_uid["UID\n(Twitter UID)"]
            FP_display["DisplayName"]
            FP_photo["PhotoURL"]
        end
        subgraph FAClaims["Token Claims\n(firebase.identities)"]
            FC_twitter["twitter.com[0]\n(Twitter UID)"]
        end
    end

    subgraph Datastore["Datastore: User エンティティ"]
        subgraph DSKey["エンティティキー (Named Key)"]
            DS_key["'KEY' + Firebase UID"]
        end
        subgraph DSCurrent["フィールド"]
            DS_userId["userId\n(Firebase UID)"]
            DS_screenName["screenName\n(Twitter スクリーン名)"]
            DS_image["image\n(プロフィール画像 URL)"]
            DS_clearStageCount["clearStageCount\n(クリアステージ数)"]
            DS_twitterUid["twitterUid\n(Twitter UID)"]
        end
        subgraph DSLegacy["廃止予定フィールド (deprecated)"]
            DS_accessToken["accessToken ❌"]
            DS_accessSecret["accessSecret ❌"]
            DS_apiToken["apiToken ❌"]
        end
    end

    FA_uid -->|キー生成| DS_key
    FA_uid -->|set| DS_userId
    FA_displayName -->|優先| DS_screenName
    FP_display -. フォールバック .-> DS_screenName
    FA_photoURL -->|優先| DS_image
    FP_photo -. フォールバック .-> DS_image
    FC_twitter -->|優先| DS_twitterUid
    FP_uid -. フォールバック .-> DS_twitterUid
```

## 各フィールドの説明

### Firebase Auth → Datastore User マッピング

| Firebase Auth フィールド | Datastore User フィールド | 優先度 | 説明 |
|---|---|---|---|
| `UserRecord.uid` | `userId` | - | Firebase UID をそのままコピー |
| `UserRecord.uid` | キー (`KEY{uid}`) | - | エンティティキーの名前として使用 |
| `UserRecord.displayName` | `screenName` | 優先 | Twitter スクリーン名（Firebase Auth の DisplayName） |
| `ProviderUserInfo["twitter.com"].DisplayName` | `screenName` | フォールバック | displayName が空の場合に使用 |
| `UserRecord.photoURL` | `image` | 優先 | Twitter プロフィール画像 URL |
| `ProviderUserInfo["twitter.com"].PhotoURL` | `image` | フォールバック | photoURL が空の場合に使用 |
| `Token.Claims["firebase"]["identities"]["twitter.com"][0]` | `twitterUid` | 優先 | Twitter UID（トークンクレームから取得） |
| `ProviderUserInfo["twitter.com"].UID` | `twitterUid` | フォールバック | クレームから取得できない場合に使用 |

### アプリケーション管理フィールド

| フィールド | 説明 |
|---|---|
| `clearStageCount` | 認証とは無関係。ユーザーがステージをクリアするたびにアプリケーション側でインクリメント |

### 廃止予定フィールド (deprecated)

旧 Twitter OAuth 直接連携で使用されていたフィールド。Firebase Auth 移行後は不使用。

| フィールド | 旧用途 | 現状 |
|---|---|---|
| `accessToken` | Twitter OAuth アクセストークン | 廃止予定（TODO: 削除） |
| `accessSecret` | Twitter OAuth アクセスシークレット | 廃止予定（TODO: 削除） |
| `apiToken` | カスタム API トークン | 廃止予定（TODO: 削除） |

## ログインフロー概要

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant FirebaseAuth as Firebase Auth
    participant Datastore

    Client->>Server: POST /login<br/>{ "token": "<Firebase ID Token>" }
    Server->>FirebaseAuth: VerifyIDToken(idToken)
    FirebaseAuth-->>Server: auth.Token { uid, Claims }
    Server->>FirebaseAuth: GetUser(uid)
    FirebaseAuth-->>Server: UserRecord { displayName, photoURL, ProviderUserInfo }
    Server->>Server: screenName / image / twitterUID を抽出
    Server->>Datastore: CreateOrUpdateUserFromFirebase(uid, screenName, image, twitterUID)
    Datastore-->>Server: User エンティティ
    Server-->>Client: ユーザー情報レスポンス
```

## kyouen-python からのマイグレーションフロー

kyouen-python (旧Pythonアプリ) では、UserエンティティのキーとしてTwitter UIDを使用していた。
Firebase Auth 移行後は Firebase UID をキーとして使用するため、初回ログイン時に自動マイグレーションが実行される。

### キー構造の違い

| システム | キー形式 | 例 |
|---|---|---|
| kyouen-python (旧) | `'KEY' + Twitter UID` | `KEY12345` |
| kyouen-server (現) | `'KEY' + Firebase UID` | `KEYabc123def` |

### マイグレーションフロー

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Datastore

    Client->>Server: POST /login (初回ログイン)
    Server->>Datastore: GetUserByID(firebaseUID) → 見つからない
    Server->>Datastore: GetUserByID(twitterUID) → 旧エンティティ発見
    Note over Server,Datastore: MigrateLegacyUser をトランザクション実行
    Server->>Datastore: Put User("KEY" + firebaseUID)<br/>clearStageCount を引き継ぎ
    Server->>Datastore: Put UserMigration<br/>(旧キー→新キーの対応を記録)
    Server->>Datastore: Delete User("KEY" + twitterUID)
    Server->>Datastore: StageUser レコードの UserKey を新キーに更新
    Server-->>Client: ユーザー情報レスポンス
```

### UserMigration エンティティ

マイグレーション実行時に `UserMigration` Kind のレコードが作成される。
このレコードにより、どのユーザーがいつマイグレーションされたかを追跡できる。

| フィールド | 内容 |
|---|---|
| `oldKey` | 旧キー名 (例: `KEY12345`) |
| `newKey` | 新キー名 (例: `KEYabc123def`) |
| `twitterUid` | Twitter UID |
| `firebaseUid` | Firebase UID |
| `migratedAt` | マイグレーション実行日時 |
