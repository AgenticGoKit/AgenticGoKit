import { useChatStore } from '../store/chat-store'
import { AgentCard } from './AgentCard'

export function WorkflowVisualizer() {
  const agents = useChatStore((state) => state.agents)
  const agentStatuses = useChatStore((state) => state.agentStatuses)
  const workflowState = useChatStore((state) => state.workflowState)

  return (
    <div className="h-full flex flex-col bg-white dark:bg-slate-900 rounded-lg shadow-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-2 border-b border-slate-200 dark:border-slate-700">
        <h2 className="text-sm font-semibold text-slate-900 dark:text-white">
          Workflow Progress
        </h2>
        <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5">
          Multi-agent collaboration
        </p>
      </div>

      {/* Progress Bar */}
      <div className="px-4 py-2 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-medium text-slate-700 dark:text-slate-300">
            Overall Progress
          </span>
          <span className="text-xs font-semibold text-blue-600 dark:text-blue-400">
            {workflowState.progress}%
          </span>
        </div>
        <div className="h-1.5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
          <div 
            className="h-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-500 ease-out"
            style={{ width: `${workflowState.progress}%` }}
          />
        </div>
        
        <div className="mt-1 text-xs text-slate-600 dark:text-slate-400 truncate">
          {workflowState.status === 'idle' && 'Ready to start'}
          {workflowState.status === 'running' && `Working: ${workflowState.currentAgent || '...'}`}
          {workflowState.status === 'completed' && 'Completed! ✅'}
          {workflowState.status === 'error' && 'Error ❌'}
        </div>
      </div>

      {/* Agent Cards */}
      <div className="flex-1 min-h-0 p-3 space-y-2 overflow-y-auto">
        {agents.map((agent) => {
          const status = agentStatuses.get(agent.name)
          return (
            <AgentCard 
              key={agent.name}
              agent={agent}
              status={status?.status || 'idle'}
              progress={status?.progress}
              isActive={workflowState.currentAgent === agent.name}
            />
          )
        })}
      </div>

      {/* Legend */}
      <div className="px-4 py-2 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800">
        <div className="flex items-center justify-around text-xs">
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-slate-300 dark:bg-slate-600" />
            <span className="text-slate-600 dark:text-slate-400">Idle</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
            <span className="text-slate-600 dark:text-slate-400">Working</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-green-500" />
            <span className="text-slate-600 dark:text-slate-400">Done</span>
          </div>
        </div>
      </div>
    </div>
  )
}
