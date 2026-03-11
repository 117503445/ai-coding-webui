export type ConnectionState = 'connecting' | 'connected' | 'disconnected'

export interface WSClientOptions {
  url: string
  onMessage: (data: unknown) => void
  onStateChange: (state: ConnectionState) => void
  maxReconnectDelay?: number
  initialReconnectDelay?: number
}

export class WSClient {
  private ws: WebSocket | null = null
  private reconnectDelay: number
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private intentionallyClosed = false
  private opts: Required<WSClientOptions>

  constructor(opts: WSClientOptions) {
    this.opts = {
      maxReconnectDelay: 30000,
      initialReconnectDelay: 1000,
      ...opts,
    }
    this.reconnectDelay = this.opts.initialReconnectDelay
  }

  connect(): void {
    this.intentionallyClosed = false
    this.doConnect()
  }

  private doConnect(): void {
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
    }

    this.opts.onStateChange('connecting')
    const ws = new WebSocket(this.opts.url)

    ws.onopen = () => {
      this.reconnectDelay = this.opts.initialReconnectDelay
      this.opts.onStateChange('connected')
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        this.opts.onMessage(data)
      } catch {
        console.warn('ws: failed to parse message', event.data)
      }
    }

    ws.onclose = () => {
      this.ws = null
      this.opts.onStateChange('disconnected')
      if (!this.intentionallyClosed) {
        this.scheduleReconnect()
      }
    }

    ws.onerror = () => {
      // onclose will fire after onerror
    }

    this.ws = ws
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return

    const delay = this.reconnectDelay
    this.reconnectDelay = Math.min(
      this.reconnectDelay * 2,
      this.opts.maxReconnectDelay,
    )

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.doConnect()
    }, delay)
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  disconnect(): void {
    this.intentionallyClosed = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.opts.onStateChange('disconnected')
  }

  get state(): ConnectionState {
    if (!this.ws) return 'disconnected'
    if (this.ws.readyState === WebSocket.OPEN) return 'connected'
    if (this.ws.readyState === WebSocket.CONNECTING) return 'connecting'
    return 'disconnected'
  }
}
