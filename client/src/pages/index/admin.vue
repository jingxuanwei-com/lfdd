<template>
  <q-page class="q-pa-md">
    <!-- ================= 未登录：登录 / 注册 ================= -->
    <div v-if="!token" class="flex flex-center" style="min-height: 60vh">
      <q-card style="width: 360px; max-width: 95vw">
        <q-tabs
          v-model="authTab"
          dense
          class="text-grey-7"
          active-color="primary"
          indicator-color="primary"
          align="justify"
          narrow-indicator
        >
          <q-tab name="login" label="登录" />
          <q-tab name="register" label="注册" />
        </q-tabs>
        <q-separator />

        <q-card-section class="column q-gutter-sm">
          <q-input
            v-if="authTab === 'register'"
            v-model="regEmail"
            label="邮箱"
            type="email"
            outlined
            dense
            autofocus
          />
          <q-input
            v-else
            v-model="account"
            label="账号（用户名 / 邮箱）"
            outlined
            dense
            autofocus
          />
          <q-input
            v-model="password"
            label="密码"
            type="password"
            outlined
            dense
            :hint="authTab === 'register' ? '至少 8 位' : undefined"
            @keyup.enter="submitAuth"
          />
          <div v-if="authError" class="text-negative text-caption">{{ authError }}</div>
          <q-btn
            color="primary"
            :label="authTab === 'login' ? '登录' : '注册并登录'"
            :loading="authBusy"
            @click="submitAuth"
          />
        </q-card-section>
      </q-card>
    </div>

    <!-- ================= 已登录：管理界面 ================= -->
    <div v-else class="admin-wrap">
      <div class="row items-center q-mb-sm">
        <div class="text-h6">后台管理</div>
        <q-space />
        <q-btn flat dense color="negative" icon="logout" label="退出登录" @click="logout" />
      </div>

      <q-tabs
        v-model="tab"
        dense
        class="text-grey-7"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
      >
        <q-tab name="goods" label="商品管理" icon="inventory_2" />
        <q-tab name="menus" label="分类管理" icon="category" />
        <q-tab name="print" label="蓝牙搜索" icon="bluetooth_searching" />
      </q-tabs>
      <q-separator />

      <q-tab-panels v-model="tab" animated class="bg-transparent">
        <!-- ============ 商品管理 ============ -->
        <q-tab-panel name="goods" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <q-select
              v-model="filterMenu"
              :options="menuFilterOptions"
              emit-value
              map-options
              dense
              outlined
              label="分类"
              style="min-width: 160px"
            />
            <q-space />
            <q-btn
              color="primary"
              icon="add"
              label="新建商品"
              :disable="menus.length === 0"
              @click="openGoods()"
            />
          </q-toolbar>

          <div v-if="menus.length === 0" class="text-grey-6 q-pa-md">
            还没有分类，先到「分类管理」创建一个分类。
          </div>

          <div v-else-if="goodsLoading" class="text-grey-6 q-pa-md">加载中…</div>

          <div v-else-if="goodsList.length === 0" class="text-grey-6 q-pa-md">
            该分类下暂无商品，点右上角「新建商品」添加。
          </div>

          <q-list v-else separator>
            <q-item v-for="g in goodsList" :key="g.id">
              <q-item-section avatar>
                <div class="goods-thumb">
                  <img v-if="g.image" :src="g.image" :alt="g.name" />
                  <q-icon v-else name="shopping_bag" size="22px" color="grey-5" />
                </div>
              </q-item-section>
              <q-item-section>
                <q-item-label class="text-weight-medium">{{ g.name }}</q-item-label>
                <q-item-label caption>
                  {{ menuName(g.menu_id) }}
                  <span v-if="g.spec"> · {{ g.spec }}</span>
                  · 库存 {{ g.stock === -1 ? "不限" : g.stock }}
                </q-item-label>
              </q-item-section>
              <q-item-section side class="text-negative text-weight-bold">
                ¥{{ (g.price / 100).toFixed(2) }}
              </q-item-section>
              <q-item-section side>
                <q-toggle
                  :model-value="g.status === 1"
                  :label="g.status === 1 ? '上架' : '下架'"
                  dense
                  color="positive"
                  @update:model-value="(v: boolean) => toggleStatus(g, v)"
                />
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap">
                  <q-btn dense flat round icon="edit" @click="openGoods(g)" />
                  <q-btn dense flat round icon="delete" color="negative" @click="removeGoods(g)" />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <!-- ============ 分类管理 ============ -->
        <q-tab-panel name="menus" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <div class="text-grey-7 text-caption">上移 / 下移调整前台侧边栏顺序</div>
            <q-space />
            <q-btn color="primary" icon="add" label="新建分类" @click="openMenu()" />
          </q-toolbar>

          <div v-if="menus.length === 0" class="text-grey-6 q-pa-md">暂无分类。</div>

          <q-list v-else separator>
            <q-item v-for="(m, idx) in menus" :key="m.id">
              <q-item-section>
                <q-item-label>{{ m.name }}</q-item-label>
                <q-item-label caption>ID {{ m.id }}</q-item-label>
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap">
                  <q-btn
                    dense
                    flat
                    round
                    icon="arrow_upward"
                    :disable="idx === 0"
                    @click="moveMenu(idx, -1)"
                  />
                  <q-btn
                    dense
                    flat
                    round
                    icon="arrow_downward"
                    :disable="idx === menus.length - 1"
                    @click="moveMenu(idx, 1)"
                  />
                  <q-btn dense flat round icon="edit" @click="openMenu(m)" />
                  <q-btn dense flat round icon="delete" color="negative" @click="removeMenu(m)" />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <!-- ============ 蓝牙搜索（双列：低功耗 / 传统） ============ -->
        <q-tab-panel name="print" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <q-btn
              color="primary"
              icon="bluetooth_searching"
              :label="scanning ? '搜索中…' : '搜索蓝牙'"
              :loading="scanning"
              @click="startScan"
            />
            <q-btn
              v-if="scanning"
              class="q-ml-sm"
              flat
              label="停止"
              @click="stopScan"
            />
            <q-space />
            <span class="text-caption text-grey-6">点击设备发送测试小票</span>
          </q-toolbar>

          <q-banner
            v-if="status"
            dense
            rounded
            class="q-mb-sm"
            :class="statusOk ? 'bg-green-1 text-green-9' : 'bg-red-1 text-red-9'"
          >
            {{ status }}
          </q-banner>

          <div class="row q-col-gutter-sm full-height">
            <!-- 低功耗蓝牙 -->
            <div class="col-6">
              <q-list bordered separator class="scan-list">
                <q-item-label header class="bg-blue-1 text-blue-8">
                  低功耗蓝牙（BLE）
                </q-item-label>
                <q-item v-if="bleDevices.length === 0">
                  <q-item-section class="text-grey-6">
                    {{ scanning ? "搜索中，等待附近设备广播…" : "未搜索到设备" }}
                  </q-item-section>
                </q-item>
                <q-item
                  v-for="d in bleDevices"
                  :key="d.deviceId"
                  clickable
                  v-ripple
                  @click="printBle(d)"
                >
                  <q-item-section avatar>
                    <q-icon name="bluetooth" color="primary" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ d.name || "未命名设备" }}</q-item-label>
                    <q-item-label caption>{{ d.deviceId }}</q-item-label>
                  </q-item-section>
                  <q-item-section v-if="d.rssi != null" side caption>
                    {{ d.rssi }} dBm
                  </q-item-section>
                </q-item>
              </q-list>
            </div>

            <!-- 传统蓝牙 -->
            <div class="col-6">
              <q-list bordered separator class="scan-list">
                <q-item-label header class="bg-teal-1 text-teal-8">
                  传统蓝牙（SPP）
                </q-item-label>
                <q-item v-if="sppDevices.length === 0">
                  <q-item-section class="text-grey-6">
                    {{ scanning ? "搜索中，等待设备进入配对模式…" : "未搜索到设备" }}
                  </q-item-section>
                </q-item>
                <q-item
                  v-for="d in sppDevices"
                  :key="d.address"
                  clickable
                  v-ripple
                  @click="printSpp(d)"
                >
                  <q-item-section avatar>
                    <q-icon name="settings_input_antenna" color="teal" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ d.name || "未命名设备" }}</q-item-label>
                    <q-item-label caption>{{ d.address }}</q-item-label>
                  </q-item-section>
                </q-item>
              </q-list>
            </div>
          </div>
        </q-tab-panel>
      </q-tab-panels>

      <!-- ============ 新建 / 编辑商品 ============ -->
      <q-dialog v-model="goodsDlg">
        <q-card style="min-width: 360px; max-width: 95vw">
          <q-card-section class="text-subtitle1">
            {{ goodsForm.id ? "编辑商品" : "新建商品" }}
          </q-card-section>

          <q-card-section class="column q-gutter-sm">
            <q-select
              v-model="goodsForm.menu_id"
              :options="menuOptions"
              emit-value
              map-options
              label="所属分类"
              outlined
              dense
            />
            <q-input v-model="goodsForm.name" label="名称" outlined dense autofocus />
            <q-input
              v-model="goodsForm.priceYuan"
              label="价格（元）"
              type="number"
              step="0.01"
              min="0"
              outlined
              dense
            />
            <q-input v-model="goodsForm.spec" label="规格（如：大杯 / 500g，可空）" outlined dense />
            <q-input
              v-model="goodsForm.stock"
              label="库存"
              type="number"
              outlined
              dense
              hint="-1 = 不限库存，0 = 售罄"
            />
            <q-input v-model="goodsForm.image" label="图片地址（URL，可空）" outlined dense />
            <div v-if="goodsForm.image" class="row items-center q-gutter-sm">
              <img
                :src="goodsForm.image"
                class="goods-thumb"
                style="width: 48px; height: 48px"
                alt=""
              />
              <span class="text-caption text-grey-6">预览</span>
            </div>
            <q-toggle v-model="goodsForm.statusOn" label="上架销售" color="positive" />
          </q-card-section>

          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn color="primary" label="保存" :loading="goodsSaving" @click="saveGoods" />
          </q-card-actions>
        </q-card>
      </q-dialog>

      <!-- ============ 新建 / 编辑分类 ============ -->
      <q-dialog v-model="menuDlg">
        <q-card style="min-width: 300px; max-width: 95vw">
          <q-card-section class="text-subtitle1">
            {{ menuForm.id ? "编辑分类" : "新建分类" }}
          </q-card-section>
          <q-card-section>
            <q-input
              v-model="menuForm.name"
              label="分类名称"
              outlined
              dense
              autofocus
              @keyup.enter="saveMenu"
            />
          </q-card-section>
          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn color="primary" label="保存" :loading="menuSaving" @click="saveMenu" />
          </q-card-actions>
        </q-card>
      </q-dialog>

    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useQuasar } from "quasar";
