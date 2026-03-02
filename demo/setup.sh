#!/usr/bin/env bash
set -euo pipefail

DEMO_DIR=/tmp/git-cx-demo

# 前回の残骸をクリーンアップ
rm -rf "$DEMO_DIR"
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"

git init -q
git config user.email "demo@example.com"
git config user.name "Demo User"

# 初期コミット
echo "# Demo Project" > README.md
git add README.md
git commit -q -m "initial commit"

# デモ用ファイルを作成してステージング
cat > auth.ts << 'EOF'
export interface JWTPayload {
  userId: string;
  exp: number;
  iat: number;
}

export async function refreshToken(token: string): Promise<string> {
  const res = await fetch("/api/auth/refresh", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("refresh failed");
  const { token: newToken } = await res.json();
  return newToken;
}

export function validateToken(token: string): JWTPayload | null {
  try {
    const [, payload] = token.split(".");
    return JSON.parse(atob(payload)) as JWTPayload;
  } catch {
    return null;
  }
}
EOF

git add auth.ts

# フェイク gemini バイナリを生成
cat > /tmp/gemini << 'AISCRIPT'
#!/usr/bin/env bash
echo "feat(auth): add JWT token refresh mechanism"
echo "feat(auth): implement OAuth2 password grant flow"
echo "feat(auth): add session expiry handling"
AISCRIPT
chmod +x /tmp/gemini

echo "Setup complete: $DEMO_DIR"
