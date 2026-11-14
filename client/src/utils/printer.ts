// 蓝牙小票机打印服务
// - 原生（Capacitor Android）：经典蓝牙 SPP + BLE 双通道，走原生插件（WebView 没有 Web Bluetooth API）
// - 桌面 / 网页（Electron、Chrome）：Web Bluetooth（仅 BLE）

// ==================== ESC/POS 指令 ====================
export const ESC_INIT = [0x1b, 0x40]; // ESC @ 初始化打印机
export const ESC_ALIGN_CENTER = [0x1b, 0x61, 0x01]; // ESC a 1 居中对齐
export const ESC_FEED = [0x1b, 0x64, 0x04]; // ESC d 4 走纸 4 行

// 测试文案「你好 蓝牙小票机」的 GBK 字节（TextEncoder 只支持 UTF-8，固定文案直接内联）
const TEST_TEXT_GBK = [
  0xc4, 0xe3, 0xba, 0xc3, 0x20, 0xc0, 0xb6, 0xd1, 0xc0, 0xd0, 0xa1, 0xc6, 0xb1, 0xbb, 0xfa
];

// 组装测试小票数据：初始化 + 居中 + 文案 + 换行 + 走纸
export function buildTestPayload(): number[] {
  return [...ESC_INIT, ...ESC_ALIGN_CENTER, ...TEST_TEXT_GBK, 0x0a, 0x0a, ...ESC_FEED];
}

// ==================== 运行环境 ====================
export function isNative(): boolean {
  return Boolean((window as unknown as { Capacitor?: any }).Capacitor?.isNativePlatform?.());
}

// ==================== 原生：经典蓝牙 SPP ====================
export interface SppDevice {
  address: string;
  name: string;
}

function sppApi(): any {
  const api = (window as unknown as { bluetoothSerial?: any }).bluetoothSerial;
  if (!api) throw new Error("SPP 插件未加载，请执行 npx cap sync 后重启应用");
  return api;
}

// cordova 插件是回调风格，包一层 Promise 并加超时
function sppCall<T>(
  run: (ok: (v: T) => void, fail: (e: unknown) => void) => void,
  timeoutMs = 15000
): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("蓝牙操作超时")), timeoutMs);
    run(
      (v) => {
        clearTimeout(timer);
        resolve(v);
      },
      (e) => {
        clearTimeout(timer);
        reject(new Error(typeof e === "string" ? e : ((e as any)?.message || String(e))));
      }
    );
  });
}

// 已配对的经典蓝牙设备
export function sppList(): Promise<SppDevice[]> {
  return sppCall<SppDevice[]>((ok, fail) => sppApi().list(ok, fail));
}

// 经典蓝牙搜索：扫描附近「未配对」设备（不返回已保存的设备）
// onDevice 每发现一个新设备回调一次；扫描结束（约 12 秒）时 resolve 完整列表
export async function sppDiscover(onDevice?: (d: SppDevice) => void): Promise<SppDevice[]> {
  const api = sppApi();
  const found = new Map<string, SppDevice>();
  const emit = (raw: SppDevice) => {
    if (!raw?.address || found.has(raw.address)) return;
    const d: SppDevice = { address: raw.address, name: raw.name || "" };
    found.set(d.address, d);
    onDevice?.(d);
  };
  api.setDeviceDiscoveredListener(emit);
  try {
    const list = await sppCall<SppDevice[]>((ok, fail) => api.discoverUnpaired(ok, fail), 30000);
    list?.forEach(emit);
    return Array.from(found.values());
  } finally {
    try {
      api.clearDeviceDiscoveredListener();
    } catch {
      // 忽略：部分版本无此方法
    }
  }
}

// 经典蓝牙打印：连接 → 写入 → 断开，返回设备名
export async function sppPrint(device: SppDevice, payload: number[]): Promise<string> {
  const api = sppApi();
  await sppCall<void>((ok, fail) => api.connect(device.address, ok, fail));
  try {
    await sppCall<void>((ok, fail) => api.write(new Uint8Array(payload), ok, fail));
    return device.name || "未知设备";
  } finally {
    // 尽力断开，失败忽略
    sppCall<void>((ok, fail) => api.disconnect(ok, fail), 3000).catch(() => undefined);
  }
}

// ==================== 原生：BLE ====================
interface BleCharacteristic {
  uuid: string;
  properties: { write?: boolean; writeWithoutResponse?: boolean };
}
interface BleService {
  uuid: string;
  characteristics: BleCharacteristic[];
}
interface BleApi {
  initialize(o?: { androidNeverForLocation?: boolean }): Promise<void>;
  requestDevice(o?: Record<string, unknown>): Promise<{ deviceId: string; name?: string }>;
  connect(o: { deviceId: string }): Promise<void>;
  disconnect(o: { deviceId: string }): Promise<void>;
  getServices(o: { deviceId: string }): Promise<{ services: BleService[] }>;
  getMtu(o: { deviceId: string }): Promise<{ value: number }>;
  write(o: { deviceId: string; service: string; characteristic: string; value: string }): Promise<void>;
  writeWithoutResponse(o: {
    deviceId: string;
    service: string;
    characteristic: string;
    value: string;
  }): Promise<void>;
  // 蓝牙搜索
  addListener(event: string, cb: (r: any) => void): Promise<{ remove(): Promise<void> }>;
  requestLEScan(o: Record<string, unknown>): Promise<void>;
  stopLEScan(): Promise<void>;
}

function bleApi(): BleApi {
  const api = (window as unknown as { Capacitor?: any }).Capacitor?.Plugins?.BluetoothLe;
  if (!api) throw new Error("BLE 插件未加载，请执行 npx cap sync 后重启应用");
  return api;
}

