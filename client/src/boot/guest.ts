import { defineBoot } from "#q-app";
import { apiPath } from "@/utils/api";
import {
  captureAccessTokenFromUrl,
  getAccessToken,
  markAccessAllowed,
  markAccessDenied
} from "@/utils/access";
import {
  activeToken,
  ensureGuestIdentity,
  hasToken,
  rotateSessionToken,
  setGuestToken
} from "@/utils/guest";

// 是否已给全局 fetch 打过补丁（HMR 可能重复执行 boot）
const PATCHED = "__lfddFetchPatched";

// 统一给 /api 请求带上身份（登录 token 或游客 token，均走 Authorization）：
// 未登录且还没有游客身份时先自动申请，游客身份失效时自动补签一次后重试。
export default defineBoot(async () => {
  // 先打补丁，保证下面的游客身份申请也会被改写成本后端地址（App 内不是同源）
  patchFetch();

  // 链接首次进入时先把访问令牌落盘，后续游客申请与接口请求才能带上它
  captureAccessTokenFromUrl();

  // 登录态自动轮换 token：启动时检查（距上次轮换不足 24 小时则跳过）
  await rotateSessionToken().catch(() => {});
  startRotateTimer();

  if (!hasToken()) {
    try {
      await ensureGuestIdentity();
      // 受控接口能成功返回，说明访问限制已放行，本会话无需再单独校验
      markAccessAllowed();
    } catch {
      // 后端不可用时忽略，页面会给出加载失败提示
    }
  }
});

function patchFetch() {
  const w = window as unknown as Record<string, unknown>;
  if (w[PATCHED]) return;
  w[PATCHED] = true;

  const rawFetch = window.fetch.bind(window);

  window.fetch = async (input, init = {}) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof Request
          ? input.url
          : String(input);
    if (!url.includes("/api/")) return rawFetch(input, init);

    // 前后端分离部署时把 /api 相对路径改写到 VITE_API_BASE（同源部署时 apiPath 原样返回）
    const target = typeof input === "string" ? apiPath(input) : input;

    const send = () => {
      const headers = new Headers(init.headers);
      const token = activeToken();
      if (token && !headers.has("Authorization"))
        headers.set("Authorization", `Bearer ${token}`);
      // 访问令牌：开启访问限制后服务端才放行（未授权返回伪装 404）
      const access = getAccessToken();
      if (access && !headers.has("X-Access-Token"))
        headers.set("X-Access-Token", access);
      return rawFetch(target, { ...init, headers });
    };

    let res = await send();

    // 访问受限：服务端返回伪装 404 并带标记，标记为拒绝并交给路由跳转
    if (isAccessDenied(res)) markAccessDenied();

    // 游客身份失效（已注销/服务端无此身份）：自动补签一次后重试
    if (res.status === 401 && !hasToken()) {
      const data = await res
        .clone()
        .json()
        .catch(() => ({}) as { need_guest?: boolean });
      if (data && data.need_guest) {
        setGuestToken("");
        try {
          await ensureGuestIdentity();
          res = await send();
          if (isAccessDenied(res)) markAccessDenied();
        } catch {
          // 补签失败：保留原 401 响应，由页面提示
        }
      }
    }
    return res;
  };
}

// 服务端访问限制的伪装 404：带 X-Access-Denied 标记（CORS 已放开该响应头）
function isAccessDenied(res: Response): boolean {
  return res.status === 404 && res.headers.get("X-Access-Denied") === "1";
}

// 长时间驻留的 App：每小时检查一次是否到轮换时间，回到前台时也检查
const ROTATE_TIMER = "__lfddRotateTimer";
function startRotateTimer() {
  const w = window as unknown as Record<string, unknown>;
  if (w[ROTATE_TIMER]) return;
  w[ROTATE_TIMER] = true;
  setInterval(() => void rotateSessionToken(), 60 * 60 * 1000);
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") void rotateSessionToken();
  });
}
