// Message types for WebSocket communication
export type MessageType =
  | 'user_message'
  | 'workflow_start'
  | 'workflow_info'
  | 'agent_start'
  | 'agent_progress'
  | 'agent_complete'
  | 'workflow_done'
  | 'error'
  | 'chat_history'
  | 'session_created'
  | 'agent_config'

export interface WSMessage {
  type: MessageType
  content?: string
  agent?: string
  step?: string
  progress?: number
  session_id?: string
  timestamp: number
  metadata?: {
    agents?: Agent[]
    workflow_name?: string
    [key: string]: unknown
  }
}

export interface Agent {
  name: string
  displayName: string
  icon: string
  color: string
  description: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'agent' | 'system'
  content: string
  agent?: string
  timestamp: number
  isStreaming?: boolean
  metadata?: {
    [key: string]: unknown
  }
}

export interface WorkflowState {
  status: 'idle' | 'running' | 'completed' | 'error'
  currentAgent?: string
  progress: number
  error?: string
}

export interface AgentStatus {
  name: string
  status: 'idle' | 'working' | 'completed' | 'error'
  progress?: number
}
