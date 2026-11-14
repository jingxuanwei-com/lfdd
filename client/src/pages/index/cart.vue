<template>
  <q-page class="cart-page column" :style-fn="pageStyle">
    <!-- 标题栏 -->
    <div class="row items-center justify-between q-px-md q-py-sm">
      <span class="text-subtitle1 text-weight-medium">购物车</span>
      <q-btn
        v-if="cart.length > 0"
        flat
        dense
        size="sm"
        color="negative"
        icon="delete_sweep"
        label="清空"
        @click="clearCart"
      />
    </div>

    <!-- 空态 -->
    <div v-if="cart.length === 0" class="column items-center justify-center empty-box">
      <q-icon name="shopping_cart" size="64px" color="grey-4" />
      <div class="text-h6 text-grey-7 q-mt-md">购物车是空的</div>
      <q-btn class="q-mt-md" color="primary" label="去点单" to="/" />
    </div>

    <!-- 明细列表（带图片） -->
    <div v-else class="cart-list">
      <div v-for="line in cart" :key="line.id" class="cart-line row items-center">
        <div class="cart-thumb">
          <img v-if="line.image" :src="imgSrc(line.image)" :alt="line.name" />
          <q-icon v-else name="shopping_bag" size="28px" color="grey-5" />
        </div>
        <div class="col">
          <div class="text-weight-medium ellipsis">{{ line.name }}</div>
          <div v-if="line.spec" class="text-caption text-grey-7 ellipsis">{{ line.spec }}</div>
          <div class="text-negative text-weight-bold q-mt-xs">
            ¥{{ fmtPrice(line.price) }}
            <span class="text-caption text-grey-6 text-weight-regular"> / 件</span>
          </div>
        </div>
        <div class="column items-end q-gutter-xs">
          <div class="row items-center q-gutter-sm">
            <q-btn
              round
              dense
              color="primary"
              icon="remove"
              size="md"
              aria-label="减少"
              @click="changeQty(line, -1)"
            />
            <span class="qty-num text-weight-bold">{{ line.qty }}</span>
            <q-btn
              round
              dense
              color="primary"
              icon="add"
              size="md"
              aria-label="增加"
              @click="changeQty(line, 1)"
            />
          </div>
          <div class="text-caption text-grey-6">小计 ¥{{ fmtPrice(line.price * line.qty) }}</div>
        </div>
      </div>
    </div>

    <!-- 底部合计条 -->
    <div v-if="cart.length > 0" class="cart-summary row items-center q-px-md">
      <div class="col">
        <div class="text-caption text-grey-7" style="line-height: 1">共 {{ cartCount }} 件</div>
        <div class="text-h6 text-negative text-weight-bold" style="line-height: 1.3">
          ¥{{ fmtPrice(cartTotal) }}
        </div>
      </div>
      <q-btn color="primary" label="去结算" unelevated @click="checkout" />
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useQuasar } from "quasar";
import {
  cart,
  cartCount,
  cartTotal,
  changeQty,
  clearCart,
  fmtPrice,
  imgSrc
} from "../../utils/cart";

const $q = useQuasar();

// QPage styleFn：header/footer 占位后剩余高度，让页面撑满、列表区各自滚动
function pageStyle(offset: number, height: number) {
  if (height > 0) {
    return { height: `${height - offset}px` };
  }
  return { height: offset !== 0 ? `calc(100vh - ${offset}px)` : "100vh" };
}

function checkout() {
  // TODO: 接订单后端（下单 → 后台查看/打印订单）
  $q.notify({ message: "结算下单功能待开发", color: "primary" });
}
</script>

<style scoped lang="scss">
.cart-page {
  overflow: hidden;
  background: #fff;
}

.empty-box {
  flex: 1;
}

.cart-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0 16px;
}

.cart-line {
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }
}

.cart-thumb {
  width: 64px;
  height: 64px;
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

.qty-num {
  min-width: 24px;
  text-align: center;
  font-size: 16px;
}

.cart-summary {
  gap: 12px;
  padding-top: 10px;
  padding-bottom: 12px;
  border-top: 1px solid #f0f0f0;
  background: #fff;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.06);
}
</style>
