import { useChatStore } from '../store/chat-store'
import { MessageList } from './MessageList'
import { InputArea } from './InputArea'

export function ChatInterface() {
  const workflowState = useChatStore((state) => state.workflowState)
  const messages = useChatStore((state) => state.messages)
  const streamingMessage = useChatStore((state) => state.streamingMessage)

  return (
    <div className="h-full flex flex-col bg-white dark:bg-slate-900 rounded-lg shadow-lg border border-slate-200 dark:border-slate-700">
      {/* Header */}
      <div className="flex-shrink-0 px-4 py-2 border-b border-slate-200 dark:border-slate-700">
        <h2 className="text-sm sm:text-base font-semibold text-slate-900 dark:text-white">
          Conversation
        </h2>
        <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5 line-clamp-1">
          {workflowState.status === 'running' 
            ? `In progress... (${workflowState.progress}%)`
            : workflowState.status === 'completed'
            ? 'Completed'
            : 'Start by describing a story idea'}
        </p>
      </div>

      {/* Messages - Scrollable */}
      <div className="flex-1 min-h-0 relative overflow-hidden">
        <MessageList 
          messages={messages}
          streamingMessage={streamingMessage}
        />
      </div>

      {/* Input - Fixed at bottom */}
      <div className="flex-shrink-0 border-t border-slate-200 dark:border-slate-700">
        <InputArea />
      </div>
    </div>
  )
}
