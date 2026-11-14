<template>
  <q-page class="flex flex-center">
    <div class="column items-center q-gutter-sm" style="min-width: 360px">
      <q-icon name="admin_panel_settings" size="64px" color="primary" />
      <div class="text-h5">后台</div>
      <div class="text-grey-7">菜单 / 商品管理（对接 /api/menu、/api/goods 管理接口）</div>

      <!-- 小票机连接测试 -->
      <q-btn
        class="q-mt-lg"
        color="primary"
        icon="print"
        label="测试小票机（打印「你好 蓝牙小票机」）"
        :loading="printing"
        @click="testPrinter"
      />
      <q-banner v-if="status" dense rounded class="q-mt-sm" :class="statusOk ? 'bg-green-1 text-green-9' : 'bg-red-1 text-red-9'">
        {{ status }}
      </q-banner>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref } from "vue";

// ================= 小票机连接测试 =================
// 协议：ESC/POS（热敏小票机常规协议）
// 中文编码：GBK（小票机默认编码；浏览器 TextEncoder 只支持 UTF-8，
// 测试文案固定，故直接内联 GBK 字节：你好 蓝牙小票机）
const TEST_TEXT_GBK = [
  0xc4, 0xe3, 0xba, 0xc3, 0x20, 0xc0, 0xb6, 0xd1, 0xc0, 0xd0, 0xa1, 0xc6, 0xb1, 0xbb, 0xfa
];

// ESC/POS 指令：ESC @ 初始化，ESC a 1 居中对齐，ESC d 4 走纸 4 行
const ESC_INIT = [0x1b, 0x40];
const ESC_ALIGN_CENTER = [0x1b, 0x61, 0x01];
const ESC_FEED = [0x1b, 0x64, 0x04];

// 常见蓝牙小票机 GATT 服务 UUID（不同厂商差异大，尽量多列）
const PRINTER_SERVICES = [
  0xff00,
  0x18f0,
  0xffe0,
  0xfff0,
  0xffe5,
  "0000ff00-0000-1000-8000-00805f9b34fb",
  "000018f0-0000-1000-8000-00805f9b34fb",
  "0000ffe0-0000-1000-8000-00805f9b34fb",
  "0000fff0-0000-1000-8000-00805f9b34fb",
  "49535343-fe7d-4ae5-8fa9-9fafd205e455"
];

const printing = ref(false);
const status = ref("");
const statusOk = ref(true);

// testPrinter 选择蓝牙设备 → 连接 GATT → 找可写特征 → 发送 ESC/POS 数据
async function testPrinter() {
  status.value = "";
  const bluetooth = (navigator as Navigator & { bluetooth?: any }).bluetooth;
  if (!bluetooth) {
    statusOk.value = false;
    status.value = "当前环境不支持 Web Bluetooth（需 Chrome/Edge/Chromium 内核 WebView）";
    return;
  }
  if (!window.isSecureContext) {
    statusOk.value = false;
    status.value = "页面不在安全上下文：改用 file:// 或 localhost 加载即可，无需 HTTPS";
    return;
  }

  printing.value = true;
  try {
    // 1. 弹出系统蓝牙设备列表，让用户选择小票机
    const device = await bluetooth.requestDevice({
      acceptAllDevices: true,
      optionalServices: PRINTER_SERVICES
    });

    // 2. 连接 GATT
    const server = await device.gatt.connect();

    // 3. 遍历服务，找第一个可写特征（打印数据都写到这里）
    const ch = await findWritableChar(server);
    if (!ch) {
      throw new Error("已连接但未找到可写服务特征，请确认这是 BLE 小票机");
    }

    // 4. 组装 ESC/POS 数据：初始化 + 居中 + 测试文案 + 换行走纸
    const payload = [...ESC_INIT, ...ESC_ALIGN_CENTER, ...TEST_TEXT_GBK, 0x0a, 0x0a, ...ESC_FEED];

    // 5. 分片写入（BLE 单包容量有限，180 字节一片）
    const CHUNK = 180;
    for (let i = 0; i < payload.length; i += CHUNK) {
      const chunk = payload.slice(i, i + CHUNK);
      if (ch.writeValueWithResponse) {
        await ch.writeValueWithResponse(chunk);
      } else {
        await ch.writeValue(chunk);
      }
    }

    statusOk.value = true;
    status.value = `已连接「${device.name || "未知设备"}」，测试小票已发送，请查看打印机出纸`;
    device.gatt.disconnect();
  } catch (e: any) {
    statusOk.value = false;
    if (e?.name === "NotFoundError" && /cancel/i.test(e?.message || "")) {
      status.value = "已取消选择设备";
    } else if (e?.name === "NotFoundError") {
      status.value = "未发现蓝牙设备，请确认打印机已开机并进入配对模式";
    } else if (e?.name === "SecurityError") {
      status.value = "安全限制：请确认页面以 file:// 或 localhost 加载（无需 HTTPS），且点击按钮触发";
    } else {
      status.value = `测试失败：${e?.message || e}`;
    }
  } finally {
    printing.value = false;
  }
}

// findWritableChar 遍历所有服务，返回第一个具备写能力的特征
async function findWritableChar(server: any) {
  const services = await server.getPrimaryServices();
  for (const svc of services) {
    const chars = await svc.getCharacteristics();
    for (const c of chars) {
      if (c.properties?.write || c.properties?.writeWithoutResponse) {
        return c;
      }
    }
  }
  return null;
}
</script>
