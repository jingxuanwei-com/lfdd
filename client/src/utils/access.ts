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

// 首次通过链接进入时，把地址里的访问令牌存入本地：
// 兼容 ?access=xxx 与 hash 模式 #/?access=xxx（hash 不会发给服务端）
export function captureAccessTokenFromUrl() {
  const search = new URLSearchParams(window.location.search).get("access");
  const hash = window.location.hash;
  const qIndex = hash.indexOf("?");
  const hashToken =
    qIndex >= 0
      ? new URLSearchParams(hash.slice(qIndex + 1)).get("access")
      : null;
  const token = (search ?? hashToken ?? "").trim();
  if (token) setAccessToken(token);
}

// 准入结果缓存（true 允许 / false 拒绝 / null 未确认）：
// 确认过就不再每次路由跳转请求校验接口；启动时游客申请成功即视为允许，
// 之后任意受控接口返回「X-Access-Denied」时由 fetch 补丁标记为拒绝并跳转
let allowedCache: boolean | null = null;
let deniedHandler: (() => void) | null = null;

// 注册「访问被拒」回调（由路由模块注入，负责跳转伪装页或登录页）
export function setAccessDeniedHandler(fn: () => void) {
  deniedHandler = fn;
}

// 服务端已放行过请求（如游客申请成功）：本会话视为允许
export function markAccessAllowed() {
  allowedCache = true;
}

// 受控接口返回了访问受限的伪装 404：标记拒绝，并在状态由允许/未知转为拒绝时通知跳转
export function markAccessDenied() {
  const changed = allowedCache !== false;
  allowedCache = false;
  if (changed) deniedHandler?.();
}

// 校验当前是否可访问受限页面：urlToken 来自地址栏，存在时优先使用并落盘。
// 返回 true 表示放行。已有缓存时不重复请求。
export async function resolveAccess(urlToken?: unknown): Promise<boolean> {
  const fromUrl = typeof urlToken === "string" ? urlToken.trim() : "";
  if (fromUrl) {
    setAccessToken(fromUrl);
    allowedCache = null; // 换了令牌，重新校验
  }
  if (allowedCache !== null) return allowedCache;

  const token = fromUrl || getAccessToken();
  try {
    const res = await fetch("/api/access/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token })
    });
    if (!res.ok) return false;
    const data = await res.json().catch(() => ({}));
    allowedCache = data?.allowed === true;
    return allowedCache;
  } catch {
    return false;
  }
}
