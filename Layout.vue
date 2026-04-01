<template>
  <el-container class="layout-container">
    <!-- 顶部导航栏 -->
    <el-header class="layout-header">
      <div class="logo">
        <span>My App</span>
      </div>
      <el-menu
        mode="horizontal"
        :ellipsis="false"
        class="layout-menu"
        :default-active="activeMenu"
        @select="handleMenuSelect"
      >
        <el-menu-item index="home">首页</el-menu-item>
        <el-menu-item index="about">关于</el-menu-item>
        <el-menu-item index="settings">设置</el-menu-item>
      </el-menu>
      <div class="user-actions">
        <el-avatar :size="32" src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" />
      </div>
    </el-header>

    <!-- 主体内容区域 -->
    <el-main class="layout-main">
      <router-view />
    </el-main>

    <!-- 底部页脚 -->
    <el-footer class="layout-footer">
      <span>© 2023 My Company. All rights reserved.</span>
    </el-footer>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// 当前激活的菜单项，根据路由路径动态计算
const activeMenu = computed(() => {
  const path = route.path
  if (path === '/') return 'home'
  if (path.startsWith('/about')) return 'about'
  if (path.startsWith('/settings')) return 'settings'
  return 'home'
})

// 菜单选择处理函数
const handleMenuSelect = (index) => {
  const pathMap = {
    home: '/',
    about: '/about',
    settings: '/settings'
  }
  if (pathMap[index]) {
    // 这里假设使用了 vue-router，实际项目中可能需要引入 useRouter
    // 为了演示布局组件，这里仅做逻辑示意
    console.log(`Navigating to: ${pathMap[index]}`)
    // window.location.href = pathMap[index] // 简单跳转示例
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  margin-right: 20px;
}

.layout-menu {
  flex: 1;
  border-bottom: none;
}

.user-actions {
  margin-left: 20px;
}

.layout-main {
  flex: 1;
  padding: 20px;
  background-color: #f5f7fa;
  overflow-y: auto;
}

.layout-footer {
  text-align: center;
  padding: 15px 0;
  background-color: #fff;
  color: #909399;
  font-size: 14px;
  border-top: 1px solid #ebeef5;
}
</style>
