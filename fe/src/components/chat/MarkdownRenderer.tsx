import { useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Copy, Check } from 'lucide-react'
import { useState } from 'react'

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [text])

  return (
    <button
      onClick={handleCopy}
      className="absolute top-2 right-2 p-1.5 rounded-md bg-white/10 hover:bg-white/20 transition-colors text-gray-400 hover:text-gray-200"
      title="复制代码"
    >
      {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
    </button>
  )
}

export function MarkdownRenderer({ content }: { content: string }) {
  return (
    <div className="prose prose-sm max-w-none dark:prose-invert prose-pre:p-0 prose-pre:bg-transparent">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        children={content}
        components={{
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '')
            const codeStr = String(children).replace(/\n$/, '')

            if (match) {
              return (
                <div className="relative group my-2 rounded-lg overflow-hidden">
                  <div className="flex items-center justify-between px-4 py-1.5 bg-[#1e1e2e] text-xs text-gray-400 border-b border-white/5">
                    <span>{match[1]}</span>
                  </div>
                  <div className="relative">
                    <CopyButton text={codeStr} />
                    <SyntaxHighlighter
                      style={oneDark}
                      language={match[1]}
                      PreTag="div"
                      customStyle={{
                        margin: 0,
                        borderRadius: 0,
                        fontSize: '0.8rem',
                      }}
                    >
                      {codeStr}
                    </SyntaxHighlighter>
                  </div>
                </div>
              )
            }

            return (
              <code className="px-1.5 py-0.5 rounded bg-primary/10 text-primary text-xs font-mono" {...props}>
                {children}
              </code>
            )
          },
          table({ children }) {
            return (
              <div className="overflow-x-auto my-2">
                <table className="min-w-full border-collapse text-sm">{children}</table>
              </div>
            )
          },
        }}
      />
    </div>
  )
}
