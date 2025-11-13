import type { WSMessage } from '../types'

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface WebSocketClientOptions {
  url: string
  onMessage: (message: WSMessage) => void
  onStatusChange: (status: WebSocketStatus) => void
  reconnectInterval?: number
  maxReconnectAttempts?: number
}

export class WebSocketClient {
  private ws: WebSocket | null = null
  private options: WebSocketClientOptions
  private reconnectAttempts = 0
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null

  constructor(options: WebSocketClientOptions) {
    this.options = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      ...options,
    }
  }

  connect(): void {
    try {
      console.log('🔌 [WS] Connecting to:', this.options.url)
      this.options.onStatusChange('connecting')
      this.ws = new WebSocket(this.options.url)

      this.ws.onopen = () => {
        console.log('✅ [WS] WebSocket connected')
        this.options.onStatusChange('connected')
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event) => {
        try {
          console.log('📨 [WS] Received message:', event.data)
          const message: WSMessage = JSON.parse(event.data)
          console.log('📨 [WS] Parsed message:', message)
          this.options.onMessage(message)
        } catch (error) {
          console.error('❌ [WS] Failed to parse message:', error)
        }
      }

      this.ws.onerror = (error) => {
        console.error('❌ [WS] WebSocket error:', error)
        this.options.onStatusChange('error')
      }

      this.ws.onclose = (event) => {
        console.log('🔌 [WS] WebSocket closed:', event.code, event.reason)
        console.log('🔌 [WS] Was clean:', event.wasClean)
        this.options.onStatusChange('disconnected')

        if (
          this.reconnectAttempts < (this.options.maxReconnectAttempts || 5) &&
          !event.wasClean
        ) {
          this.scheduleReconnect()
        }
      }
    } catch (error) {
      console.error('❌ [WS] Connection error:', error)
      this.options.onStatusChange('error')
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= (this.options.maxReconnectAttempts || 5)) {
      console.error('❌ [WS] Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    console.log(`⏳ [WS] Reconnecting... (attempt ${this.reconnectAttempts})`)

    this.reconnectTimeout = setTimeout(() => {
      this.connect()
    }, this.options.reconnectInterval)
  }

  send(message: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  disconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}
