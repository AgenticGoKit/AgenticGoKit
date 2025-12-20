import { useState, type FormEvent } from 'react'
import { useChatStore } from '../store/chat-store'
import { Send, Loader2 } from 'lucide-react'

export function InputArea() {
  const [input, setInput] = useState('')
  const sendMessage = useChatStore((state) => state.sendMessage)
  const workflowState = useChatStore((state) => state.workflowState)
  const wsStatus = useChatStore((state) => state.wsStatus)

  const isDisabled = workflowState.status === 'running' || wsStatus !== 'connected'

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (input.trim() && !isDisabled) {
      sendMessage(input.trim())
      setInput('')
    }
  }

  return (
    <div className="p-3">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={
            isDisabled 
              ? workflowState.status === 'running'
                ? 'Working...'
                : 'Connecting...'
              : 'Describe your story idea...'
          }
          disabled={isDisabled}
          className="flex-1 px-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <button
          type="submit"
          disabled={isDisabled || !input.trim()}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2 disabled:cursor-not-allowed text-sm"
        >
          {workflowState.status === 'running' ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              <span className="hidden sm:inline">Working</span>
            </>
          ) : (
            <>
              <Send className="w-4 h-4" />
              <span className="hidden sm:inline">Send</span>
            </>
          )}
        </button>
      </form>
      
      {workflowState.status === 'error' && workflowState.error && (
        <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-xs text-red-700 dark:text-red-300">
            <strong>Error:</strong> {workflowState.error}
          </p>
        </div>
      )}
    </div>
  )
}
