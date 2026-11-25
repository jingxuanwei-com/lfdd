// 访问控制：开启限制后，除 /admin 外页面都必须带有效令牌（URL 参数 ?access=<token>）。
// 令牌首次从 URL 读取后存入本地，站内跳转无需重复携带；每次路由跳转向后端校验一次
// （令牌可能被删除或过期），后端不可达时按拒绝处理。

const ACCESS_KEY = "lfdd_access_token";

// 读取已保存的访问令牌（空串表示没有）
export function getAccessToken(): string {
  try {
    return localStorage.getItem(ACCESS_KEY) ?? "";
  } catch {
    return "";
  }
}

// 保存 / 清除访问令牌
export function setAccessToken(token: string) {
  try {
    if (token) localStorage.setItem(ACCESS_KEY, token);
    else localStorage.removeItem(ACCESS_KEY);
  } catch {
    // 本地存储不可用：仅本次会话内的请求能读到 token
  }
}

// 校验当前是否可访问受限页面：urlToken 来自地址栏，存在时优先使用并落盘。
// 返回 true 表示放行。
export async function resolveAccess(urlToken?: unknown): Promise<boolean> {
  const fromUrl = typeof urlToken === "string" ? urlToken.trim() : "";
  if (fromUrl) setAccessToken(fromUrl);

  const token = fromUrl || getAccessToken();
  try {
    const res = await fetch("/api/access/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token })
    });
    if (!res.ok) return false;
    const data = await res.json().catch(() => ({}));
    return data?.allowed === true;
  } catch {
    return false;
  }
}
