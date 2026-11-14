import { computed, ref } from "vue";

// 购物车共享状态（模块级：商品页加购 → 购物车页查看，跨路由保留）
export interface CartLine {
  id: string;
  name: string;
  price: number; // 单位：分
  stock: number; // -1 不限，0 售罄
  image: string; // 哈希或完整 URL
  spec: string;
  qty: number;
}

export const cart = ref<CartLine[]>([]);

export const cartCount = computed(() => cart.value.reduce((s, l) => s + l.qty, 0));
export const cartTotal = computed(() => cart.value.reduce((s, l) => s + l.price * l.qty, 0));

export function lineQty(id: string): number {
  return cart.value.find((l) => l.id === id)?.qty ?? 0;
}

// 增减数量：delta=1 加，-1 减；受库存限制（stock=-1 不限），减到 0 自动移出行
export function changeQty(
  item: { id: string; name: string; price: number; stock: number; image?: string; spec?: string },
  delta: number
) {
  const idx = cart.value.findIndex((l) => l.id === item.id);
  if (delta > 0) {
    if (idx >= 0) {
      const line = cart.value[idx]!;
      if (item.stock > 0 && line.qty >= item.stock) return; // 超库存忽略
      line.qty += 1;
    } else {
      cart.value.push({ ...item, image: item.image ?? "", spec: item.spec ?? "", qty: 1 });
    }
  } else if (idx >= 0) {
    const line = cart.value[idx]!;
    line.qty -= 1;
    if (line.qty <= 0) cart.value.splice(idx, 1);
  }
}

export function clearCart() {
  cart.value = [];
}

export function fmtPrice(price: number): string {
  return (price / 100).toFixed(2);
}

// image 字段只存哈希，展示时拼成 /api/upload/{hash}（兼容旧数据里的完整 URL/路径）
export function imgSrc(v: string): string {
  if (!v) return "";
  if (v.startsWith("/") || /^https?:\/\//.test(v)) return v;
  return `/api/upload/${v}`;
}
