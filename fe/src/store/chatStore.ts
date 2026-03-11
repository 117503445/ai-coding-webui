import { create } from 'zustand'
import type { ConnectionState } from '@/lib/ws'
import { WSClient } from '@/lib/ws'

export interface ThinkingBlock {
  type: 'thinking'
  content: string
}

export interface TextBlock {
  type: 'text'
  content: string
}

export interface ToolUseBlock {
  type: 'tool_use'
  id: string
  name: string
  input: string
}

export interface ToolResultBlock {
  type: 'tool_result'
  toolUseId: string
  content: string
}

export type ContentBlock = ThinkingBlock | TextBlock | ToolUseBlock | ToolResultBlock

export interface ChatMessage {
  role: 'user' | 'assistant'
  blocks: ContentBlock[]
  timestamp: number
}

interface StreamState {
  currentBlocks: ContentBlock[]
  blockIndex: number
}

export type WorkStatus = 'idle' | 'working'

interface ChatStore {
  connectionState: ConnectionState
  workStatus: WorkStatus
  workDetail: string
  sessionId: string
  messages: ChatMessage[]
  stream: StreamState
  wsClient: WSClient | null

  init: () => void
  cleanup: () => void
  sendMessage: (content: string) => void
  sendCommand: (command: string) => void
  abort: () => void
  newChat: () => void
}

const STORAGE_KEY = 'claude-webui-session'

function loadSession(): { sessionId: string; messages: ChatMessage[] } {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const data = JSON.parse(raw)
      return { sessionId: data.sessionId || '', messages: data.messages || [] }
    }
  } catch { /* ignore */ }
  return { sessionId: '', messages: [] }
}

function saveSession(sessionId: string, messages: ChatMessage[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ sessionId, messages }))
  } catch { /* ignore */ }
}

