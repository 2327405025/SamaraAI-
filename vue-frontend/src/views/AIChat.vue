<template>
  <div class="chat-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">S</div>
        <span class="brand-name">SamaraAI</span>
      </div>

      <button class="btn-new-chat" @click="createNewSession">
        <span class="icon-plus">+</span>
        开启新对话
      </button>

      <div class="session-scroll">
        <p v-if="sessions.length === 0" class="session-empty">暂无历史对话</p>
        <div
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id && !tempSession }]"
        >
          <button class="session-btn" @click="switchSession(session.id)">
            <span class="session-title">{{ session.name || '新对话' }}</span>
          </button>
          <button
            class="session-delete"
            type="button"
            title="删除对话"
            @click.stop="deleteSession(session.id)"
          >
            ×
          </button>
        </div>
      </div>

      <div class="sidebar-footer">
        <button class="footer-link" @click="$router.push('/menu')">应用中心</button>
      </div>
    </aside>

    <!-- 主区域 -->
    <main class="main-panel">
      <header class="main-header">
        <div class="header-left">
          <select v-model="selectedModel" class="model-picker">
            <option value="1">DeepSeek 兼容 · 百炼</option>
            <option value="2">RAG 知识库</option>
            <option value="3">MCP 工具</option>
          </select>
          <span class="mode-hint">{{ modeHint }}</span>
        </div>
        <button
          class="icon-btn"
          title="上传 RAG 知识库文档（.md / .txt，不影响当前对话模式）"
          :disabled="uploading"
          @click="triggerFileUpload"
        >
          {{ uploading ? '…' : '📎' }}
        </button>
        <input
          ref="fileInput"
          type="file"
          accept=".md,.txt,text/markdown,text/plain"
          hidden
          @change="handleFileUpload"
        />
      </header>

      <div v-if="uploadedFiles.length > 0" class="rag-files-bar">
        <span class="rag-files-label">知识库文档</span>
        <div class="rag-files-list">
          <span v-for="file in uploadedFiles" :key="file.fileId" class="rag-file-chip" :title="file.fileName">
            <span class="rag-file-name">📄 {{ file.fileName }}</span>
            <button
              type="button"
              class="rag-file-delete"
              title="删除文档"
              @click="deleteRagFile(file)"
            >
              ×
            </button>
          </span>
        </div>
        <button
          v-if="selectedModel !== '2'"
          type="button"
          class="rag-switch-btn"
          @click="selectedModel = '2'"
        >
          用 RAG 提问
        </button>
        <span v-else class="rag-files-note">当前 RAG 模式会使用以上文档</span>
      </div>

      <div class="messages-wrap" ref="messagesRef">
        <!-- 空状态 -->
        <div v-if="currentMessages.length === 0" class="welcome">
          <div class="welcome-logo">S</div>
          <h2>我是 SamaraAI，有什么可以帮你？</h2>
          <p>选择模型后直接输入问题，支持流式回复与文档 RAG</p>
          <div class="welcome-hints">
            <button @click="inputMessage = '帮我写一段 Python 快速排序'">写一段代码</button>
            <button @click="inputMessage = '解释一下什么是 RAG'">什么是 RAG</button>
            <button @click="inputMessage = '总结这篇文档的要点'">文档总结</button>
          </div>
        </div>

        <!-- 消息列表 -->
        <div v-else class="messages-inner">
          <div
            v-for="message in currentMessages"
            :key="message.id"
            :class="['msg-row', message.role === 'user' ? 'msg-row-user' : 'msg-row-ai']"
          >
            <div class="avatar" :class="message.role">
              {{ message.role === 'user' ? '你' : 'AI' }}
            </div>
            <div class="msg-body">
              <div
                v-if="message.role === 'assistant'"
                :key="`${message.id}-${message.renderKey ?? (message.meta?.status === 'streaming' ? 's' : 'd')}`"
                class="msg-content ai-content"
              >
                <template v-if="message.meta?.status === 'streaming'">
                  <span class="streaming-text">{{ message.content }}</span>
                  <span class="typing">▍</span>
                </template>
                <div v-else class="ai-rendered" v-html="renderMarkdown(message.content)" />
              </div>
              <div v-else class="msg-content user-content">{{ message.content }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="composer-wrap">
        <div class="composer">
          <textarea
            ref="messageInput"
            v-model="inputMessage"
            placeholder="给 SamaraAI 发送消息"
            rows="1"
            :disabled="loading"
            @keydown.enter.exact.prevent="sendMessage"
            @input="autoResize"
          />
          <button
            class="send-btn"
            :disabled="!inputMessage.trim() || loading"
            @click="sendMessage"
          >
            <svg v-if="!loading" width="20" height="20" viewBox="0 0 24 24" fill="none">
              <path d="M12 19V5M12 5L5 12M12 5L19 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span v-else class="spinner" />
          </button>
        </div>
        <p class="composer-tip">{{ composerTip }}</p>
      </div>
    </main>
  </div>
</template>

<script>
import { ref, nextTick, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../utils/api'
import { consumeChatSSE, getStreamChatUrl, renderSimpleMarkdown } from '../utils/streamChat'

let messageSeq = 0
function createMessage(role, content, meta) {
  return {
    id: `msg-${Date.now()}-${++messageSeq}`,
    role,
    content,
    meta
  }
}

export default {
  name: 'AIChat',
  setup() {
    const sessions = ref({})
    const currentSessionId = ref(null)
    const tempSession = ref(false)
    const currentMessages = ref([])
    const inputMessage = ref('')
    const loading = ref(false)
    const messagesRef = ref(null)
    const messageInput = ref(null)
    const selectedModel = ref('1')
    const uploading = ref(false)
    const uploadedFiles = ref([])
    const fileInput = ref(null)

    const renderMarkdown = renderSimpleMarkdown

    const modeHints = {
      1: '普通对话，不读取知识库文档',
      2: '基于已上传文档回答',
      3: '调用 MCP 工具（如天气），不读取文档'
    }

    const modeHint = computed(() => modeHints[selectedModel.value] || '')

    const composerTip = computed(() => {
      if (selectedModel.value === '2') {
        return uploadedFiles.value.length > 0
          ? 'Enter 发送 · RAG 模式将检索上方知识库文档'
          : 'Enter 发送 · 请先用 📎 上传 .md / .txt 文档'
      }
      if (selectedModel.value === '3') {
        return 'Enter 发送 · MCP 模式可查询天气等外部信息'
      }
      return 'Enter 发送 · 内容由 AI 生成，请仔细甄别'
    })

    const autoResize = (e) => {
      const el = e.target
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 100) + 'px'
    }

    /** 同一条消息在 currentMessages 与 session 缓存里保持同一引用/内容 */
    const syncMessage = (messageId, updater) => {
      const ci = currentMessages.value.findIndex(m => m.id === messageId)
      if (ci === -1) return
      const next = typeof updater === 'function'
        ? updater({ ...currentMessages.value[ci] })
        : { ...currentMessages.value[ci], ...updater }
      currentMessages.value[ci] = next

      const sid = currentSessionId.value
      const sessMsgs = sessions.value[sid]?.messages
      if (sessMsgs) {
        const si = sessMsgs.findIndex(m => m.id === messageId)
        if (si !== -1) {
          sessMsgs[si] = next
        }
      }
    }

    const finalizeAssistantMessage = (index) => {
      const msg = currentMessages.value[index]
      if (!msg) return
      syncMessage(msg.id, (m) => ({
        id: m.id,
        role: 'assistant',
        content: m.content,
        renderKey: Date.now()
      }))
    }

    const appendAssistantChunk = (index, chunk) => {
      const msg = currentMessages.value[index]
      if (!msg) return
      syncMessage(msg.id, (m) => ({ ...m, content: m.content + chunk }))
    }

    const loadUploadedFiles = async () => {
      try {
        const response = await api.get('/file/list')
        if (response.data?.status_code === 1000 && Array.isArray(response.data.files)) {
          uploadedFiles.value = response.data.files
        }
      } catch (error) {
        console.error('Load uploaded files error:', error)
      }
    }

    const loadSessions = async () => {
      try {
        const response = await api.get('/AI/chat/sessions')
        if (response.data?.status_code === 1000 && Array.isArray(response.data.sessions)) {
          const sessionMap = {}
          response.data.sessions.forEach(s => {
            const sid = String(s.sessionId)
            sessionMap[sid] = {
              id: sid,
              name: s.name || `对话 ${sid.slice(0, 8)}`,
              messages: []
            }
          })
          sessions.value = sessionMap
        }
      } catch (error) {
        console.error('Load sessions error:', error)
      }
    }

    const createNewSession = () => {
      currentSessionId.value = 'temp'
      tempSession.value = true
      currentMessages.value = []
      nextTick(() => messageInput.value?.focus())
    }

    const switchSession = async (sessionId) => {
      if (!sessionId) return
      currentSessionId.value = String(sessionId)
      tempSession.value = false

      if (!sessions.value[sessionId]?.messages?.length) {
        try {
          const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
          if (response.data?.status_code === 1000 && Array.isArray(response.data.history)) {
            sessions.value[sessionId].messages = response.data.history.map(item => createMessage(
              item.is_user ? 'user' : 'assistant',
              item.content
            ))
          }
        } catch (err) {
          console.error('Load history error:', err)
        }
      }

      currentMessages.value = [...(sessions.value[sessionId]?.messages || [])]
      await nextTick()
      scrollToBottom()
    }

    const deleteSession = async (sessionId) => {
      if (!sessionId || loading.value) return
      try {
        await ElMessageBox.confirm('确定删除该对话？删除后无法恢复。', '删除对话', {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        })
      } catch {
        return
      }

      try {
        const response = await api.post('/AI/chat/delete-session', { sessionId })
        if (response.data?.status_code !== 1000) {
          ElMessage.error(response.data?.status_msg || '删除失败')
          return
        }

        delete sessions.value[sessionId]
        if (currentSessionId.value === sessionId) {
          const remaining = Object.keys(sessions.value)
          if (remaining.length > 0) {
            await switchSession(remaining[0])
          } else {
            createNewSession()
          }
        }
        ElMessage.success('对话已删除')
      } catch (error) {
        console.error('Delete session error:', error)
        ElMessage.error('删除失败')
      }
    }

    const shouldUseNewSession = () =>
      tempSession.value || !currentSessionId.value || currentSessionId.value === 'temp'

    const sendMessage = async () => {
      if (!inputMessage.value?.trim()) {
        ElMessage.warning('请输入消息')
        return
      }
      if (loading.value) return
      if (selectedModel.value === '2' && uploadedFiles.value.length === 0) {
        ElMessage.warning('RAG 模式需要知识库文档，请先用 📎 上传 .md / .txt')
        return
      }

      const userMessage = createMessage('user', inputMessage.value)
      const currentInput = inputMessage.value
      inputMessage.value = ''
      if (messageInput.value) messageInput.value.style.height = 'auto'

      appendMessage(userMessage)

      await nextTick()
      scrollToBottom()

      try {
        loading.value = true
        await streamMessage(currentInput)
      } catch (err) {
        console.error('Send message error:', err)
        ElMessage.error(err?.message || '发送失败')
        currentMessages.value.pop()
        if (currentMessages.value.at(-1)?.role === 'assistant' &&
            currentMessages.value.at(-1)?.meta?.status === 'streaming') {
          currentMessages.value.pop()
        }
      } finally {
        loading.value = false
        await nextTick()
        scrollToBottom()
      }
    }

    async function streamMessage(question) {
      const aiMessage = createMessage('assistant', '', { status: 'streaming' })
      const aiMessageIndex = currentMessages.value.length
      appendMessage(aiMessage)

      const useNewSession = shouldUseNewSession()
      const streamPath = useNewSession
        ? '/AI/chat/send-stream-new-session'
        : '/AI/chat/send-stream'

      const body = useNewSession
        ? { question, modelType: selectedModel.value }
        : { question, modelType: selectedModel.value, sessionId: currentSessionId.value }

      try {
        const response = await fetch(getStreamChatUrl(streamPath), {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${localStorage.getItem('token') || ''}`
          },
          body: JSON.stringify(body)
        })

        const contentType = response.headers.get('content-type') || ''
        if (contentType.includes('application/json')) {
          const data = await response.json()
          throw new Error(data.status_msg || '登录已过期，请重新登录')
        }
        if (!response.ok) {
          throw new Error('网络请求失败')
        }

        await consumeChatSSE(response, {
          onChunk: (data) => {
            appendAssistantChunk(aiMessageIndex, data)
            scheduleScroll()
          },
          onSessionId: (newSid) => {
            if (useNewSession) {
              sessions.value[newSid] = {
                id: newSid,
                name: question.slice(0, 20) || '新对话',
                messages: [...currentMessages.value]
              }
              currentSessionId.value = newSid
              tempSession.value = false
            }
          },
          onDone: async () => {
            finalizeAssistantMessage(aiMessageIndex)
            await nextTick()
            scheduleScroll()
          },
          onError: (msg) => {
            loading.value = false
            ElMessage.error(msg)
          }
        })
      } catch (err) {
        currentMessages.value[aiMessageIndex].meta = { status: 'error' }
        throw err
      }
    }

    const scrollToBottom = () => {
      if (messagesRef.value) {
        messagesRef.value.scrollTop = messagesRef.value.scrollHeight
      }
    }

    let scrollScheduled = false
    const scheduleScroll = () => {
      if (scrollScheduled) return
      scrollScheduled = true
      requestAnimationFrame(() => {
        scrollScheduled = false
        scrollToBottom()
      })
    }

    const appendToSession = (msg) => {
      if (tempSession.value || !currentSessionId.value || currentSessionId.value === 'temp') {
        return
      }
      const session = sessions.value[currentSessionId.value]
      if (!session) return
      if (!session.messages) {
        session.messages = []
      }
      session.messages.push(msg)
    }

    const appendMessage = (msg) => {
      currentMessages.value.push(msg)
      appendToSession(msg)
    }

    const triggerFileUpload = () => fileInput.value?.click()

    const handleFileUpload = async (event) => {
      const file = event.target.files?.[0]
      if (!file) return
      const name = file.name.toLowerCase()
      if (!name.endsWith('.md') && !name.endsWith('.txt')) {
        ElMessage.error('仅支持 .md / .txt')
        return
      }
      try {
        uploading.value = true
        const formData = new FormData()
        formData.append('file', file)
        const response = await api.post('/file/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        if (response.data?.status_code === 1000) {
          const item = {
            fileId: response.data.file_id,
            fileName: response.data.file_name || file.name
          }
          if (item.fileId) {
            uploadedFiles.value = [
              item,
              ...uploadedFiles.value.filter(f => f.fileId !== item.fileId)
            ]
          } else {
            await loadUploadedFiles()
          }
          ElMessage.success(
            selectedModel.value === '2'
              ? `「${item.fileName || file.name}」已加入知识库，可直接提问`
              : `「${item.fileName || file.name}」已加入知识库，总结文档请切换到 RAG 模式或点「用 RAG 提问」`
          )
        } else {
          ElMessage.error(response.data?.status_msg || '上传失败')
        }
      } catch {
        ElMessage.error('上传失败')
      } finally {
        uploading.value = false
        if (fileInput.value) fileInput.value.value = ''
      }
    }

    const deleteRagFile = async (file) => {
      if (!file?.fileId) return
      try {
        await ElMessageBox.confirm(
          `确定删除「${file.fileName}」？删除后无法恢复。`,
          '删除文档',
          {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
      } catch {
        return
      }

      try {
        const response = await api.post('/file/delete', { fileId: file.fileId })
        if (response.data?.status_code !== 1000) {
          ElMessage.error(response.data?.status_msg || '删除失败')
          return
        }
        uploadedFiles.value = uploadedFiles.value.filter(f => f.fileId !== file.fileId)
        ElMessage.success('文档已删除')
      } catch (error) {
        console.error('Delete rag file error:', error)
        ElMessage.error('删除失败')
      }
    }

    onMounted(async () => {
      await Promise.all([loadSessions(), loadUploadedFiles()])
      if (!currentSessionId.value) {
        createNewSession()
      }
    })

    return {
      sessions: computed(() => Object.values(sessions.value)),
      currentSessionId,
      tempSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      modeHint,
      composerTip,
      uploading,
      uploadedFiles,
      fileInput,
      renderMarkdown,
      createNewSession,
      switchSession,
      deleteSession,
      sendMessage,
      triggerFileUpload,
      handleFileUpload,
      deleteRagFile,
      autoResize
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

/* ---- 侧边栏 ---- */
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
  padding: 20px 16px 12px;
}

.brand-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--ds-primary);
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-name {
  font-size: 17px;
  font-weight: 600;
  color: var(--ds-text);
}

.btn-new-chat {
  margin: 8px 12px 12px;
  padding: 10px 14px;
  border: 1px solid var(--ds-border);
  border-radius: var(--ds-radius);
  background: var(--ds-bg);
  color: var(--ds-text);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background 0.15s, border-color 0.15s;
}

.btn-new-chat:hover {
  background: #eef0f3;
  border-color: #d0d0d0;
}

.icon-plus {
  font-size: 18px;
  line-height: 1;
  color: var(--ds-primary);
}

.session-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px;
}

.session-empty {
  font-size: 13px;
  color: var(--ds-text-secondary);
  text-align: center;
  padding: 24px 12px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 2px;
  border-radius: 8px;
}

.session-item:hover .session-delete {
  opacity: 1;
}

.session-item.active {
  background: var(--ds-bg);
  box-shadow: var(--ds-shadow);
}

.session-btn {
  flex: 1;
  min-width: 0;
  text-align: left;
  padding: 10px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--ds-text);
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.session-btn:hover {
  background: rgba(0, 0, 0, 0.05);
}

.session-item.active .session-btn {
  font-weight: 500;
}

.session-delete {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  margin-right: 6px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--ds-text-secondary);
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s, background 0.15s, color 0.15s;
}

.session-delete:hover {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
}

.session-item.active .session-delete {
  opacity: 1;
}

.session-title {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.footer-link:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--ds-text);
}

/* ---- 主面板 ---- */
.main-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--ds-bg);
}

.main-header {
  height: 52px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid var(--ds-border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex: 1;
}

.mode-hint {
  font-size: 12px;
  color: var(--ds-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.model-picker {
  padding: 6px 12px;
  border: 1px solid var(--ds-border);
  border-radius: 8px;
  background: var(--ds-bg);
  font-size: 13px;
  color: var(--ds-text);
  cursor: pointer;
  outline: none;
}

.model-picker:focus {
  border-color: var(--ds-primary);
}

.rag-files-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 20px;
  border-bottom: 1px solid var(--ds-border);
  background: var(--ds-bg-muted, #f8fafc);
  flex-shrink: 0;
  min-height: 40px;
  flex-wrap: wrap;
}

.rag-files-label {
  font-size: 12px;
  color: var(--ds-text-secondary);
  white-space: nowrap;
}

.rag-files-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.rag-file-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 260px;
  padding: 4px 6px 4px 10px;
  border-radius: 999px;
  background: var(--ds-bg);
  border: 1px solid var(--ds-border);
  font-size: 12px;
  color: var(--ds-text);
}

.rag-file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.rag-file-delete {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--ds-text-secondary);
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
}

.rag-file-delete:hover {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
}

.rag-switch-btn {
  flex-shrink: 0;
  padding: 4px 12px;
  border: 1px solid var(--ds-primary, #2563eb);
  border-radius: 999px;
  background: var(--ds-bg);
  color: var(--ds-primary, #2563eb);
  font-size: 12px;
  cursor: pointer;
}

.rag-switch-btn:hover {
  background: rgba(37, 99, 235, 0.08);
}

.rag-files-note {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--ds-text-secondary);
}

.icon-btn {
  width: 36px;
  height: 36px;
  border: 1px solid var(--ds-border);
  border-radius: 8px;
  background: var(--ds-bg);
  cursor: pointer;
  font-size: 16px;
}

.icon-btn:hover:not(:disabled) {
  background: var(--ds-bg-muted);
}

.icon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---- 消息区 ---- */
.messages-wrap {
  flex: 1;
  overflow-y: auto;
  padding: 24px 16px 16px;
}

.messages-inner {
  max-width: 768px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.welcome {
  max-width: 520px;
  margin: 80px auto 0;
  text-align: center;
}

.welcome-logo {
  width: 56px;
  height: 56px;
  margin: 0 auto 20px;
  border-radius: 14px;
  background: var(--ds-primary);
  color: #fff;
  font-size: 28px;
  font-weight: 700;
  line-height: 56px;
}

.welcome h2 {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 8px;
}

.welcome p {
  font-size: 14px;
  color: var(--ds-text-secondary);
  margin-bottom: 28px;
}

.welcome-hints {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.welcome-hints button {
  padding: 10px 16px;
  border: 1px solid var(--ds-border);
  border-radius: var(--ds-radius-lg);
  background: var(--ds-bg);
  font-size: 13px;
  color: var(--ds-text);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.welcome-hints button:hover {
  border-color: var(--ds-primary);
  background: #f5f7ff;
}

/* 消息行 */
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
  max-width: calc(100% - 46px);
}

.msg-row-user .msg-body {
  display: flex;
  justify-content: flex-end;
  max-width: 75%;
  margin-left: auto;
}

.user-content {
  display: inline-block;
  max-width: 100%;
  padding: 8px 14px;
  background: var(--ds-user-bubble);
  border-radius: var(--ds-radius-lg);
  font-size: 15px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-content {
  font-size: 15px;
  line-height: 1.75;
  color: var(--ds-text);
  word-break: break-word;
}

.streaming-text {
  display: block;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.75;
}

.ai-rendered {
  line-height: 1.75;
  word-break: break-word;
}

.ai-rendered :deep(p) {
  margin: 0 0 0.75em;
}

.ai-rendered :deep(p:last-child) {
  margin-bottom: 0;
}

.ai-rendered :deep(ul),
.ai-rendered :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.4em;
}

.ai-content :deep(code) {
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.ai-content .typing {
  animation: blink 1s step-end infinite;
  color: var(--ds-primary);
}

@keyframes blink {
  50% { opacity: 0; }
}

/* ---- 输入区 ---- */
.composer-wrap {
  flex-shrink: 0;
  padding: 12px 16px 20px;
  background: linear-gradient(transparent, var(--ds-bg) 20%);
}

.composer {
  max-width: 768px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px 8px 14px;
  min-height: 44px;
  border: 1px solid var(--ds-border);
  border-radius: var(--ds-radius-lg);
  background: var(--ds-bg);
  box-shadow: var(--ds-shadow);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.composer:focus-within {
  border-color: var(--ds-primary);
  box-shadow: 0 0 0 3px rgba(77, 107, 254, 0.12);
}

.composer textarea {
  flex: 1;
  border: none;
  outline: none;
  resize: none;
  font-size: 15px;
  line-height: 1.4;
  min-height: 22px;
  max-height: 100px;
  padding: 2px 0;
  background: transparent;
  font-family: inherit;
  color: var(--ds-text);
}

.composer textarea::placeholder {
  color: #b0b0b0;
}

.send-btn {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border: none;
  border-radius: 10px;
  background: var(--ds-primary);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}

.send-btn:hover:not(:disabled) {
  background: var(--ds-primary-hover);
}

.send-btn:disabled {
  background: #c8c8c8;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.composer-tip {
  max-width: 768px;
  margin: 8px auto 0;
  text-align: center;
  font-size: 12px;
  color: var(--ds-text-secondary);
}

@media (max-width: 768px) {
  .sidebar {
    display: none;
  }
  .welcome {
    margin-top: 40px;
    padding: 0 16px;
  }
}
</style>
