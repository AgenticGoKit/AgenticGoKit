import type { Agent } from '../types'
import { cn } from '../lib/utils'
import { CheckCircle2, Loader2, Circle } from 'lucide-react'

interface AgentCardProps {
  agent: Agent
  status: 'idle' | 'working' | 'completed' | 'error'
  progress?: number
  isActive: boolean
}

export function AgentCard({ agent, status, progress, isActive }: AgentCardProps) {
  return (
    <div className={cn(
      'relative p-3 rounded-lg border-2 transition-all duration-300',
      isActive 
        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 shadow-lg'
        : status === 'completed'
        ? 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/10'
        : status === 'error'
        ? 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10'
        : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800'
    )}>
      {/* Status indicator */}
      <div className="absolute top-2 right-2">
        {status === 'working' ? (
          <Loader2 className="w-4 h-4 text-blue-600 dark:text-blue-400 animate-spin" />
        ) : status === 'completed' ? (
          <CheckCircle2 className="w-4 h-4 text-green-600 dark:text-green-400" />
        ) : (
          <Circle className="w-4 h-4 text-slate-300 dark:text-slate-600" />
        )}
      </div>

      {/* Agent Info */}
      <div className="flex items-start gap-2">
        <div className={cn(
          'flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center text-xl shadow-sm',
          agent.color === 'blue' && 'bg-blue-100 dark:bg-blue-900',
          agent.color === 'green' && 'bg-green-100 dark:bg-green-900',
          agent.color === 'purple' && 'bg-purple-100 dark:bg-purple-900'
        )}>
          {agent.icon}
        </div>

        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
            {agent.displayName}
          </h3>
          <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5 line-clamp-1">
            {agent.description}
          </p>

          {/* Status text */}
          <div className="mt-1.5">
            <span className={cn(
              'inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full',
              status === 'idle' && 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300',
              status === 'working' && 'bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-300',
              status === 'completed' && 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-300',
              status === 'error' && 'bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300'
            )}>
              {status === 'idle' && 'Ready'}
              {status === 'working' && 'Working...'}
              {status === 'completed' && 'Done'}
              {status === 'error' && 'Error'}
            </span>
          </div>

          {/* Progress bar for working status */}
          {status === 'working' && progress !== undefined && (
            <div className="mt-2">
              <div className="flex items-center justify-between mb-0.5">
                <span className="text-xs text-slate-600 dark:text-slate-400">
                  Progress
                </span>
                <span className="text-xs font-medium text-slate-700 dark:text-slate-300">
                  {progress}%
                </span>
              </div>
              <div className="h-1 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                <div 
                  className="h-full bg-blue-600 dark:bg-blue-500 transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