function getWsUrl(): string {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws`
}

export const useChatStore = create<ChatStore>((set, get) => ({
  connectionState: 'disconnected' as ConnectionState,
  workStatus: 'idle' as WorkStatus,
  workDetail: '',
  sessionId: '',
  messages: [],
  stream: { currentBlocks: [], blockIndex: -1 },
  wsClient: null,

  init: () => {
    const saved = loadSession()
    set({ sessionId: saved.sessionId, messages: saved.messages })

    const client = new WSClient({
      url: getWsUrl(),
      onMessage: (data) => handleWSMessage(data, set, get),
      onStateChange: (state) => set({ connectionState: state }),
    })
    client.connect()
    set({ wsClient: client })
  },

  cleanup: () => {
    get().wsClient?.disconnect()
    set({ wsClient: null })
  },

  sendMessage: (content: string) => {
    const { wsClient, sessionId, messages } = get()
    if (!wsClient) return

    const userMsg: ChatMessage = {
      role: 'user',
      blocks: [{ type: 'text', content }],
      timestamp: Date.now(),
    }

    const newMessages = [...messages, userMsg]
    set({ messages: newMessages, stream: { currentBlocks: [], blockIndex: -1 } })
    saveSession(sessionId, newMessages)

    wsClient.send({ type: 'chat', content, session_id: sessionId })
  },

  sendCommand: (command: string) => {
    get().wsClient?.send({ type: 'command', command })
  },

  abort: () => {
    get().wsClient?.send({ type: 'abort' })
  },

  newChat: () => {
    const { wsClient } = get()
    wsClient?.send({ type: 'command', command: '/new' })
    set({
      sessionId: '',
      messages: [],
      stream: { currentBlocks: [], blockIndex: -1 },
      workStatus: 'idle',
      workDetail: '',
    })
    saveSession('', [])
  },
}))

function handleWSMessage(
  data: unknown,
  set: (partial: Partial<ChatStore>) => void,
  get: () => ChatStore,
) {
  const msg = data as Record<string, unknown>
  switch (msg.type) {
    case 'status':
      set({
        workStatus: (msg.status as WorkStatus) || 'idle',
        workDetail: (msg.detail as string) || '',
      })
      if (msg.session_id) {
        set({ sessionId: msg.session_id as string })
      }
      break

    case 'stream':
      handleStreamEvent(msg.event, set, get)
      break

    case 'complete': {
      const { messages, stream } = get()
      const sid = (msg.session_id as string) || get().sessionId

      if (stream.currentBlocks.length > 0) {
        const assistantMsg: ChatMessage = {
          role: 'assistant',
          blocks: [...stream.currentBlocks],
          timestamp: Date.now(),
        }
        const newMessages = [...messages, assistantMsg]
        set({
          messages: newMessages,
          stream: { currentBlocks: [], blockIndex: -1 },
          sessionId: sid,
        })
        saveSession(sid, newMessages)
      } else {
        set({ sessionId: sid })
        saveSession(sid, messages)
      }
      break
    }

    case 'error':
      console.error('ws error:', msg.message)
      break

    case 'command_result':
      break
  }
}

function parseContentBlocks(content: unknown[]): ContentBlock[] {
  const blocks: ContentBlock[] = []
  for (const item of content) {
    const c = item as Record<string, unknown>
    const blockType = c.type as string
    if (blockType === 'thinking') {
      blocks.push({ type: 'thinking', content: (c.thinking as string) || '' })
    } else if (blockType === 'text') {
      blocks.push({ type: 'text', content: (c.text as string) || '' })
    } else if (blockType === 'tool_use') {
      blocks.push({
        type: 'tool_use',
        id: (c.id as string) || '',
        name: (c.name as string) || '',
        input: typeof c.input === 'string' ? c.input : JSON.stringify(c.input || ''),
      })
    } else if (blockType === 'tool_result') {
      blocks.push({
        type: 'tool_result',
        toolUseId: (c.tool_use_id as string) || '',
        content: typeof c.content === 'string' ? c.content : JSON.stringify(c.content || ''),
      })
    }
  }
  return blocks
}

function handleStreamEvent(
  event: unknown,
  set: (partial: Partial<ChatStore>) => void,
  get: () => ChatStore,
) {
  if (!event) return
  const raw = event as Record<string, unknown>
  const eventType = raw.type as string

  // claude CLI stream-json 格式：assistant 事件包含 message.content 数组
  if (eventType === 'assistant') {
    const message = raw.message as Record<string, unknown> | undefined
    if (message) {
      const content = message.content as unknown[] | undefined
      if (content && Array.isArray(content) && content.length > 0) {
        const newBlocks = parseContentBlocks(content)
        if (newBlocks.length > 0) {
          const { stream } = get()
          const merged = [...stream.currentBlocks, ...newBlocks]
          set({ stream: { currentBlocks: merged, blockIndex: merged.length - 1 } })
        }
      }
    }
    return
  }

  // system/init 事件不需要处理
  if (eventType === 'system') return

  // Anthropic API 原生流式格式（content_block_start/delta/stop）
  if (eventType === 'content_block_start' ||
      eventType === 'content_block_delta' ||
      eventType === 'content_block_stop' ||
      eventType === 'result') {
    handleAnthropicStreamEvent(raw, eventType, set, get)
    return
  }

  // stream_event 包装器（某些 claude CLI 版本使用此格式）
  if (eventType === 'stream_event') {
    handleStreamEvent(raw.event, set, get)
    return
  }
}

function handleAnthropicStreamEvent(
  raw: Record<string, unknown>,
  eventType: string,
  set: (partial: Partial<ChatStore>) => void,
  get: () => ChatStore,
) {
  const { stream } = get()
  const blocks = [...stream.currentBlocks]

  if (eventType === 'content_block_start') {
    const cb = raw.content_block as Record<string, unknown>
    const blockType = cb?.type as string
    const idx = (raw.index as number) ?? blocks.length

    if (blockType === 'thinking') {
      blocks[idx] = { type: 'thinking', content: '' }
    } else if (blockType === 'text') {
      blocks[idx] = { type: 'text', content: '' }
    } else if (blockType === 'tool_use') {
      blocks[idx] = {
        type: 'tool_use',
        id: (cb.id as string) || '',
        name: (cb.name as string) || '',
        input: '',
      }
    }
    set({ stream: { currentBlocks: blocks, blockIndex: idx } })
    return
  }

  if (eventType === 'content_block_delta') {
    const idx = (raw.index as number) ?? stream.blockIndex
    const delta = raw.delta as Record<string, unknown>
    if (!delta || idx < 0 || idx >= blocks.length) return

    const block = blocks[idx]
    const deltaType = delta.type as string

    if (deltaType === 'thinking_delta' && block.type === 'thinking') {
      blocks[idx] = { ...block, content: block.content + (delta.thinking as string || '') }
    } else if (deltaType === 'text_delta' && block.type === 'text') {
      blocks[idx] = { ...block, content: block.content + (delta.text as string || '') }
    } else if (deltaType === 'input_json_delta' && block.type === 'tool_use') {
      blocks[idx] = { ...block, input: block.input + (delta.partial_json as string || '') }
    }
    set({ stream: { currentBlocks: blocks, blockIndex: idx } })
    return
  }

  if (eventType === 'result') {
    const subtype = raw.subtype as string
    if (subtype === 'tool_result') {
      const toolUseId = raw.tool_use_id as string || ''
      const content = typeof raw.content === 'string' ? raw.content : JSON.stringify(raw.content)
      blocks.push({ type: 'tool_result', toolUseId, content })
      set({ stream: { currentBlocks: blocks, blockIndex: blocks.length - 1 } })
    }
  }
}