// 初始化 BLE：Android 上会弹「附近设备」运行时权限（SPP 也依赖这次授权）
export function bleInit(): Promise<void> {
  return bleApi().initialize({ androidNeverForLocation: true });
}

// 系统设备选择弹窗 → 连接 → 打印
export async function blePrint(payload: number[]): Promise<string> {
  const device = await bleApi().requestDevice({}); // 弹系统设备选择框
  return blePrintTo(device.deviceId, payload, device.name);
}

// 连接指定 BLE 设备并打印（用于搜索列表点击打印），返回设备名
export async function blePrintTo(deviceId: string, payload: number[], name = ""): Promise<string> {
  const api = bleApi();
  await api.connect({ deviceId });
  try {
    const { services } = await api.getServices({ deviceId });
    const target = findWritable(services);
    if (!target) throw new Error("已连接但未找到可写特征，请确认这是 BLE 小票机");

    // 单次写入上限 = MTU - 3（ATT 协议头）
    let mtu = 23;
    try {
      mtu = (await api.getMtu({ deviceId })).value;
    } catch {
      // 取不到就按最小值保守分片
    }
    const chunkSize = Math.max(20, mtu - 3);

    const bytes = new Uint8Array(payload);
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const opts = {
        deviceId,
        service: target.service,
        characteristic: target.characteristic,
        value: toHex(bytes.subarray(i, i + chunkSize)) // 原生插件的 value 要求 hex 字符串
      };
      if (target.useResponse) {
        await api.write(opts);
      } else {
        await api.writeWithoutResponse(opts);
      }
    }
    return name || "未知设备";
  } finally {
    api.disconnect({ deviceId }).catch(() => undefined);
  }
}

// ================= 蓝牙搜索 =================
export interface BleScanDevice {
  deviceId: string;
  name: string;
  rssi?: number;
}

// BLE 搜索：开始扫描附近广播设备（不含已保存配对关系也能发现）
// 返回停止函数；onDevice 每发现一个新设备回调一次
export async function bleScanStart(onDevice: (d: BleScanDevice) => void): Promise<() => Promise<void>> {
  const api = bleApi();
  const seen = new Set<string>();
  const sub = await api.addListener("onScanResult", (r: any) => {
    const dev = r?.device ?? {};
    const id: string = dev.deviceId ?? dev.id ?? "";
    if (!id || seen.has(id)) return;
    seen.add(id);
    onDevice({ deviceId: id, name: r.localName || dev.name || "", rssi: r.rssi });
  });
  try {
    await api.requestLEScan({ services: [], allowDuplicates: false, scanMode: 2 });
  } catch (e) {
    await sub.remove().catch(() => undefined);
    throw e;
  }
  return async () => {
    await sub.remove().catch(() => undefined);
    await api.stopLEScan().catch(() => undefined);
  };
}

// 找第一个具备写能力的特征
function findWritable(services: BleService[]) {
  for (const svc of services) {
    for (const c of svc.characteristics) {
      if (c.properties?.write) return { service: svc.uuid, characteristic: c.uuid, useResponse: true };
      if (c.properties?.writeWithoutResponse)
        return { service: svc.uuid, characteristic: c.uuid, useResponse: false };
    }
  }
  return null;
}

function toHex(bytes: Uint8Array): string {
  let hex = "";
  for (const b of bytes) hex += b.toString(16).padStart(2, "0");
  return hex;
}

// ==================== Web Bluetooth（Electron / Chrome） ====================
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

export async function webBluetoothPrint(payload: number[]): Promise<string> {
  const bluetooth = (navigator as Navigator & { bluetooth?: any }).bluetooth;
  if (!bluetooth) {
    throw new Error("当前环境不支持 Web Bluetooth（需 Chrome/Edge/Chromium 内核）");
  }
  if (!window.isSecureContext) {
    throw new Error("页面不在安全上下文：改用 file:// 或 localhost 加载即可，无需 HTTPS");
  }

  const device = await bluetooth.requestDevice({
    acceptAllDevices: true,
    optionalServices: PRINTER_SERVICES
  });
  const server = await device.gatt.connect();
  try {
    const ch = await findWritableChar(server);
    if (!ch) throw new Error("已连接但未找到可写服务特征，请确认这是 BLE 小票机");

    const bytes = new Uint8Array(payload);
    const CHUNK = 180; // BLE 单包容量有限
    for (let i = 0; i < bytes.length; i += CHUNK) {
      const chunk = bytes.slice(i, i + CHUNK);
      if (ch.writeValueWithResponse) {
        await ch.writeValueWithResponse(chunk);
      } else {
        await ch.writeValue(chunk);
      }
    }
    return device.name || "未知设备";
  } finally {
    server.disconnect();
  }
}

// 遍历所有服务，返回第一个具备写能力的特征
async function findWritableChar(server: any) {
  const services = await server.getPrimaryServices();
  for (const svc of services) {
    const chars = await svc.getCharacteristics();
    for (const c of chars) {
      if (c.properties?.write || c.properties?.writeWithoutResponse) return c;
    }
  }
  return null;
}

// ==================== 错误文案 ====================
export function describeError(e: any): string {
  const name = e?.name;
  const msg = String(e?.message || e || "");
  if (name === "NotFoundError" && /cancel/i.test(msg)) return "已取消选择设备";
  if (name === "NotFoundError") return "未发现蓝牙设备，请确认打印机已开机并进入配对模式";
  if (name === "SecurityError") {
    return "安全限制：请确认页面以 file:// 或 localhost 加载（无需 HTTPS），且由点击按钮触发";
  }
  if (/permission/i.test(msg)) {
    return `蓝牙权限被拒绝，请在系统设置中允许本应用的「附近设备」权限：${msg}`;
  }
  return msg || "测试失败";
}
