import { useEffect, useRef } from 'react'
import type { ChatMessage, ContentBlock } from '@/store/chatStore'
import { MessageItem } from './MessageItem'
import { Bot } from 'lucide-react'

interface MessageListProps {
  messages: ChatMessage[]
  streamingBlocks: ContentBlock[]
  isWorking: boolean
}

export function MessageList({ messages, streamingBlocks, isWorking }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = bottomRef.current
    if (el) {
      el.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, streamingBlocks])

  if (messages.length === 0 && streamingBlocks.length === 0 && !isWorking) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto mb-4">
            <Bot className="w-8 h-8 text-primary" />
          </div>
          <h2 className="text-xl font-semibold mb-2">Claude Code WebUI</h2>
          <p className="text-muted-foreground text-sm">
            发送消息开始对话。支持多轮交互，输入 <code className="px-1 py-0.5 bg-muted rounded text-xs">/</code> 查看可用命令。
          </p>
        </div>
      </div>
    )
  }

  const streamMsg: ChatMessage | null = streamingBlocks.length > 0
    ? { role: 'assistant', blocks: streamingBlocks, timestamp: Date.now() }
    : null

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto">
      <div className="max-w-3xl mx-auto divide-y divide-border/30">
        {messages.map((msg, i) => (
          <MessageItem key={i} message={msg} />
        ))}
        {streamMsg && (
          <MessageItem message={streamMsg} isStreaming />
        )}
        {isWorking && streamingBlocks.length === 0 && (
          <div className="flex gap-3 px-4 py-4 bg-muted/20">
            <div className="w-7 h-7 rounded-lg bg-accent/10 text-accent flex items-center justify-center flex-shrink-0">
              <Bot className="w-4 h-4" />
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <div className="flex gap-1">
                <span className="w-1.5 h-1.5 bg-primary rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="w-1.5 h-1.5 bg-primary rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="w-1.5 h-1.5 bg-primary rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
              <span>Claude 正在思考...</span>
            </div>
          </div>
        )}
      </div>
      <div ref={bottomRef} className="h-4" />
    </div>
  )
}
