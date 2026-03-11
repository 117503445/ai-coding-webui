import { Loader2, WifiOff } from 'lucide-react'
import type { ConnectionState } from '@/lib/ws'

interface ConnectionOverlayProps {
  state: ConnectionState
}

export function ConnectionOverlay({ state }: ConnectionOverlayProps) {
  if (state === 'connected') return null

  return (
    <div className="absolute inset-0 bg-background/60 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-card border border-border rounded-xl p-6 shadow-lg text-center max-w-xs mx-4">
        {state === 'connecting' ? (
          <>
            <Loader2 className="w-10 h-10 text-primary animate-spin mx-auto mb-3" />
            <h3 className="font-semibold mb-1">正在连接...</h3>
            <p className="text-sm text-muted-foreground">
              正在尝试连接到服务器，请稍候。
            </p>
          </>
        ) : (
          <>
            <WifiOff className="w-10 h-10 text-destructive mx-auto mb-3" />
            <h3 className="font-semibold mb-1">连接已断开</h3>
            <p className="text-sm text-muted-foreground">
              正在自动重连，请检查网络连接。
            </p>
          </>
        )}
      </div>
    </div>
  )
}
