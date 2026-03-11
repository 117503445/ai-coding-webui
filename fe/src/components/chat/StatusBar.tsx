import { Loader2, Wifi, WifiOff, Zap, Plus } from 'lucide-react'
import type { ConnectionState } from '@/lib/ws'
import type { WorkStatus } from '@/store/chatStore'

interface StatusBarProps {
  connectionState: ConnectionState
  workStatus: WorkStatus
  workDetail: string
  onNewChat: () => void
}

export function StatusBar({ connectionState, workStatus, workDetail, onNewChat }: StatusBarProps) {
  return (
    <header className="h-12 border-b border-border bg-background/80 backdrop-blur-sm flex items-center justify-between px-4 flex-shrink-0">
      <div className="flex items-center gap-2">
        <div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center">
          <Zap className="w-4 h-4 text-primary-foreground" />
        </div>
        <span className="font-semibold text-sm hidden sm:inline">Claude Code</span>
      </div>

      <div className="flex items-center gap-3">
        {workStatus === 'working' && (
          <div className="flex items-center gap-1.5 text-xs text-primary">
            <div className="w-1.5 h-1.5 bg-primary rounded-full animate-pulse" />
            <span className="hidden sm:inline max-w-[180px] truncate">{workDetail || '工作中...'}</span>
          </div>
        )}

        <ConnectionBadge state={connectionState} />

        <button
          onClick={onNewChat}
          className="p-1.5 rounded-md hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
          title="新对话"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>
    </header>
  )
}

function ConnectionBadge({ state }: { state: ConnectionState }) {
  if (state === 'connected') {
    return (
      <div className="flex items-center gap-1 text-xs text-green-600">
        <Wifi className="w-3.5 h-3.5" />
        <span className="hidden sm:inline">已连接</span>
      </div>
    )
  }

  if (state === 'connecting') {
    return (
      <div className="flex items-center gap-1 text-xs text-amber-500">
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
        <span className="hidden sm:inline">连接中...</span>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-1 text-xs text-destructive">
      <WifiOff className="w-3.5 h-3.5" />
      <span className="hidden sm:inline">已断开</span>
    </div>
  )
}
