<template>
  <q-page class="goods-page" :style-fn="pageStyle">
    <div class="row no-wrap goods-body">
      <!-- 左侧分类栏（蜜雪冰城/肯德基式窄边栏） -->
      <div ref="sideRef" class="goods-sidebar">
        <div
          class="sidebar-item"
          :class="{ active: activeMenuId === '0' }"
          @click="selectMenu('0')"
        >
          全部
        </div>
        <div
          v-for="m in menus"
          :key="m.id"
          class="sidebar-item"
          :class="{ active: activeMenuId === m.id }"
          @click="selectMenu(m.id)"
        >
          {{ m.name }}
        </div>
        <div v-if="menus.length === 0" class="text-caption text-grey-6 q-pa-sm text-center">
          暂无分类
        </div>
      </div>

      <!-- 右侧商品列表 -->
      <div ref="mainRef" class="goods-main">
        <div v-if="errorMsg" class="error-bar text-negative text-caption">
          {{ errorMsg }}
        </div>

        <div v-if="loadingGoods" class="state-box text-grey-6">加载中…</div>
        <div v-else-if="goods.length === 0 && !errorMsg" class="state-box text-grey-6">
          暂无商品
        </div>

        <template v-else>
          <div v-for="g in goods" :key="g.id" class="goods-card">
            <div class="goods-img">
              <img v-if="g.image" :src="g.image" :alt="g.name" />
              <q-icon v-else name="shopping_bag" size="32px" color="grey-5" />
              <div v-if="g.stock === 0" class="sold-out">售罄</div>
            </div>
            <div class="col">
              <div class="goods-name text-weight-medium">{{ g.name }}</div>
              <div v-if="g.spec" class="text-caption text-grey-7 ellipsis">{{ g.spec }}</div>
              <div class="row items-end justify-between q-mt-xs">
                <div class="text-negative text-weight-bold">¥{{ fmtPrice(g.price) }}</div>
                <span class="text-caption text-grey-6">
                  {{ g.stock === 0 ? '已售罄' : g.stock === -1 ? '' : `库存 ${g.stock}` }}
                </span>
              </div>
            </div>
          </div>

          <div class="text-center q-pb-md">
            <q-btn
              v-if="hasMore"
              flat
              dense
              color="primary"
              :loading="loadingMore"
              label="加载更多"
              @click="loadGoods(false)"
            />
            <span v-else class="text-caption text-grey-6">— 没有更多了 —</span>
          </div>
        </template>
      </div>
    </div>
  </q-page>
</template>

<script lang="ts">
// 模块级：跨路由切换保留的商品页位置（组件卸载后仍然存在）
// 切到购物车/订单/后台再切回来时，恢复分类和两栏滚动位置
const pagePos = {
  activeMenuId: "0",
  side: 0,
  main: 0
};
</script>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";

// 与后端 service 对齐的接口返回结构
interface SidebarMenu {
  id: string; // `,string` tag，如 "101"
  name: string;
}

interface GoodsItem {
  id: string;
  menu_id: string;
  name: string;
  price: number; // 单位：分（1850 = 18.50 元）
  spec: string;
  stock: number; // -1 = 不限库存，0 = 售罄
  image: string;
  status: number;
}

const PAGE_SIZE = 20;

const menus = ref<SidebarMenu[]>([]);
const activeMenuId = ref(pagePos.activeMenuId); // 恢复上次选中的分类
const goods = ref<GoodsItem[]>([]);
const total = ref(0);
const page = ref(1);
const loadingGoods = ref(false);
const loadingMore = ref(false);
const errorMsg = ref("");

const sideRef = ref<HTMLElement | null>(null);
const mainRef = ref<HTMLElement | null>(null);

let reqSeq = 0; // 递增序号，丢弃切换分类时已过期的响应

const hasMore = computed(() => goods.value.length < total.value);

// QPage styleFn：header/footer 占位后剩余高度，让页面撑满且两栏各自滚动
function pageStyle(offset: number, height: number) {
  if (height > 0) {
    return { height: `${height - offset}px` };
  }
  return { height: offset !== 0 ? `calc(100vh - ${offset}px)` : "100vh" };
}

