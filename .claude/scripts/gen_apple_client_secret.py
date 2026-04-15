# /// script
# dependencies = [
#   "PyJWT==2.12.1",
#   "cryptography==46.0.7",
# ]
# ///
"""
Apple Sign In 用の client_secret JWT を生成するスクリプト。

Usage:
    uv run .claude/scripts/gen_apple_client_secret.py \
        --p8 <path_to_AuthKey.p8> \
        --team-id <TEAM_ID> \
        --key-id <KEY_ID> \
        --client-id <SERVICE_ID>
"""

import argparse
import base64
import json
import time

import jwt
from cryptography.hazmat.primitives.serialization import load_pem_private_key


def generate(p8_path: str, team_id: str, key_id: str, client_id: str) -> str:
    with open(p8_path, "rb") as f:
        key = load_pem_private_key(f.read(), password=None)

    now = int(time.time())
    payload = {
        "iss": team_id,
        "iat": now,
        "exp": now + 15777000,  # 約6ヶ月（Apple の上限）
        "aud": "https://appleid.apple.com",
        "sub": client_id,
    }
    return jwt.encode(payload, key, algorithm="ES256", headers={"kid": key_id})


def decode_payload(token: str) -> dict:
    payload_b64 = token.split(".")[1]
    padding = 4 - len(payload_b64) % 4
    payload_b64 += "=" * (padding % 4)
    return json.loads(base64.b64decode(payload_b64))


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate Apple Sign In client_secret JWT")
    parser.add_argument("--p8", required=True, help=".p8 秘密鍵ファイルのパス")
    parser.add_argument("--team-id", required=True, help="Apple Developer Team ID")
    parser.add_argument("--key-id", required=True, help="Key ID（.p8 ファイル名の AuthKey_XXXXXXXXXX の部分）")
    parser.add_argument("--client-id", required=True, help="Service ID（apple_client_id）")
    args = parser.parse_args()

    token = generate(args.p8, args.team_id, args.key_id, args.client_id)

    print("=== Generated client_secret ===")
    print(token)
    print()
    print("=== JWT Payload (verification) ===")
    print(json.dumps(decode_payload(token), indent=2))


if __name__ == "__main__":
    main()
