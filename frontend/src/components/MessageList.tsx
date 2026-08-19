import { AlertCircle, Bot, Copy, RotateCcw } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import type { Message } from '../types'
import { MessageContent } from './MessageContent'

type MessageListProps = {
  messages: Message[]
  onRetry: (messageId: string) => void
}

export function MessageList({ messages, onRetry }: MessageListProps) {
  const endRef = useRef<HTMLDivElement>(null)
  const [copied, setCopied] = useState<string>()

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [messages])

  const copy = async (message: Message) => {
    await navigator.clipboard.writeText(message.content)
    setCopied(message.id)
    window.setTimeout(() => setCopied(undefined), 1400)
  }

  return (
    <div className="message-list" aria-live="polite">
      {messages.map((message) => (
        <article className={`message message-${message.role}`} key={message.id}>
          {message.role === 'assistant' ? (
            <div className="model-rail" aria-hidden="true">
              <span><Bot size={15} /></span>
              <i />
            </div>
          ) : null}

          <div className="message-body">
            <div className="message-meta">
              <strong>{message.role === 'user' ? 'You' : message.model?.name ?? 'Assistant'}</strong>
              {message.role === 'assistant' ? <span>{message.model?.provider}</span> : null}
            </div>

            {message.status === 'error' ? (
              <div className="message-error"><AlertCircle size={16} /> {message.content}</div>
            ) : message.status === 'streaming' && !message.content ? (
              <div className="thinking" aria-label="Model is thinking"><span /><span /><span /></div>
            ) : (
              <MessageContent content={message.content} />
            )}

            {message.role === 'assistant' && message.content ? (
              <div className="message-actions">
                <button type="button" onClick={() => void copy(message)} aria-label="Copy response">
                  <Copy size={14} /> {copied === message.id ? 'Copied' : 'Copy'}
                </button>
                {message.status === 'error' || message.status === 'stopped' ? (
                  <button type="button" onClick={() => onRetry(message.id)}>
                    <RotateCcw size={14} /> Try again
                  </button>
                ) : null}
              </div>
            ) : null}
          </div>
        </article>
      ))}
      <div ref={endRef} />
    </div>
  )
}
