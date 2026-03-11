import { useEffect } from 'react'
import { useChatStore } from '@/store/chatStore'
import { StatusBar } from './StatusBar'
import { MessageList } from './MessageList'
import { InputArea } from './InputArea'
import { ConnectionOverlay } from './ConnectionOverlay'

export function ChatContainer() {
  const {
    connectionState,
    workStatus,
    workDetail,
    messages,
    stream,
    init,
    cleanup,
    sendMessage,
    sendCommand,
    abort,
    newChat,
  } = useChatStore()

  useEffect(() => {
    init()
    return () => cleanup()
  }, [init, cleanup])

  return (
    <div className="h-dvh flex flex-col relative bg-background">
      <StatusBar
        connectionState={connectionState}
        workStatus={workStatus}
        workDetail={workDetail}
        onNewChat={newChat}
      />

      <MessageList
        messages={messages}
        streamingBlocks={stream.currentBlocks}
        isWorking={workStatus === 'working'}
      />

      <InputArea
        onSend={sendMessage}
        onCommand={sendCommand}
        onAbort={abort}
        workStatus={workStatus}
        disabled={connectionState !== 'connected'}
      />

      <ConnectionOverlay state={connectionState} />
    </div>
  )
}
