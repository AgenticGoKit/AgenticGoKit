import { create } from 'zustand'
import type { ChatMessage, WorkflowState, AgentStatus, Agent } from '../types'
import { WebSocketClient, type WebSocketStatus } from '../lib/websocket-client'
import { generateId } from '../lib/utils'

interface ChatStore {
  // WebSocket
  wsClient: WebSocketClient | null
  wsStatus: WebSocketStatus
  
  // Messages
  messages: ChatMessage[]
  streamingMessage: ChatMessage | null
  
  // Workflow
  workflowState: WorkflowState
  agentStatuses: Map<string, AgentStatus>
  workflowName: string
  
  // Agents config
  agents: Agent[]
  
  // Actions
  connect: (url: string) => void
  disconnect: () => void
  sendMessage: (content: string) => void
  addMessage: (message: ChatMessage) => void
  updateStreamingMessage: (content: string, agent?: string) => void
  finalizeStreamingMessage: () => void
  setWorkflowState: (state: Partial<WorkflowState>) => void
  updateAgentStatus: (agent: string, status: Partial<AgentStatus>) => void
  setAgents: (agents: Agent[]) => void
  setWorkflowName: (name: string) => void
  reset: () => void
}

export const useChatStore = create<ChatStore>((set, get) => ({
  wsClient: null,
  wsStatus: 'disconnected',
  messages: [],
  streamingMessage: null,
  workflowState: {
    status: 'idle',
    progress: 0,
  },
  agentStatuses: new Map(),
  agents: [],
  workflowName: 'Workflow',

  connect: (url: string) => {
    const client = new WebSocketClient({
      url,
      onMessage: (message) => {
        const store = get()
        
        switch (message.type) {
          case 'agent_config':
            // Dynamically set agents from server
            if (message.metadata?.agents) {
              set({ agents: message.metadata.agents })
            }
            if (message.metadata?.workflow_name) {
              set({ workflowName: message.metadata.workflow_name as string })
            }
            break

          case 'workflow_start':
            set({
              workflowState: {
                status: 'running',
                progress: 0,
              },
            })
            if (message.content) {
              store.addMessage({
                id: generateId(),
                role: 'system',
                content: message.content,
                timestamp: message.timestamp,
              })
            }
            break

          case 'agent_start':
            if (message.agent) {
              store.updateAgentStatus(message.agent, {
                status: 'working',
                progress: message.progress,
              })
              set({
                workflowState: {
                  ...store.workflowState,
                  currentAgent: message.agent,
                  progress: message.progress || 0,
                },
              })
            }
            break

          case 'agent_progress':
            if (message.content && message.agent) {
              store.updateStreamingMessage(message.content, message.agent)
            }
            break

          case 'agent_complete':
            if (message.agent) {
              store.updateAgentStatus(message.agent, { status: 'completed' })
              store.finalizeStreamingMessage()
            }
            break

          case 'workflow_done':
            set({
              workflowState: {
                status: 'completed',
                progress: 100,
              },
            })
            if (message.content) {
              store.addMessage({
                id: generateId(),
                role: 'system',
                content: 'Workflow completed! ✅',
                timestamp: message.timestamp,
              })
            }
            break

          case 'error':
            set({
              workflowState: {
                status: 'error',
                progress: 0,
                error: message.content,
              },
            })
            if (message.content) {
              store.addMessage({
                id: generateId(),
                role: 'system',
                content: `Error: ${message.content}`,
                timestamp: message.timestamp,
              })
            }
            break
        }
      },
      onStatusChange: (status) => {
        set({ wsStatus: status })
      },
    })

    client.connect()
    set({ wsClient: client })
  },

  disconnect: () => {
    const { wsClient } = get()
    wsClient?.disconnect()
    set({ wsClient: null, wsStatus: 'disconnected' })
  },

  sendMessage: (content: string) => {
    const { wsClient, messages } = get()
    
    console.log('📤 [STORE] Sending message:', content)
    console.log('📤 [STORE] WS Client:', wsClient)
    console.log('📤 [STORE] WS Connected:', wsClient?.isConnected())
    
    // Add user message to UI
    const userMessage: ChatMessage = {
      id: generateId(),
      role: 'user',
      content,
      timestamp: Date.now() / 1000,
    }
    
    set({ messages: [...messages, userMessage] })

    // Send to server
    if (wsClient?.isConnected()) {
      const message = {
        type: 'user_message',
        content,
        timestamp: Date.now() / 1000,
      }
      console.log('📤 [STORE] Sending to WebSocket:', message)
      wsClient.send(message)
    } else {
      console.error('❌ [STORE] WebSocket not connected!')
    }
  },

  addMessage: (message: ChatMessage) => {
    set((state) => ({
      messages: [...state.messages, message],
    }))
  },

  updateStreamingMessage: (content: string, agent?: string) => {
    set((state) => {
      if (state.streamingMessage) {
        return {
          streamingMessage: {
            ...state.streamingMessage,
            content: state.streamingMessage.content + content,
          },
        }
      } else {
        return {
          streamingMessage: {
            id: generateId(),
            role: 'agent',
            content,
            agent,
            timestamp: Date.now() / 1000,
            isStreaming: true,
          },
        }
      }
    })
  },

  finalizeStreamingMessage: () => {
    set((state) => {
      if (state.streamingMessage) {
        return {
          messages: [
            ...state.messages,
            { ...state.streamingMessage, isStreaming: false },
          ],
          streamingMessage: null,
        }
      }
      return state
    })
  },

  setWorkflowState: (newState: Partial<WorkflowState>) => {
    set((state) => ({
      workflowState: { ...state.workflowState, ...newState },
    }))
  },

  updateAgentStatus: (agent: string, status: Partial<AgentStatus>) => {
    set((state) => {
      const newStatuses = new Map(state.agentStatuses)
      const current = newStatuses.get(agent) || { name: agent, status: 'idle' }
      newStatuses.set(agent, { ...current, ...status })
      return { agentStatuses: newStatuses }
    })
  },

  setAgents: (agents: Agent[]) => {
    set({ agents })
  },

  setWorkflowName: (name: string) => {
    set({ workflowName: name })
  },

  reset: () => {
    set({
      messages: [],
      streamingMessage: null,
      workflowState: {
        status: 'idle',
        progress: 0,
      },
      agentStatuses: new Map(),
    })
  },
}))
