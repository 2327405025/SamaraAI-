<template>
  <div class="chat-layout">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">S</div>
        <span class="brand-name">SamaraAI</span>
      </div>
      <nav class="nav-list">
        <button class="nav-item active">图像识别</button>
        <button class="nav-item" @click="$router.push('/ai-chat')">智能对话</button>
      </nav>
      <div class="sidebar-footer">
        <button class="footer-link" @click="$router.push('/menu')">应用中心</button>
      </div>
    </aside>

    <main class="main-panel">
      <header class="main-header">
        <h2>图像识别</h2>
      </header>

      <div class="messages-wrap" ref="chatContainerRef">
        <div v-if="messages.length === 0" class="welcome">
          <div class="welcome-logo">🖼</div>
          <h2>上传图片进行识别</h2>
          <p>支持常见图片格式，AI 将返回分类结果</p>
        </div>

        <div v-else class="messages-inner">
          <div
            v-for="(message, index) in messages"
            :key="index"
            :class="['msg-row', message.role === 'user' ? 'msg-row-user' : 'msg-row-ai']"
          >
            <div class="avatar" :class="message.role">{{ message.role === 'user' ? '你' : 'AI' }}</div>
            <div class="msg-body">
              <div v-if="message.role === 'user'" class="user-content">{{ message.content }}</div>
              <div v-else class="ai-content">{{ message.content }}</div>
              <img v-if="message.imageUrl" :src="message.imageUrl" class="preview-img" alt="preview" />
            </div>
          </div>
        </div>
      </div>

      <div class="composer-wrap">
        <div class="upload-box">
          <input ref="fileInputRef" type="file" accept="image/*" hidden @change="handleFileSelect" />
          <button type="button" class="pick-btn" @click="fileInputRef?.click()">
            {{ selectedFile ? selectedFile.name : '选择图片' }}
          </button>
          <button type="button" class="send-btn-wide" :disabled="!selectedFile" @click="handleSubmit">
            开始识别
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { ref, nextTick } from 'vue'
import api from '../utils/api'

export default {
  name: 'ImageRecognition',
  setup() {
    const messages = ref([])
    const selectedFile = ref(null)
    const fileInputRef = ref()
    const chatContainerRef = ref()

    const handleFileSelect = (e) => {
      selectedFile.value = e.target.files?.[0] || null
    }

    const handleSubmit = async () => {
      if (!selectedFile.value) return

      const file = selectedFile.value
      const imageUrl = URL.createObjectURL(file)

      messages.value.push({
        role: 'user',
        content: `已上传：${file.name}`,
        imageUrl
      })

      await nextTick()
      scrollToBottom()

      const formData = new FormData()
      formData.append('image', file)

      try {
        const response = await api.post('/image/recognize', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        if (response.data?.class_name) {
          messages.value.push({ role: 'assistant', content: `识别结果：${response.data.class_name}` })
        } else {
          const msg = response.data?.status_msg || '识别失败'
          messages.value.push({ role: 'assistant', content: `识别失败：${msg}` })
        }
      } catch (error) {
        messages.value.push({ role: 'assistant', content: `请求失败：${error.message}` })
      } finally {
        URL.revokeObjectURL(imageUrl)
        selectedFile.value = null
        if (fileInputRef.value) fileInputRef.value.value = ''
        await nextTick()
        scrollToBottom()
      }
    }

    const scrollToBottom = () => {
      if (chatContainerRef.value) {
        chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
      }
    }

    return {
      messages,
      selectedFile,
      fileInputRef,
      chatContainerRef,
      handleFileSelect,
      handleSubmit
    }
  }
}
</script>

<style scoped>
.chat-layout {
  display: flex;
  height: 100vh;
  background: var(--ds-bg);
}

.sidebar {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--ds-sidebar);
  border-right: 1px solid var(--ds-border);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 16px;
}

.brand-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--ds-primary);
  color: #fff;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-name {
  font-size: 17px;
  font-weight: 600;
}

.nav-list {
  padding: 8px;
  flex: 1;
}

.nav-item {
  width: 100%;
  text-align: left;
  padding: 10px 12px;
  margin-bottom: 4px;
  border: none;
  border-radius: 8px;
  background: transparent;
  font-size: 14px;
  cursor: pointer;
  color: var(--ds-text);
}

.nav-item:hover {
  background: rgba(0, 0, 0, 0.05);
}

.nav-item.active {
  background: var(--ds-bg);
  font-weight: 500;
  box-shadow: var(--ds-shadow);
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--ds-border);
}

.footer-link {
  width: 100%;
  padding: 8px;
  border: none;
  background: none;
  color: var(--ds-text-secondary);
  font-size: 13px;
  cursor: pointer;
  border-radius: 8px;
}

.main-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.main-header {
  height: 52px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  border-bottom: 1px solid var(--ds-border);
}

.main-header h2 {
  font-size: 16px;
  font-weight: 600;
}

.messages-wrap {
  flex: 1;
  overflow-y: auto;
  padding: 24px 16px;
}

.welcome {
  max-width: 400px;
  margin: 80px auto 0;
  text-align: center;
}

.welcome-logo {
  font-size: 48px;
  margin-bottom: 16px;
}

.welcome h2 {
  font-size: 20px;
  margin-bottom: 8px;
}

.welcome p {
  color: var(--ds-text-secondary);
  font-size: 14px;
}

.messages-inner {
  max-width: 768px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.msg-row {
  display: flex;
  gap: 14px;
  max-width: 768px;
  margin: 0 auto;
  width: 100%;
}

.msg-row-user {
  flex-direction: row-reverse;
}

.avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar.user {
  background: var(--ds-primary);
  color: #fff;
}

.avatar.assistant {
  background: #f0f0f0;
  color: #666;
}

.msg-body {
  flex: 1;
  min-width: 0;
}

.msg-row-user .msg-body {
  text-align: right;
}

.user-content {
  display: inline-block;
  padding: 10px 16px;
  background: var(--ds-user-bubble);
  border-radius: 16px;
  font-size: 15px;
}

.ai-content {
  font-size: 15px;
  line-height: 1.6;
}

.preview-img {
  max-width: 280px;
  margin-top: 10px;
  border-radius: 12px;
  border: 1px solid var(--ds-border);
}

.composer-wrap {
  padding: 16px 24px 24px;
  border-top: 1px solid var(--ds-border);
}

.upload-box {
  max-width: 768px;
  margin: 0 auto;
  display: flex;
  gap: 12px;
}

.pick-btn {
  flex: 1;
  padding: 12px 16px;
  border: 1px dashed var(--ds-border);
  border-radius: 12px;
  background: var(--ds-bg-muted);
  font-size: 14px;
  cursor: pointer;
  text-align: left;
  color: var(--ds-text-secondary);
}

.pick-btn:hover {
  border-color: var(--ds-primary);
}

.send-btn-wide {
  padding: 12px 24px;
  border: none;
  border-radius: 12px;
  background: var(--ds-primary);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.send-btn-wide:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
