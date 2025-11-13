import { useEffect, useRef } from 'react'
import type { ChatMessage } from '../types'
import { MessageBubble } from './MessageBubble'

interface MessageListProps {
  messages: ChatMessage[]
  streamingMessage: ChatMessage | null
}

export function MessageList({ messages, streamingMessage }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingMessage])

  if (messages.length === 0 && !streamingMessage) {
    return (
      <div className="h-full flex items-center justify-center p-6">
        <div className="text-center max-w-md">
          <div className="text-6xl mb-4">📝</div>
          <h3 className="text-xl font-semibold text-slate-900 dark:text-white mb-2">
            Start Your Story
          </h3>
          <p className="text-slate-600 dark:text-slate-400">
            Describe a story idea and watch as Writer, Editor, and Publisher collaborate to create an amazing tale!
          </p>
        </div>
      </div>
    )
  }

  return (
    <div 
      className="absolute inset-0 overflow-y-scroll p-3 space-y-3 scrollbar scrollbar-thumb-slate-400 dark:scrollbar-thumb-slate-600 scrollbar-track-slate-100 dark:scrollbar-track-slate-800 scrollbar-thumb-rounded"
    >
      {messages.map((message) => (
        <MessageBubble key={message.id} message={message} />
      ))}
      
      {streamingMessage && (
        <MessageBubble message={streamingMessage} isStreaming />
      )}
      
      <div ref={messagesEndRef} />
    </div>
  )
}
