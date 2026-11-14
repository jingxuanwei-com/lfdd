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
      <q-banner
        v-if="status"
        dense
        rounded
        class="q-mt-sm"
        :class="statusOk ? 'bg-green-1 text-green-9' : 'bg-red-1 text-red-9'"
      >
        {{ status }}
      </q-banner>
    </div>

    <!-- 打印机选择：已配对的经典蓝牙设备 + BLE 搜索 -->
    <q-dialog v-model="picker">
      <q-card style="min-width: 320px">
        <q-card-section class="text-subtitle1">选择打印机</q-card-section>
        <q-list>
          <q-item
            v-for="d in sppDevices"
            :key="d.address"
            v-close-popup
            clickable
            @click="runPrint(() => sppPrint(d, buildTestPayload()), 'SPP')"
          >
            <q-item-section avatar>
              <q-icon name="bluetooth" color="primary" />
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ d.name || "未知设备" }}</q-item-label>
              <q-item-label caption>经典蓝牙 · {{ d.address }}</q-item-label>
            </q-item-section>
          </q-item>
          <q-item
            v-close-popup
            clickable
            @click="runPrint(() => blePrint(buildTestPayload()), 'BLE')"
          >
            <q-item-section avatar>
              <q-icon name="bluetooth_searching" color="primary" />
            </q-item-section>
            <q-item-section>
              <q-item-label>搜索 BLE 设备…</q-item-label>
              <q-item-label caption>低功耗蓝牙打印机</q-item-label>
            </q-item-section>
          </q-item>
        </q-list>
        <q-card-actions align="right">
          <q-btn v-close-popup flat label="取消" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  bleInit,
  blePrint,
  buildTestPayload,
  describeError,
  isNative,
  sppList,
  sppPrint,
  webBluetoothPrint,
  type SppDevice
} from "@/utils/printer";

const printing = ref(false);
const status = ref("");
const statusOk = ref(true);

const picker = ref(false);
const sppDevices = ref<SppDevice[]>([]);

// testPrinter：原生（Capacitor）走 SPP + BLE 插件，桌面/网页走 Web Bluetooth
async function testPrinter() {
  status.value = "";
  if (isNative()) {
    try {
      await bleInit(); // 触发「附近设备」运行时权限，SPP 也依赖这次授权
      const devices = await sppList(); // 已配对的经典蓝牙设备
      if (devices.length > 0) {
        sppDevices.value = devices;
        picker.value = true;
        return;
      }
      await runPrint(() => blePrint(buildTestPayload()), "BLE");
    } catch (e) {
      statusOk.value = false;
      status.value = describeError(e);
    }
    return;
  }
  await runPrint(() => webBluetoothPrint(buildTestPayload()), "");
}

// runPrint：统一执行打印并回填状态
async function runPrint(run: () => Promise<string>, mode: string) {
  status.value = "";
  printing.value = true;
  try {
    const name = await run();
    statusOk.value = true;
    status.value = `已连接${mode ? mode + " " : ""}「${name}」，测试小票已发送，请查看打印机出纸`;
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  } finally {
    printing.value = false;
  }
}
</script>
