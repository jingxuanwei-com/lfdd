<template>
  <view class="page">
    <!-- 左侧分类侧边栏 -->
    <scroll-view class="side-bar" scroll-y>
      <view
        v-for="cat in categories"
        :key="cat.id"
        class="side-item"
        :class="{ active: currentCatId === cat.id }"
        @tap="switchCat(cat.id)"
      >
        <text class="side-text">{{ cat.name }}</text>
      </view>
    </scroll-view>

    <!-- 右侧双排商品列表 -->
    <scroll-view
      class="goods-panel"
      scroll-y
      :scroll-into-view="scrollTarget"
      :scroll-with-animation="true"
      @scroll="onScroll"
    >
      <view v-for="cat in categories" :key="cat.id" :id="'cat-' + cat.id" class="cat-section">
        <!-- 分类标题 -->
        <view class="cat-header">
          <text class="cat-title">{{ cat.name }}</text>
        </view>
        <!-- 双排商品 -->
        <view class="goods-grid">
          <view
            v-for="item in cat.goods"
            :key="item.id"
            class="goods-card"
            @tap="onGoodsTap(item)"
          >
            <view class="goods-thumb">
              <text class="goods-emoji">{{ item.icon }}</text>
            </view>
            <view class="goods-body">
              <text class="goods-name">{{ item.name }}</text>
              <text class="goods-desc">{{ item.desc }}</text>
              <view class="goods-footer">
                <text class="goods-price">¥{{ item.price }}</text>
                <view class="btn-add" @tap.stop="addToCart(item)">
                  <text class="btn-text">+</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- 底部购物车浮条 -->
    <view class="cart-bar" v-if="cartTotal > 0">
      <view class="cart-left">
        <view class="cart-badge-wrap">
          <text class="cart-icon">🛒</text>
          <view class="cart-badge"><text class="badge-num">{{ cartTotal }}</text></view>
        </view>
        <text class="cart-price">¥{{ cartPrice }}</text>
      </view>
      <view class="cart-btn" @tap="goCheckout">
        <text class="cart-btn-text">去结算</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'

// -------- 数据 --------
const categories = ref([
  {
    id: 1, name: '热销推荐',
    goods: [
      { id: 101, name: '招牌炒饭', icon: '🍚', desc: '大火爆炒 粒粒分明', price: 18, count: 0 },
      { id: 102, name: '红烧牛肉面', icon: '🍜', desc: '浓郁汤底 大块牛肉', price: 22, count: 0 },
      { id: 103, name: '香煎鸡排', icon: '🍗', desc: '外酥里嫩 鲜嫩多汁', price: 25, count: 0 },
      { id: 104, name: '番茄蛋汤', icon: '🍲', desc: '酸甜开胃 营养丰富', price: 10, count: 0 },
    ],
  },
  {
    id: 2, name: '经典套餐',
    goods: [
      { id: 201, name: '商务A套餐', icon: '🍱', desc: '两荤一素 含汤', price: 35, count: 0 },
      { id: 202, name: '商务B套餐', icon: '🍱', desc: '三荤一素 含汤', price: 38, count: 0 },
      { id: 203, name: '学生套餐', icon: '🍱', desc: '一荤两素 含饭', price: 25, count: 0 },
      { id: 204, name: '家庭套餐', icon: '🍱', desc: '四荤两素 3-4人', price: 88, count: 0 },
    ],
  },
  {
    id: 3, name: '凉菜小食',
    goods: [
      { id: 301, name: '凉拌黄瓜', icon: '🥒', desc: '清脆爽口', price: 8, count: 0 },
      { id: 302, name: '皮蛋豆腐', icon: '🥚', desc: '经典凉菜', price: 12, count: 0 },
      { id: 303, name: '口水鸡', icon: '🐔', desc: '麻辣鲜香', price: 18, count: 0 },
      { id: 304, name: '酸辣土豆丝', icon: '🥔', desc: '酸辣开胃', price: 10, count: 0 },
    ],
  },
  {
    id: 4, name: '饮品甜点',
    goods: [
      { id: 401, name: '柠檬水', icon: '🍋', desc: '冰镇解渴', price: 6, count: 0 },
      { id: 402, name: '奶茶', icon: '🧋', desc: '香浓丝滑', price: 12, count: 0 },
      { id: 403, name: '双皮奶', icon: '🍮', desc: '顺德风味', price: 15, count: 0 },
      { id: 404, name: '红豆冰沙', icon: '🍧', desc: '冰爽甜蜜', price: 14, count: 0 },
    ],
  },
])

