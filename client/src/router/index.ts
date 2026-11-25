import { defineRouter } from "#q-app";
import { routes, handleHotUpdate } from "vue-router/auto-routes";
import {
  createMemoryHistory,
  createRouter,
  createWebHashHistory,
  createWebHistory
} from "vue-router";
import { resolveAccess, setAccessDeniedHandler } from "@/utils/access";

// App（Capacitor 原生）内置前端资源、不对外暴露，未授权时直接回 /admin 登录；网页端跳伪装页
function isNativeApp(): boolean {
  return Boolean(
    (
      window as unknown as {
        Capacitor?: { isNativePlatform?: () => boolean };
      }
    ).Capacitor?.isNativePlatform?.()
  );
}

/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */

export default defineRouter((/* { store, ssrContext } */) => {
  const createHistory = import.meta.env.QUASAR_SERVER
    ? createMemoryHistory
    : import.meta.env.QUASAR_VUE_ROUTER_MODE === "history"
      ? createWebHistory
      : createWebHashHistory;

  const Router = createRouter({
    scrollBehavior: () => ({ left: 0, top: 0 }),
    routes,

    // Leave this as is and make changes in quasar.conf.js instead!
    // quasar.conf.js -> build -> vueRouterMode
    // quasar.conf.js -> build -> publicPath
    history: createHistory(import.meta.env.QUASAR_VUE_ROUTER_BASE)
  });

  // enable HMR for it
  if (import.meta.hot) {
    handleHotUpdate(Router);
  }

  // 任意受控接口返回「访问受限」时自动跳转（不必每次路由跳转都请求校验接口）
  setAccessDeniedHandler(() => {
    const path = Router.currentRoute.value.path;
    if (path === "/admin" || path === "/denied") return;
    void Router.push(isNativeApp() ? "/admin" : "/denied");
  });

  // 访问限制守卫：开启后除 /admin（配置入口）与 /denied（提示页）外，
  // 其余页面都必须持有效访问令牌（URL 参数 ?access=xxx 或本地已保存）。
  // App（Capacitor 原生）内置前端资源、不对外暴露，未授权时直接回 /admin 登录；
  // 网页端则跳转未授权提示页。
  Router.beforeEach(async to => {
    if (to.path === "/admin" || to.path === "/denied") return true;

    // hash 模式下参数在 #/ 之后；兼容写在 ? 后的形式
    const searchToken = new URLSearchParams(window.location.search).get(
      "access"
    );
    const allowed = await resolveAccess(to.query.access ?? searchToken);
    if (allowed) return true;

    return { path: isNativeApp() ? "/admin" : "/denied" };
  });

  return Router;
});