function fmtPrice(price: number) {
  return (price / 100).toFixed(2);
}

function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

async function fetchJson(url: string) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`请求失败（${res.status}）`);
  }
  return res.json();
}

// 左侧分类列表
async function loadMenus() {
  try {
    const data = await fetchJson("/api/goods/sidebar");
    menus.value = Array.isArray(data.menus) ? data.menus : [];
  } catch (e) {
    menus.value = [];
    // 分类失败不阻断商品加载（menu_id=0 仍可看全部）
    console.error("加载侧边分类失败:", e);
  }
}

// 商品列表（reset=true 重新第一页，false 加载更多）
async function loadGoods(reset: boolean) {
  const seq = ++reqSeq;
  if (reset) {
    page.value = 1;
    errorMsg.value = "";
    loadingGoods.value = true;
  } else {
    loadingMore.value = true;
  }

  try {
    const data = await fetchJson(
      `/api/goods/page?menu_id=${activeMenuId.value}&page=${page.value}&page_size=${PAGE_SIZE}`
    );
    if (seq !== reqSeq) return; // 已有更新的请求，丢弃本次结果

    const list: GoodsItem[] = Array.isArray(data.list) ? data.list : [];
    goods.value = reset ? list : [...goods.value, ...list];
    total.value = typeof data.total === "number" ? data.total : 0;
    if (list.length > 0) {
      page.value += 1;
    }
  } catch (e) {
    if (seq !== reqSeq) return;
    if (reset) {
      goods.value = [];
      total.value = 0;
    }
    errorMsg.value = `加载失败：${describeError(e)}，请确认后端服务已启动`;
  } finally {
    if (seq === reqSeq) {
      loadingGoods.value = false;
      loadingMore.value = false;
    }
  }
}

function selectMenu(id: string) {
  activeMenuId.value = id;
  pagePos.activeMenuId = id;
  void loadGoods(true);
  // 切分类后商品区回到顶部
  if (mainRef.value) mainRef.value.scrollTop = 0;
}

onMounted(async () => {
  await Promise.all([loadMenus(), loadGoods(true)]);
  // 数据渲染后恢复上次的滚动位置
  await nextTick();
  if (sideRef.value) sideRef.value.scrollTop = pagePos.side;
  if (mainRef.value) mainRef.value.scrollTop = pagePos.main;
});

onBeforeUnmount(() => {
  pagePos.activeMenuId = activeMenuId.value;
  pagePos.side = sideRef.value?.scrollTop ?? 0;
  pagePos.main = mainRef.value?.scrollTop ?? 0;
});
</script>

<style scoped lang="scss">
.goods-page {
  overflow: hidden;
}

.goods-body {
  height: 100%;
}

/* 左侧分类栏 */
.goods-sidebar {
  width: 88px;
  flex-shrink: 0;
  overflow-y: auto;
  background: #f5f5f5;
}

.sidebar-item {
  position: relative;
  padding: 15px 6px;
  text-align: center;
  font-size: 13px;
  line-height: 1.3;
  color: #666;
  cursor: pointer;
  word-break: break-all;

  &.active {
    background: #fff;
    color: $primary;
    font-weight: 600;

    &::before {
      content: "";
      position: absolute;
      left: 0;
      top: 20%;
      bottom: 20%;
      width: 3px;
      border-radius: 0 3px 3px 0;
      background: $primary;
    }
  }
}

/* 右侧商品区 */
.goods-main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding: 8px 12px;
}

.error-bar {
  padding: 8px 0 4px;
}

.state-box {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.goods-card {
  display: flex;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-of-type {
    border-bottom: none;
  }
}

.goods-img {
  position: relative;
  width: 72px;
  height: 72px;
  flex-shrink: 0;
  border-radius: 8px;
  background: #fafafa;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.sold-out {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 13px;
}

.goods-name {
  font-size: 14px;
}
</style>