import {
  bleInit,
  blePrintTo,
  bleScanStart,
  encodeGbk,
  isNative,
  sppDiscover,
  sppPrint,
  type BleScanDevice,
  type SppDevice
} from "@/utils/printer";

const $q = useQuasar();

const TOKEN_KEY = "lfdd_token";

// ================= 类型 =================
interface SidebarMenu {
  id: string;
  name: string;
}

interface AdminGoods {
  id: string;
  menu_id: string;
  name: string;
  price: number; // 分
  spec: string;
  stock: number;
  image: string;
  sort: number;
  status: number;
}

// ================= 通用 =================
function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

// ================= 认证 =================
const token = ref(localStorage.getItem(TOKEN_KEY) ?? "");
const authTab = ref("login");
const account = ref("");
const regEmail = ref("");
const password = ref("");
const authBusy = ref(false);
const authError = ref("");

function setToken(t: string) {
  token.value = t;
  localStorage.setItem(TOKEN_KEY, t);
}

// 统一请求：自动带 Bearer token，401 时清理登录态
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function api(path: string, opts: { method?: string; body?: unknown } = {}): Promise<any> {
  const res = await fetch(path, {
    method: opts.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(token.value ? { Authorization: `Bearer ${token.value}` } : {})
    },
    ...(opts.body !== undefined ? { body: JSON.stringify(opts.body) } : {})
  });

  if (res.status === 401) {
    token.value = "";
    localStorage.removeItem(TOKEN_KEY);
    throw new Error("登录已过期，请重新登录");
  }

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.message || `请求失败（${res.status}）`);
  }
  return data;
}

