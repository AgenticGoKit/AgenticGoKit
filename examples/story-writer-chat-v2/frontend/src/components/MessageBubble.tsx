import { useMemo } from 'react'
import type { ChatMessage } from '../types'
import { useChatStore } from '../store/chat-store'
import { formatTimestamp, cn } from '../lib/utils'

interface MessageBubbleProps {
  message: ChatMessage
  isStreaming?: boolean
}

export function MessageBubble({ message, isStreaming = false }: MessageBubbleProps) {
  const agents = useChatStore((state) => state.agents)
  
  const agent = useMemo(() => {
    if (message.agent) {
      return agents.find(a => a.name === message.agent)
    }
    return null
  }, [message.agent, agents])

  const isUser = message.role === 'user'
  const isSystem = message.role === 'system'

  if (isSystem) {
    return (
      <div className="flex justify-center">
        <div className="bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 px-4 py-2 rounded-full text-sm">
          {message.content}
        </div>
      </div>
    )
  }

  return (
    <div className={cn(
      'flex gap-3 animate-fade-in',
      isUser ? 'flex-row-reverse' : 'flex-row'
    )}>
      {/* Avatar */}
      {!isUser && (
        <div className={cn(
          'flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center text-xl shadow-sm',
          agent?.color === 'blue' && 'bg-blue-100 dark:bg-blue-900',
          agent?.color === 'green' && 'bg-green-100 dark:bg-green-900',
          agent?.color === 'purple' && 'bg-purple-100 dark:bg-purple-900',
          !agent && 'bg-slate-100 dark:bg-slate-800'
        )}>
          {agent?.icon || '🤖'}
        </div>
      )}

      {/* Message Content */}
      <div className={cn(
        'flex flex-col gap-1 max-w-[80%]',
        isUser ? 'items-end' : 'items-start'
      )}>
        {/* Agent Name & Timestamp */}
        {!isUser && agent && (
          <div className="flex items-center gap-2 px-2">
            <span className="text-sm font-semibold text-slate-900 dark:text-white">
              {agent.displayName}
            </span>
            <span className="text-xs text-slate-500 dark:text-slate-400">
              {formatTimestamp(message.timestamp)}
            </span>
          </div>
        )}

        {/* Message Bubble */}
        <div className={cn(
          'px-4 py-3 rounded-2xl shadow-sm',
          isUser 
            ? 'bg-blue-600 text-white rounded-tr-sm'
            : 'bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white rounded-tl-sm'
        )}>
          <div className="whitespace-pre-wrap break-words">
            {message.content}
          </div>
          
          {isStreaming && (
            <span className="inline-block w-2 h-4 bg-current animate-pulse ml-1" />
          )}
        </div>

        {/* User timestamp */}
        {isUser && (
          <span className="text-xs text-slate-500 dark:text-slate-400 px-2">
            {formatTimestamp(message.timestamp)}
          </span>
        )}
      </div>

      {/* User Avatar */}
      {isUser && (
        <div className="flex-shrink-0 w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center text-white text-xl shadow-sm">
          👤
        </div>
      )}
    </div>
  )
}