// -------- 状态 --------
const currentCatId = ref(categories.value[0].id)
const scrollTarget = ref('')
const _scrollLock = ref(false)

// -------- 计算 --------
const cartTotal = computed(() =>
  categories.value.reduce((s, c) => s + c.goods.reduce((g, i) => g + i.count, 0), 0)
)
const cartPrice = computed(() =>
  categories.value
    .reduce((s, c) => s + c.goods.reduce((g, i) => g + i.price * i.count, 0), 0)
    .toFixed(2)
)

// -------- 方法 --------
function switchCat(id: number) {
  _scrollLock.value = true
  currentCatId.value = id
  scrollTarget.value = ''
  nextTick(() => { scrollTarget.value = 'cat-' + id })
  setTimeout(() => { _scrollLock.value = false }, 400)
}

function onScroll(e: any) {
  if (_scrollLock.value) return
  const sys = uni.getSystemInfoSync()
  const query = uni.createSelectorQuery().in(this)
  // 简易方案：根据 scrollTop 区间判断
  // uni-app 中用 createSelectorQuery 需要组件上下文，这里用节流简单判断
  const top = e.detail.scrollTop
  const sectionHeight = (sys.windowHeight || 600) * 0.6 // 粗略
  const idx = Math.floor(top / sectionHeight)
  const cat = categories.value[idx]
  if (cat) currentCatId.value = cat.id
}

function addToCart(item: any) {
  item.count++
}

function goCheckout() {
  uni.showToast({ title: '去结算', icon: 'none' })
}

function onGoodsTap(item: any) {
  if (item.count === 0) item.count++
}
</script>

<style>
/* ===== 页面 ===== */
.page {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
  overflow: hidden;
}

/* ===== 左侧分类栏 ===== */
.side-bar {
  width: 180rpx;
  height: 100%;
  background: #f3f3f3;
  flex-shrink: 0;
}

.side-item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100rpx;
  border-left: 6rpx solid transparent;
}

.side-item.active {
  background: #fff;
  border-left-color: #ff6347;
}

.side-item.active .side-text {
  color: #ff6347;
  font-weight: bold;
}

.side-text {
  font-size: 26rpx;
  color: #666;
}

/* ===== 右侧商品区 ===== */
.goods-panel {
  flex: 1;
  height: 100%;
  background: #fff;
}

.cat-section {
  padding-bottom: 16rpx;
}

.cat-header {
  padding: 20rpx 20rpx 10rpx;
}

.cat-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

/* ===== 双排商品 ===== */
.goods-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 0 12rpx;
  gap: 12rpx;
}

.goods-card {
  width: calc(50% - 6rpx);
  background: #fafafa;
  border-radius: 16rpx;
  overflow: hidden;
}

.goods-thumb {
  width: 100%;
  height: 160rpx;
  background: linear-gradient(135deg, #fff5f3, #ffe8e3);
  display: flex;
  align-items: center;
  justify-content: center;
}

.goods-emoji {
  font-size: 68rpx;
}

.goods-body {
  padding: 12rpx 16rpx 16rpx;
}

.goods-name {
  font-size: 26rpx;
  color: #333;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.goods-desc {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.goods-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12rpx;
}

.goods-price {
  font-size: 30rpx;
  color: #ff6347;
  font-weight: bold;
}

.btn-add {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: #ff6347;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-text {
  color: #fff;
  font-size: 32rpx;
  line-height: 1;
}

/* ===== 底部购物车 ===== */
.cart-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #2d2d2d;
  display: flex;
  align-items: center;
  padding: 0 24rpx;
  z-index: 99;
}

.cart-left {
  flex: 1;
  display: flex;
  align-items: center;
}

.cart-badge-wrap {
  position: relative;
  width: 80rpx;
  height: 80rpx;
  background: #3d3d3d;
  border-radius: 50%;
  margin-top: -24rpx;
  border: 4rpx solid #2d2d2d;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cart-icon {
  font-size: 36rpx;
}

.cart-badge {
  position: absolute;
  top: -6rpx;
  right: -6rpx;
  background: #ff6347;
  border-radius: 20rpx;
  min-width: 28rpx;
  height: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6rpx;
}

.badge-num {
  font-size: 18rpx;
  color: #fff;
}

.cart-price {
  font-size: 34rpx;
  color: #fff;
  font-weight: bold;
  margin-left: 16rpx;
}

.cart-btn {
  background: #ff6347;
  height: 72rpx;
  padding: 0 36rpx;
  border-radius: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cart-btn-text {
  font-size: 28rpx;
  color: #fff;
  font-weight: bold;
}
</style>
