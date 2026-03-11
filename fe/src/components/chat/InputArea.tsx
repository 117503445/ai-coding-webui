import { useState, useRef, useCallback, useEffect } from 'react'
import { Send, Square, Slash } from 'lucide-react'
import type { WorkStatus } from '@/store/chatStore'

const SLASH_COMMANDS = [
  { command: '/new', description: '开始新的对话' },
  { command: '/clear', description: '清除当前对话' },
  { command: '/compact', description: '压缩对话上下文' },
  { command: '/cost', description: '显示 token 用量' },
  { command: '/help', description: '显示可用命令' },
]

interface InputAreaProps {
  onSend: (content: string) => void
  onCommand: (command: string) => void
  onAbort: () => void
  workStatus: WorkStatus
  disabled: boolean
}

export function InputArea({ onSend, onCommand, onAbort, workStatus, disabled }: InputAreaProps) {
  const [input, setInput] = useState('')
  const [showCommands, setShowCommands] = useState(false)
  const [commandFilter, setCommandFilter] = useState('')
  const [selectedIdx, setSelectedIdx] = useState(0)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const isWorking = workStatus === 'working'

  const filteredCommands = SLASH_COMMANDS.filter(
    c => c.command.includes(commandFilter.toLowerCase())
  )

  useEffect(() => {
    setSelectedIdx(0)
  }, [commandFilter])

  const handleInputChange = useCallback((value: string) => {
    setInput(value)
    if (value.startsWith('/')) {
      setShowCommands(true)
      setCommandFilter(value)
    } else {
      setShowCommands(false)
      setCommandFilter('')
    }
  }, [])

  const handleSend = useCallback(() => {
    const trimmed = input.trim()
    if (!trimmed) return

    if (trimmed.startsWith('/')) {
      const matched = SLASH_COMMANDS.find(c => c.command === trimmed)
      if (matched) {
        onCommand(matched.command)
        setInput('')
        setShowCommands(false)
        return
      }
    }

    onSend(trimmed)
    setInput('')
    setShowCommands(false)
  }, [input, onSend, onCommand])

  const selectCommand = useCallback((command: string) => {
    onCommand(command)
    setInput('')
    setShowCommands(false)
    textareaRef.current?.focus()
  }, [onCommand])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (showCommands && filteredCommands.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedIdx(i => (i + 1) % filteredCommands.length)
        return
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedIdx(i => (i - 1 + filteredCommands.length) % filteredCommands.length)
        return
      }
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        selectCommand(filteredCommands[selectedIdx].command)
        return
      }
      if (e.key === 'Escape') {
        setShowCommands(false)
        return
      }
    }

    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!isWorking && !disabled) {
        handleSend()
      }
    }
  }, [showCommands, filteredCommands, selectedIdx, isWorking, disabled, handleSend, selectCommand])

  const autoResize = useCallback((el: HTMLTextAreaElement) => {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 200) + 'px'
  }, [])

  return (
    <div className="border-t border-border bg-background/80 backdrop-blur-sm">
      <div className="max-w-3xl mx-auto px-4 py-3 relative">
        {showCommands && filteredCommands.length > 0 && (
          <div className="absolute bottom-full left-4 right-4 mb-1 bg-card border border-border rounded-lg shadow-lg overflow-hidden z-10">
            {filteredCommands.map((cmd, i) => (
              <button
                key={cmd.command}
                onClick={() => selectCommand(cmd.command)}
                className={`w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-muted/50 transition-colors ${
                  i === selectedIdx ? 'bg-muted/50' : ''
                }`}
              >
                <Slash className="w-3.5 h-3.5 text-muted-foreground flex-shrink-0" />
                <span className="font-mono font-medium">{cmd.command}</span>
                <span className="text-muted-foreground text-xs">{cmd.description}</span>
              </button>
            ))}
          </div>
        )}

        <div className="flex items-end gap-2">
          <div className="flex-1 relative">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => {
                handleInputChange(e.target.value)
                autoResize(e.target)
              }}
              onKeyDown={handleKeyDown}
              placeholder={isWorking ? 'Claude 正在工作中...' : '发送消息，或输入 / 查看命令...'}
              disabled={disabled || isWorking}
              rows={1}
              className="w-full resize-none rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 disabled:opacity-50 transition-all placeholder:text-muted-foreground/50"
            />
          </div>

          {isWorking ? (
            <button
              onClick={onAbort}
              className="flex-shrink-0 w-9 h-9 rounded-lg bg-destructive text-destructive-foreground flex items-center justify-center hover:opacity-90 transition-opacity"
              title="终止"
            >
              <Square className="w-4 h-4" />
            </button>
          ) : (
            <button
              onClick={handleSend}
              disabled={disabled || !input.trim()}
              className="flex-shrink-0 w-9 h-9 rounded-lg bg-primary text-primary-foreground flex items-center justify-center hover:opacity-90 transition-opacity disabled:opacity-30"
              title="发送"
            >
              <Send className="w-4 h-4" />
            </button>
          )}
        </div>

        <div className="text-[10px] text-muted-foreground/50 mt-1.5 text-center">
          Enter 发送 · Shift+Enter 换行 · 输入 / 查看命令
        </div>
      </div>
    </div>
  )
}
