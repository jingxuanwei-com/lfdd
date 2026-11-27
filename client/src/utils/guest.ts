// 游客身份：未登录访问时由后端签发（POST /api/auth/guest），与登录 token 共用 Authorization 头，
// 游客 token 以 guest_ 开头，服务端按前缀分流。未持有效身份的请求，业务接口一律不返回数据
// （401 + need_guest）；登录后忽略并作废游客身份，也可调用 revokeGuestIdentity 主动注销。

import { ref } from "vue";

// 登录态存储键（与 admin 页共用）
export const TOKEN_KEY = "lfdd_token";

const GUEST_KEY = "lfdd_guest_token";

// 上次轮换 token 的时间戳（毫秒）：多标签页共享，避免重复轮换
const ROTATED_AT_KEY = "lfdd_token_rotated_at";
// 轮换中防抖锁：避免多标签页同时轮换互相踢下线
const ROTATE_LOCK_KEY = "lfdd_token_rotating_at";
// 轮换周期：24 小时；轮换后旧 token 立即失效，有效期顺延 60 天
const ROTATE_INTERVAL = 24 * 60 * 60 * 1000;

// 登录态（响应式）：供底部导航等界面随登录/退出即时刷新
export const loggedIn = ref(false);

// 当前登录 token（响应式）：轮换/登录/退出时同步更新，页面请求直接读取
export const authToken = ref("");

function readStoredToken(): string {
  try {
    return localStorage.getItem(TOKEN_KEY) ?? "";
  } catch {
    return "";
  }
}

// 读取登录 token（空串表示未登录）：优先本地存储，其他标签页轮换后自动同步
export function getToken(): string {
  const stored = readStoredToken();
  if (stored !== authToken.value) authToken.value = stored;
  return stored;
}

// 保存/清除登录 token；写入即视为刚换发（重置轮换计时）
export function setAuthToken(t: string) {
  authToken.value = t;
  try {
    if (t) {
      localStorage.setItem(TOKEN_KEY, t);
      localStorage.setItem(ROTATED_AT_KEY, String(Date.now()));
    } else {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(ROTATED_AT_KEY);
    }
  } catch {
    // 本地存储不可用：仅本次会话有效
  }
}

// 标记已登录 / 已退出（登录、退出、401 失效时调用）
export function markLoggedIn() {
  loggedIn.value = true;
}

export function markLoggedOut() {
  loggedIn.value = false;
}

// 是否已登录
export function hasToken(): boolean {
  return getToken() !== "";
}

// 初始登录态：以本地 token 为准（refresh 时会再校正）
loggedIn.value = hasToken();

// 读取本地游客 token（空串表示还没申请）
export function getGuestToken(): string {
  try {
    return localStorage.getItem(GUEST_KEY) ?? "";
  } catch {
    return "";
  }
}

// 保存/清除游客 token
export function setGuestToken(token: string) {
  try {
    if (token) localStorage.setItem(GUEST_KEY, token);
    else localStorage.removeItem(GUEST_KEY);
  } catch {
    // 本地存储不可用：仅本次会话有效
  }
}

// 当前生效的身份 token：已登录用登录 token，否则用游客 token
export function activeToken(): string {
  return getToken() || getGuestToken();
}

// 是否游客：未登录且有游客身份（游客不显示/不使用结单、打印、改价、废单，也看不到已结单）
export function isGuestMode(): boolean {
  return !hasToken() && getGuestToken() !== "";
}

// 申请游客身份（已有则直接返回）：唯一允许无身份调用的接口
export async function ensureGuestIdentity(): Promise<string> {
  const token = getGuestToken();
  if (token) return token;

  const res = await fetch("/api/auth/guest", { method: "POST" });
  const data = await res.json().catch(() => ({}));
  if (!res.ok || typeof data.guest_id !== "string" || !data.guest_id) {
    throw new Error(data.message || `申请游客身份失败（${res.status}）`);
  }
  setGuestToken(data.guest_id);
  return data.guest_id;
}

// 主动注销游客身份：注销后该身份立即失效，下次请求会自动重新申请
export async function revokeGuestIdentity(): Promise<void> {
  const token = getGuestToken();
  if (!token) return;
  try {
    await fetch("/api/auth/guest/revoke", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` }
    });
  } catch {
    // 后端不可用时只清本地身份
  }
  setGuestToken("");
}

// 登录成功：清掉本地游客身份（后端登录时已作废该身份，登录后一律以登录身份为准）
export function clearGuestIdentity() {
  setGuestToken("");
}

// ================= 登录 token 自动轮换 =================
function lastRotatedAt(): number {
  try {
    return Number(localStorage.getItem(ROTATED_AT_KEY) || 0) || 0;
  } catch {
    return 0;
  }
}

// 轮换登录 token：旧 token 立即失效、有效期顺延 60 天。
// 每 24 小时最多一次；多标签页用本地时间戳与锁避免并发轮换互相踢下线。
export async function rotateSessionToken(): Promise<void> {
  const token = getToken();
  if (!token) return;
  if (Date.now() - lastRotatedAt() < ROTATE_INTERVAL) return;

  try {
    const lockAt = Number(localStorage.getItem(ROTATE_LOCK_KEY) || 0) || 0;
    if (lockAt && Date.now() - lockAt < 15000) return;
    localStorage.setItem(ROTATE_LOCK_KEY, String(Date.now()));
  } catch {
    // 本地存储不可用时不做防抖，直接轮换
  }

  try {
    const res = await fetch("/api/auth/refresh", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` }
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok && typeof data.token === "string" && data.token) {
      setAuthToken(data.token);
      return;
    }
    // 旧 token 已被其他标签页轮换：本地已换成新 token，直接采用
    getToken();
  } catch {
    // 网络异常：保留旧 token，下次再试
  } finally {
    try {
      localStorage.removeItem(ROTATE_LOCK_KEY);
    } catch {
      // 忽略
    }
  }
}
