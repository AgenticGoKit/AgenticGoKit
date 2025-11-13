import { useEffect } from 'react'
import { useChatStore } from './store/chat-store'
import { ChatInterface } from './components/ChatInterface'
import { WorkflowVisualizer } from './components/WorkflowVisualizer'

function App() {
  const connect = useChatStore((state) => state.connect)
  const disconnect = useChatStore((state) => state.disconnect)
  const wsStatus = useChatStore((state) => state.wsStatus)
  const workflowName = useChatStore((state) => state.workflowName)

  useEffect(() => {
    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = import.meta.env.DEV 
      ? 'localhost:8080' 
      : window.location.host
    const wsUrl = `${protocol}//${host}/ws`
    
    connect(wsUrl)

    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return (
    <div className="h-screen w-screen flex flex-col bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 overflow-hidden">
      {/* Header */}
      <header className="flex-shrink-0 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 px-4 py-2 sm:py-3 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <div className="text-2xl flex-shrink-0">📝</div>
            <div className="min-w-0">
              <h1 className="text-base sm:text-lg font-bold text-slate-900 dark:text-white truncate">
                {workflowName}
              </h1>
              <p className="hidden md:block text-xs text-slate-600 dark:text-slate-400 truncate">
                AI-powered collaborative workflow
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-2 flex-shrink-0">
            <div className={`flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${
              wsStatus === 'connected' 
                ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                : wsStatus === 'connecting'
                ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
                : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
            }`}>
              <div className={`w-1.5 h-1.5 rounded-full ${
                wsStatus === 'connected' ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
              }`} />
              <span className="hidden sm:inline">
                {wsStatus === 'connected' ? 'Connected' : wsStatus === 'connecting' ? 'Connecting...' : 'Disconnected'}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 min-h-0 overflow-hidden">
        <div className="h-full p-3 sm:p-4">
          <div className="h-full grid grid-cols-1 lg:grid-cols-3 gap-3 sm:gap-4">
            {/* Chat Interface - Full width on mobile, 2 columns on large screens */}
            <div className="lg:col-span-2 h-full overflow-hidden">
              <ChatInterface />
            </div>

            {/* Workflow Visualizer - Hidden on small screens, shown on large screens */}
            <div className="hidden lg:block lg:col-span-1 h-full overflow-hidden">
              <WorkflowVisualizer />
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="flex-shrink-0 bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-700 px-4 py-2">
        <div className="text-center text-xs text-slate-600 dark:text-slate-400">
          Powered by <span className="font-semibold text-slate-900 dark:text-white">AgenticGoKit</span> v.next
        </div>
      </footer>
    </div>
  )
}

export default App
