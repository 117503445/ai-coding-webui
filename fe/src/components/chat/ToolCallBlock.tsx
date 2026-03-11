import { useState } from 'react'
import { ChevronDown, ChevronRight, Wrench, CheckCircle2 } from 'lucide-react'

interface ToolCallBlockProps {
  name: string
  input: string
  result?: string
  isStreaming?: boolean
}

export function ToolCallBlock({ name, input, result, isStreaming }: ToolCallBlockProps) {
  const [expanded, setExpanded] = useState(false)

  let parsedInput = input
  try {
    parsedInput = JSON.stringify(JSON.parse(input), null, 2)
  } catch { /* raw string */ }

  return (
    <div className="my-2 rounded-lg border border-border/50 bg-card overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-3 py-2 text-xs hover:bg-muted/30 transition-colors"
      >
        {isStreaming ? (
          <Wrench className="w-3.5 h-3.5 flex-shrink-0 text-amber-500 animate-spin" />
        ) : (
          <CheckCircle2 className="w-3.5 h-3.5 flex-shrink-0 text-green-500" />
        )}
        <span className="font-medium font-mono text-foreground">{name}</span>
        {isStreaming && (
          <span className="text-muted-foreground">运行中...</span>
        )}
        {expanded ? (
          <ChevronDown className="w-3.5 h-3.5 ml-auto text-muted-foreground" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5 ml-auto text-muted-foreground" />
        )}
      </button>

      {expanded && (
        <div className="border-t border-border/30">
          {parsedInput && (
            <div className="px-3 py-2">
              <div className="text-[10px] font-medium text-muted-foreground mb-1 uppercase tracking-wide">输入</div>
              <pre className="text-xs font-mono bg-muted/30 rounded p-2 overflow-x-auto whitespace-pre-wrap max-h-48 overflow-y-auto">
                {parsedInput}
              </pre>
            </div>
          )}
          {result && (
            <div className="px-3 py-2 border-t border-border/30">
              <div className="text-[10px] font-medium text-muted-foreground mb-1 uppercase tracking-wide">结果</div>
              <pre className="text-xs font-mono bg-muted/30 rounded p-2 overflow-x-auto whitespace-pre-wrap max-h-48 overflow-y-auto">
                {result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