async function submitAuth() {
  authError.value = "";
  authBusy.value = true;
  try {
    const isRegister = authTab.value === "register";
    const data = await api(isRegister ? "/api/auth/register" : "/api/auth/login", {
      method: "POST",
      body: isRegister
        ? { email: regEmail.value.trim(), password: password.value }
        : { account: account.value.trim(), password: password.value }
    });
    setToken(data.token);
    password.value = "";
    await loadMenus();
    await loadGoods();
  } catch (e) {
    authError.value = describeError(e);
  } finally {
    authBusy.value = false;
  }
}

async function logout() {
  try {
    await api("/api/auth/logout", { method: "POST" });
  } catch {
    // token 已失效也照常退出
  }
  token.value = "";
  localStorage.removeItem(TOKEN_KEY);
  menus.value = [];
  goodsList.value = [];
}

// ================= 分类 =================
const menus = ref<SidebarMenu[]>([]);
const menuDlg = ref(false);
const menuSaving = ref(false);
const menuForm = reactive({ id: "", name: "" });

const menuOptions = computed(() => menus.value.map((m) => ({ label: m.name, value: m.id })));
const menuFilterOptions = computed(() => [
  { label: "全部分类", value: "0" },
  ...menuOptions.value
]);

function menuName(id: string) {
  return menus.value.find((m) => m.id === id)?.name ?? "-";
}

