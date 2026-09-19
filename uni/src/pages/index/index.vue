<template>
  <view class="home-page">
    <view class="banner">
      <text class="banner-title">欢迎光临</text>
      <text class="banner-subtitle">发现好物</text>
    </view>
    <view class="section">
      <text class="section-title">热门推荐</text>
      <view class="goods-grid">
        <view class="goods-item" v-for="item in goodsList" :key="item.id">
          <view class="goods-img">📦</view>
          <text class="goods-name">{{ item.name }}</text>
          <text class="goods-price">¥{{ item.price }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const goodsList = ref([
  { id: 1, name: '商品一', price: '99.00' },
  { id: 2, name: '商品二', price: '199.00' },
  { id: 3, name: '商品三', price: '299.00' },
  { id: 4, name: '商品四', price: '399.00' },
])
</script>

<style>
.home-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.banner {
  background: linear-gradient(135deg, #ff6347, #ff8c69);
  padding: 60rpx 40rpx 80rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.banner-title {
  font-size: 48rpx;
  color: #fff;
  font-weight: bold;
  margin-bottom: 16rpx;
}

.banner-subtitle {
  font-size: 28rpx;
  color: rgba(255, 255, 255, 0.85);
}

.section {
  padding: 30rpx;
  margin-top: -30rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 24rpx;
}

.goods-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.goods-item {
  width: calc(50% - 10rpx);
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  box-sizing: border-box;
}

.goods-img {
  width: 100%;
  height: 200rpx;
  background: #f9f9f9;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64rpx;
  margin-bottom: 16rpx;
}

.goods-name {
  font-size: 28rpx;
  color: #333;
  margin-bottom: 8rpx;
}

.goods-price {
  font-size: 32rpx;
  color: #ff6347;
  font-weight: bold;
}
</style>
