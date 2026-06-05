<template>
  <div class="hub-page">
    <header class="hub-header">
      <div class="hub-brand">
        <span class="hub-logo">S</span>
        <span>SamaraAI</span>
      </div>
      <button class="hub-logout" @click="handleLogout">退出</button>
    </header>

    <main class="hub-main">
      <h1 class="hub-title">选择功能</h1>
      <p class="hub-sub">智能对话与视觉识别，一站式 AI 工作台</p>

      <div class="hub-grid">
        <article class="hub-card" @click="$router.push('/ai-chat')">
          <div class="card-icon chat">💬</div>
          <h3>智能对话</h3>
          <p>流式聊天 · RAG 文档 · 多模型切换</p>
          <span class="card-arrow">→</span>
        </article>
        <article class="hub-card" @click="$router.push('/image-recognition')">
          <div class="card-icon vision">🖼</div>
          <h3>图像识别</h3>
          <p>上传图片，获取 AI 分类识别结果</p>
          <span class="card-arrow">→</span>
        </article>
      </div>
    </main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'MenuView',
  setup() {
    const router = useRouter()

    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm('确定退出登录？', '提示', {
          confirmButtonText: '退出',
          cancelButtonText: '取消',
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success('已退出')
        router.push('/login')
      } catch {
        /* cancelled */
      }
    }

    return { handleLogout }
  }
}
</script>

<style scoped>
.hub-page {
  min-height: 100vh;
  background: var(--ds-bg-muted);
}

.hub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 32px;
  background: var(--ds-bg);
  border-bottom: 1px solid var(--ds-border);
}

.hub-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
}

.hub-logo {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--ds-primary);
  color: #fff;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
}

.hub-logout {
  padding: 8px 16px;
  border: 1px solid var(--ds-border);
  border-radius: 8px;
  background: var(--ds-bg);
  font-size: 14px;
  cursor: pointer;
  color: var(--ds-text-secondary);
}

.hub-logout:hover {
  border-color: #f56c6c;
  color: #f56c6c;
}

.hub-main {
  max-width: 720px;
  margin: 0 auto;
  padding: 48px 24px 64px;
  text-align: center;
}

.hub-title {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 8px;
}

.hub-sub {
  font-size: 15px;
  color: var(--ds-text-secondary);
  margin-bottom: 40px;
}

.hub-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  text-align: left;
}

.hub-card {
  position: relative;
  padding: 28px 24px;
  background: var(--ds-bg);
  border: 1px solid var(--ds-border);
  border-radius: 16px;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.2s;
}

.hub-card:hover {
  border-color: var(--ds-primary);
  box-shadow: 0 8px 24px rgba(77, 107, 254, 0.1);
  transform: translateY(-2px);
}

.card-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  margin-bottom: 16px;
}

.card-icon.chat {
  background: #eef3fe;
}

.card-icon.vision {
  background: #f0fdf4;
}

.hub-card h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.hub-card p {
  font-size: 14px;
  color: var(--ds-text-secondary);
  line-height: 1.5;
}

.card-arrow {
  position: absolute;
  top: 28px;
  right: 24px;
  font-size: 18px;
  color: var(--ds-text-secondary);
  transition: transform 0.2s;
}

.hub-card:hover .card-arrow {
  transform: translateX(4px);
  color: var(--ds-primary);
}
</style>