async function loadMenus() {
  const data = await api("/api/goods/sidebar");
  menus.value = Array.isArray(data.menus) ? data.menus : [];
}

function openMenu(m?: SidebarMenu) {
  menuForm.id = m?.id ?? "";
  menuForm.name = m?.name ?? "";
  menuDlg.value = true;
}

async function saveMenu() {
  const name = menuForm.name.trim();
  if (!name) {
    $q.notify({ message: "请输入分类名称", color: "negative" });
    return;
  }
  menuSaving.value = true;
  try {
    if (menuForm.id) {
      await api("/api/menu/update", { method: "POST", body: { id: menuForm.id, name } });
    } else {
      await api("/api/menu/create", { method: "POST", body: { name, sort: menus.value.length } });
    }
    $q.notify({ message: "保存成功", color: "positive" });
    menuDlg.value = false;
    await loadMenus();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    menuSaving.value = false;
  }
}

function removeMenu(m: SidebarMenu) {
  $q.dialog({
    title: "删除分类",
    message: `确定删除「${m.name}」？分类下有商品时会被后端拒绝。`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/menu/delete", { method: "POST", body: { id: m.id } });
      $q.notify({ message: "已删除", color: "positive" });
      await loadMenus();
      if (filterMenu.value === m.id) filterMenu.value = "0";
      await loadGoods();
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

async function moveMenu(idx: number, dir: -1 | 1) {
  const j = idx + dir;
  if (j < 0 || j >= menus.value.length) return;
  const ids = menus.value.map((m) => m.id);
  const tmp = ids[idx]!;
  ids[idx] = ids[j]!;
  ids[j] = tmp;
  try {
    await api("/api/menu/sort", { method: "POST", body: { ids } });
    await loadMenus();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  }
}

// ================= 商品 =================
const tab = ref("goods");
const filterMenu = ref("0");
const goodsList = ref<AdminGoods[]>([]);
const goodsLoading = ref(false);
const goodsDlg = ref(false);
const goodsSaving = ref(false);
const goodsForm = reactive({
  id: "",
  menu_id: "",
  name: "",
  priceYuan: "0",
  spec: "",
  stock: "-1",
  image: "",
  statusOn: true,
  sort: 0
});

async function loadGoods() {
  goodsLoading.value = true;
  try {
    const params = new URLSearchParams({ all: "1", page: "1", page_size: "100" });
    if (filterMenu.value !== "0") params.set("menu_id", filterMenu.value);
    const data = await api(`/api/goods/page?${params.toString()}`);
    goodsList.value = Array.isArray(data.list) ? data.list : [];
  } finally {
    goodsLoading.value = false;
  }
}

function openGoods(g?: AdminGoods) {
  if (menus.value.length === 0) {
    $q.notify({ message: "请先创建分类", color: "negative" });
    return;
  }
  if (g) {
    goodsForm.id = g.id;
    goodsForm.menu_id = g.menu_id;
    goodsForm.name = g.name;
    goodsForm.priceYuan = (g.price / 100).toFixed(2);
    goodsForm.spec = g.spec;
    goodsForm.stock = String(g.stock);
    goodsForm.image = g.image;
    goodsForm.statusOn = g.status === 1;
    goodsForm.sort = g.sort;
  } else {
    goodsForm.id = "";
    goodsForm.menu_id = filterMenu.value !== "0" ? filterMenu.value : menus.value[0]!.id;
    goodsForm.name = "";
    goodsForm.priceYuan = "0";
    goodsForm.spec = "";
    goodsForm.stock = "-1";
    goodsForm.image = "";
    goodsForm.statusOn = true;
    goodsForm.sort = 0;
  }
  goodsDlg.value = true;
}

async function saveGoods() {
  const name = goodsForm.name.trim();
  const price = Math.round(parseFloat(goodsForm.priceYuan || "0") * 100);
  const stock = parseInt(goodsForm.stock, 10);

  if (!name) {
    $q.notify({ message: "请输入商品名称", color: "negative" });
    return;
  }
  if (!goodsForm.menu_id) {
    $q.notify({ message: "请选择所属分类", color: "negative" });
    return;
  }
  if (!Number.isFinite(price) || price < 0) {
    $q.notify({ message: "价格无效", color: "negative" });
    return;
  }
  if (!Number.isInteger(stock)) {
    $q.notify({ message: "库存需为整数（-1 = 不限）", color: "negative" });
    return;
  }

  goodsSaving.value = true;
  try {
    const body = {
      menu_id: goodsForm.menu_id,
      name,
      price,
      spec: goodsForm.spec.trim(),
      stock,
      image: goodsForm.image.trim(),
      status: goodsForm.statusOn ? 1 : 0,
      sort: goodsForm.sort
    };
    if (goodsForm.id) {
      await api("/api/goods/update", { method: "POST", body: { id: goodsForm.id, ...body } });
    } else {
      await api("/api/goods/create", { method: "POST", body });
    }
    $q.notify({ message: "保存成功", color: "positive" });
    goodsDlg.value = false;
    await loadGoods();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    goodsSaving.value = false;
  }
}

async function toggleStatus(g: AdminGoods, on: boolean) {
  try {
    await api("/api/goods/update", {
      method: "POST",
      body: {
        id: g.id,
        menu_id: g.menu_id,
        name: g.name,
        price: g.price,
        spec: g.spec,
        stock: g.stock,
        image: g.image,
        status: on ? 1 : 0,
        sort: g.sort
      }
    });
    g.status = on ? 1 : 0;
    $q.notify({ message: on ? "已上架" : "已下架", color: "positive" });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
    await loadGoods();
  }
}

function removeGoods(g: AdminGoods) {
  $q.dialog({
    title: "删除商品",
    message: `确定删除「${g.name}」？`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/goods/delete", { method: "POST", body: { id: g.id } });
      $q.notify({ message: "已删除", color: "positive" });
      await loadGoods();
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

// ================= 蓝牙搜索（双列：低功耗 / 传统） =================
const scanning = ref(false);
const status = ref("");
const statusOk = ref(false);
const bleDevices = ref<BleScanDevice[]>([]);
const sppDevices = ref<SppDevice[]>([]);
let stopBle: (() => Promise<void>) | null = null;

function buildTestPayload() {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  const lines = [
    "--------------------------------",
    "你好 蓝牙小票机",
    `当前时间：${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ` +
      `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`,
    "--------------------------------"
  ];
  // ESC/POS：初始化 + 居中放大 + 文本 + 走纸切纸
  // 文本必须用 GBK 编码：国产热敏机按 GBK 解释字节流，发 UTF-8 会乱码
  const parts: Uint8Array[] = [
    new Uint8Array([0x1b, 0x40]), // ESC @ 初始化
    new Uint8Array([0x1b, 0x61, 0x01]), // ESC a 1 居中
    new Uint8Array([0x1d, 0x21, 0x11]), // GS ! 双倍宽高
    encodeGbk(lines.join("\n")),
    new Uint8Array([0x1b, 0x61, 0x00]), // ESC a 0 左对齐
    new Uint8Array([0x1d, 0x56, 0x41, 0x03]) // GS V A 3 走纸并切纸
  ];
  const total = parts.reduce((n, p) => n + p.length, 0);
  const payload = new Uint8Array(total);
  let offset = 0;
  for (const p of parts) {
    payload.set(p, offset);
    offset += p.length;
  }
  return Array.from(payload);
}

// 点击 BLE 设备 → 打印测试小票
async function printBle(d: BleScanDevice) {
  try {
    const name = await blePrintTo(d.deviceId, buildTestPayload(), d.name);
    statusOk.value = true;
    status.value = `已连接「${name}」，测试小票已发送，请查看打印机出纸`;
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  }
}

// 点击传统蓝牙设备 → 打印测试小票
async function printSpp(d: SppDevice) {
  try {
    const name = await sppPrint(d, buildTestPayload());
    statusOk.value = true;
    status.value = `已连接「${name}」，测试小票已发送，请查看打印机出纸`;
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  }
}

// 开始搜索：BLE 实时回调 + 传统蓝牙 discoverUnpaired（仅未配对设备，不返回已保存的）
async function startScan() {
  status.value = "";
  bleDevices.value = [];
  sppDevices.value = [];
  scanning.value = true;
  try {
    if (isNative()) {
      // 初始化 BLE（触发 Android「附近设备」运行时权限，扫描也依赖它）
      try {
        await bleInit();
      } catch {
        // 已初始化过则忽略
      }
      stopBle = await bleScanStart((d) => {
        if (!bleDevices.value.some((x) => x.deviceId === d.deviceId)) {
          bleDevices.value.push(d);
        }
      });
    } else {
      statusOk.value = true;
      status.value = "当前为网页环境，仅展示传统蓝牙搜索；蓝牙搜索请在 App 内使用";
    }

    // 传统蓝牙扫描（约 12 秒，结束后 resolve）
    // 单独 catch：权限被拒等错误只影响传统蓝牙一列，不打断 BLE 搜索
    try {
      const found = await sppDiscover((d) => {
        if (!sppDevices.value.some((x) => x.address === d.address)) {
          sppDevices.value.push(d);
        }
      });
      if (found.length === 0 && bleDevices.value.length === 0 && status.value === "") {
        statusOk.value = false;
        status.value = "未搜索到设备，请确认打印机已开机并进入配对/广播模式";
      }
    } catch (e) {
      const msg = describeError(e);
      statusOk.value = false;
      status.value = msg.includes("Location permission")
        ? "搜索传统蓝牙需要位置权限：请在系统设置 → 应用 → 权限中允许「位置」后重试（低功耗蓝牙不受影响）"
        : `传统蓝牙搜索失败：${msg}`;
    }
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  } finally {
    scanning.value = false;
    // BLE 扫描由插件持续进行，直到点「停止」或离开页面
  }
}

// 停止搜索
async function stopScan() {
  scanning.value = false;
  const fn = stopBle;
  stopBle = null;
  if (fn) await fn().catch(() => undefined);
}

// ================= 初始化 =================
onMounted(async () => {
  if (!token.value) return;
  try {
    await loadMenus();
    await loadGoods();
  } catch (e) {
    // 401 时 token 已被清理，静默回到登录页；其他错误提示
    if (token.value) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  }
});
</script>

<style scoped lang="scss">
.admin-wrap {
  max-width: 960px;
  margin: 0 auto;
}

.goods-thumb {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  background: #f5f5f5;
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

.scan-list {
  max-height: calc(100vh - 320px);
  overflow-y: auto;
}
</style>
