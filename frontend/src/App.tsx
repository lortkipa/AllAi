import { Plus, RefreshCw } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Brand } from './components/Brand'
import { Composer } from './components/Composer'
import { EmptyState } from './components/EmptyState'
import { MessageList } from './components/MessageList'
import { ModelPicker } from './components/ModelPicker'
import { getModels, streamChat } from './lib/api'
import type { Message, Model, StreamEvent } from './types'

export default function App() {
  const [models, setModels] = useState<Model[]>([])
  const [selectedModel, setSelectedModel] = useState<Model>()
  const [messages, setMessages] = useState<Message[]>([])
  const [draft, setDraft] = useState('')
  const [configured, setConfigured] = useState(true)
  const [modelsLoading, setModelsLoading] = useState(true)
  const [modelError, setModelError] = useState('')
  const [streaming, setStreaming] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  const loadModels = useCallback(async () => {
    setModelsLoading(true)
    setModelError('')
    try {
      const response = await getModels()
      setModels(response.models)
      setConfigured(response.configured)
      setSelectedModel((current) =>
        response.models.find((model) => model.id === current?.id) ?? response.models[0],
      )
    } catch (error) {
      setModelError(error instanceof Error ? error.message : 'Could not load models.')
    } finally {
      setModelsLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    getModels(controller.signal)
      .then((response) => {
        setModels(response.models)
        setConfigured(response.configured)
        setSelectedModel(response.models[0])
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setModelError(error instanceof Error ? error.message : 'Could not load models.')
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) setModelsLoading(false)
      })

    return () => {
      controller.abort()
      abortRef.current?.abort()
    }
  }, [])

  const runChat = useCallback(async (history: Message[], model: Model) => {
    const assistantId = crypto.randomUUID()
    const controller = new AbortController()
    abortRef.current = controller
    setStreaming(true)
    setMessages([...history, {
      id: assistantId,
      role: 'assistant',
      content: '',
      model,
      status: 'streaming',
    }])

    const applyEvent = (event: StreamEvent) => {
      setMessages((current) => current.map((message) => {
        if (message.id !== assistantId) return message
        if (event.type === 'delta') return { ...message, content: message.content + (event.content ?? '') }
        if (event.type === 'done') return { ...message, status: 'complete' }
        return { ...message, content: event.message ?? 'The response stopped unexpectedly.', status: 'error' }
      }))
    }

    try {
      await streamChat({ model: model.id, messages: history, signal: controller.signal, onEvent: applyEvent })
      setMessages((current) => current.map((message) =>
        message.id === assistantId && message.status === 'streaming'
          ? { ...message, status: 'complete' }
          : message,
      ))
    } catch (error) {
      if (controller.signal.aborted) {
        setMessages((current) => current.map((message) =>
          message.id === assistantId ? { ...message, status: 'stopped' } : message,
        ))
      } else {
        const errorMessage = error instanceof Error ? error.message : 'The model could not respond.'
        setMessages((current) => current.map((message) =>
          message.id === assistantId ? { ...message, content: errorMessage, status: 'error' } : message,
        ))
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }, [])

  const send = useCallback(() => {
    const content = draft.trim()
    if (!content || !selectedModel || streaming || !configured) return
    const userMessage: Message = { id: crypto.randomUUID(), role: 'user', content }
    const validHistory = messages.filter((message) =>
      message.role === 'user' || message.status === 'complete',
    )
    const history = [...validHistory, userMessage]
    setDraft('')
    void runChat(history, selectedModel)
  }, [configured, draft, messages, runChat, selectedModel, streaming])

  const retry = useCallback((messageId: string) => {
    if (streaming) return
    const assistantIndex = messages.findIndex((message) => message.id === messageId)
    const failed = messages[assistantIndex]
    if (assistantIndex < 0 || !failed.model) return
    void runChat(messages.slice(0, assistantIndex), failed.model)
  }, [messages, runChat, streaming])

  const startNewChat = () => {
    abortRef.current?.abort()
    setMessages([])
    setDraft('')
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <Brand />
        <div className="topbar-actions">
          {messages.length > 0 ? (
            <button className="new-chat" type="button" onClick={startNewChat}>
              <Plus size={16} /> New chat
            </button>
          ) : null}
          <ModelPicker
            models={models}
            selected={selectedModel}
            loading={modelsLoading}
            disabled={streaming}
            onSelect={setSelectedModel}
          />
        </div>
      </header>

      <main className={`chat-main ${messages.length === 0 ? 'chat-main-empty' : ''}`}>
        {modelError ? (
          <div className="catalog-error" role="alert">
            <span>{modelError}</span>
            <button type="button" onClick={() => void loadModels()}><RefreshCw size={14} /> Retry</button>
          </div>
        ) : null}

        {messages.length === 0 ? (
          <EmptyState configured={configured} onPrompt={setDraft} />
        ) : (
          <MessageList messages={messages} onRetry={retry} />
        )}
      </main>

      <footer className="composer-dock">
        <Composer
          value={draft}
          disabled={!configured || modelsLoading || !selectedModel}
          streaming={streaming}
          modelName={selectedModel?.name}
          onChange={setDraft}
          onSend={send}
          onStop={() => abortRef.current?.abort()}
        />
      </footer>
    </div>
  )
}
