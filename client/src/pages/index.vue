<template>
  <!-- 应用布局容器 排序：顶部导航栏、左侧抽屉、页面内容、底部导航栏 -->
  <q-layout view="lHh Lpr lFf">

    
    <!-- 顶部导航栏 -->
    <q-header elevated>
      <!-- 顶部导航栏菜单 -->
      <q-toolbar>
        <q-btn flat dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" />

        <q-toolbar-title> Quasar App </q-toolbar-title>

      </q-toolbar>
    </q-header>

    
    <!-- 左侧抽屉菜单 -->
    <q-drawer v-model="leftDrawerOpen" show-if-above bordered>
      <q-list>
        <q-item-label header> Essential Links </q-item-label>

        <EssentialLink v-for="link in linksList" :key="link.label" v-bind="link" />
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

    <!-- 底部导航栏 -->
    <q-footer elevated>
      <q-tabs v-model="tab" dense class="bg-white text-grey-7" active-color="primary" indicator-color="primary"
        align="justify" narrow-indicator>
        <q-route-tab to="/" name="menu" icon="shopping_bag" label="商品" />
        <q-route-tab to="/cart" name="cart" icon="shopping_cart" label="购物车" />
        <q-route-tab to="/order" name="order" icon="receipt_long" label="订单" />
        <q-route-tab to="/admin" name="admin" icon="admin_panel_settings" label="后台" />
      </q-tabs>
    </q-footer>
  </q-layout>
</template>

<script setup lang="ts">
import { ref } from "vue";
import EssentialLink, {
  type EssentialLinkProps
} from "@/components/EssentialLink.vue";

const linksList: EssentialLinkProps[] = [
  {
    label: "Docs",
    caption: "quasar.dev",
    icon: "school",
    link: "https://quasar.dev"
  },
  {
    label: "GitHub",
    caption: "github.com/quasarframework",
    icon: "code",
    link: "https://github.com/quasarframework"
  },
  {
    label: "Discord Chat Channel",
    caption: "chat.quasar.dev",
    icon: "chat",
    link: "https://chat.quasar.dev"
  },
  {
    label: "Forum",
    caption: "forum.quasar.dev",
    icon: "record_voice_over",
    link: "https://forum.quasar.dev"
  },
  {
    label: "Twitter",
    caption: "@quasarframework",
    icon: "rss_feed",
    link: "https://twitter.quasar.dev"
  },
  {
    label: "Facebook",
    caption: "@QuasarFramework",
    icon: "public",
    link: "https://facebook.quasar.dev"
  },
  {
    label: "Quasar Awesome",
    caption: "Community Quasar projects",
    icon: "favorite",
    link: "https://awesome.quasar.dev"
  }
];

const leftDrawerOpen = ref(false);

// 底部导航当前选中项
const tab = ref("menu");

function toggleLeftDrawer() {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}
</script>
