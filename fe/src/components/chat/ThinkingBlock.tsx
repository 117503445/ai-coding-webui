import { useState } from 'react'
import { ChevronDown, ChevronRight, Brain } from 'lucide-react'

interface ThinkingBlockProps {
  content: string
  isStreaming?: boolean
}

export function ThinkingBlock({ content, isStreaming }: ThinkingBlockProps) {
  const [expanded, setExpanded] = useState(false)

  if (!content && !isStreaming) return null

  const lines = content.split('\n').length
  const preview = content.slice(0, 120).replace(/\n/g, ' ')

  return (
    <div className="my-2 rounded-lg border border-border/50 bg-muted/30 overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground hover:bg-muted/50 transition-colors"
      >
        <Brain className={`w-3.5 h-3.5 flex-shrink-0 ${isStreaming ? 'animate-pulse text-primary' : ''}`} />
        <span className="font-medium">
          {isStreaming ? '思考中...' : `思考过程 (${lines} 行)`}
        </span>
        {expanded ? (
          <ChevronDown className="w-3.5 h-3.5 ml-auto" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5 ml-auto" />
        )}
      </button>
      {!expanded && content && (
        <div className="px-3 pb-2 text-xs text-muted-foreground/70 italic truncate">
          {preview}...
        </div>
      )}
      {expanded && (
        <div className="px-3 pb-3 text-xs text-muted-foreground/80 italic whitespace-pre-wrap font-mono leading-relaxed max-h-96 overflow-y-auto">
          {content}
        </div>
      )}
    </div>
  )
}
