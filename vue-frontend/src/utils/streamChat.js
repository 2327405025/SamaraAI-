/**
 * 读取聊天 SSE 流，解析 data: 行并回调。
 */
export async function consumeChatSSE(response, handlers) {
  const { onChunk, onSessionId, onDone, onError } = handlers
  if (!response.ok) {
    throw new Error('Network response was not ok')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data:')) continue

      const data = trimmed.slice(5).trim()
      if (data === '[DONE]') {
        onDone?.()
        continue
      }
      if (data.startsWith('{')) {
        try {
          const parsed = JSON.parse(data)
          if (parsed.sessionId) {
            onSessionId?.(String(parsed.sessionId))
          } else if (parsed.message) {
            onError?.(parsed.message)
          }
        } catch {
          onChunk?.(data)
        }
      } else {
        onChunk?.(data)
      }
    }
  }
  onDone?.()
}

export function escapeHtml(text) {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

export function renderSimpleMarkdown(text) {
  if (!text && text !== '') return ''
  const safe = escapeHtml(text)
  return safe
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}
