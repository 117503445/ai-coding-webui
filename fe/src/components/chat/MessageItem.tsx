import { User, Bot } from 'lucide-react'
import type { ChatMessage, ContentBlock } from '@/store/chatStore'
import { MarkdownRenderer } from './MarkdownRenderer'
import { ThinkingBlock } from './ThinkingBlock'
import { ToolCallBlock } from './ToolCallBlock'

interface MessageItemProps {
  message: ChatMessage
  isStreaming?: boolean
}

function renderBlock(block: ContentBlock, index: number, allBlocks: ContentBlock[], isStreaming?: boolean) {
  switch (block.type) {
    case 'thinking':
      return (
        <ThinkingBlock
          key={index}
          content={block.content}
          isStreaming={isStreaming}
        />
      )
    case 'text':
      return (
        <MarkdownRenderer key={index} content={block.content} />
      )
    case 'tool_use': {
      const resultBlock = allBlocks.find(
        b => b.type === 'tool_result' && b.toolUseId === block.id
      )
      return (
        <ToolCallBlock
          key={index}
          name={block.name}
          input={block.input}
          result={resultBlock?.type === 'tool_result' ? resultBlock.content : undefined}
          isStreaming={isStreaming && !resultBlock}
        />
      )
    }
    case 'tool_result':
      return null
    default:
      return null
  }
}

export function MessageItem({ message, isStreaming }: MessageItemProps) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex gap-3 px-4 py-4 ${isUser ? '' : 'bg-muted/20'}`}>
      <div className={`w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5 ${
        isUser ? 'bg-primary/10 text-primary' : 'bg-accent/10 text-accent'
      }`}>
        {isUser ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
      </div>
      <div className="flex-1 min-w-0 overflow-hidden">
        <div className="text-xs font-medium text-muted-foreground mb-1">
          {isUser ? '你' : 'Claude'}
        </div>
        <div className="space-y-1">
          {message.blocks.map((block, i) =>
            renderBlock(block, i, message.blocks, isStreaming && i === message.blocks.length - 1)
          )}
        </div>
      </div>
    </div>
  )
}
