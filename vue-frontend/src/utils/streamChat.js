/**
 * 读取聊天 SSE 流，解析 data: 行并回调。
 */
export function getStreamChatUrl(path) {
  const base = (process.env.VUE_APP_STREAM_BASE || '/api').replace(/\/$/, '')
  return `${base}${path}`
}

export async function consumeChatSSE(response, handlers) {
  const { onChunk, onSessionId, onDone, onError } = handlers
  if (!response.ok) {
    throw new Error('Network response was not ok')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let pendingDataLines = []
  let sawDone = false

  const finish = () => {
    if (sawDone) return
    sawDone = true
    onDone?.()
  }

  const dispatchEvent = (dataLines) => {
    if (!dataLines.length) return
    const data = dataLines.join('\n')
    if (data === '[DONE]') {
      finish()
      return
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

  const processLine = (line) => {
    if (line === '') {
      dispatchEvent(pendingDataLines)
      pendingDataLines = []
      return
    }
    if (line.startsWith(':')) return
    if (line.startsWith('data:')) {
      pendingDataLines.push(line.slice(5).replace(/^\s/, ''))
    }
  }

  const drainBufferLines = () => {
    let idx = buffer.indexOf('\n')
    while (idx !== -1) {
      let line = buffer.slice(0, idx)
      buffer = buffer.slice(idx + 1)
      if (line.endsWith('\r')) line = line.slice(0, -1)
      processLine(line)
      idx = buffer.indexOf('\n')
    }
  }

  let readResult = await reader.read()
  while (!readResult.done) {
    buffer += decoder.decode(readResult.value, { stream: true })
    drainBufferLines()
    readResult = await reader.read()
  }

  buffer += decoder.decode()
  if (buffer.length > 0) {
    let line = buffer
    if (line.endsWith('\r')) line = line.slice(0, -1)
    processLine(line)
  }
  dispatchEvent(pendingDataLines)
  pendingDataLines = []
  finish()
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
  const withCode = safe.replace(/`([^`]+)`/g, '<code>$1</code>')
  const blocks = withCode.split(/\n{2,}/)
  return blocks
    .map(block => {
      const inline = block
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/\n/g, '<br>')
      return `<p>${inline}</p>`
    })
    .join('')
}
