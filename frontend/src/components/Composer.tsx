import { ArrowUp, Square } from 'lucide-react'
import { useEffect, useRef } from 'react'

type ComposerProps = {
  value: string
  disabled: boolean
  streaming: boolean
  modelName?: string
  onChange: (value: string) => void
  onSend: () => void
  onStop: () => void
}

export function Composer({ value, disabled, streaming, modelName, onChange, onSend, onStop }: ComposerProps) {
  const textAreaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    const element = textAreaRef.current
    if (!element) return
    element.style.height = 'auto'
    element.style.height = `${Math.min(element.scrollHeight, 180)}px`
  }, [value])

  return (
    <div className="composer-wrap">
      <div className={`composer ${streaming ? 'composer-streaming' : ''}`}>
        <textarea
          ref={textAreaRef}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter' && !event.shiftKey) {
              event.preventDefault()
              if (!disabled && value.trim()) onSend()
            }
          }}
          placeholder={disabled ? 'Add your OpenRouter key to begin' : `Message ${modelName ?? 'a model'}…`}
          aria-label="Message"
          rows={1}
          disabled={disabled}
        />
        {streaming ? (
          <button className="send-button stop-button" type="button" onClick={onStop} aria-label="Stop response">
            <Square size={14} fill="currentColor" />
          </button>
        ) : (
          <button className="send-button" type="button" onClick={onSend} disabled={disabled || !value.trim()} aria-label="Send message">
            <ArrowUp size={19} strokeWidth={2.2} />
          </button>
        )}
      </div>
      <p className="composer-note">Free models can be slow or rate-limited. Check important answers.</p>
    </div>
  )
}

