export type Model = {
  id: string
  name: string
  provider: string
  description?: string
  contextLength?: number
}

export type Message = {
  id: string
  role: 'user' | 'assistant'
  content: string
  model?: Model
  status?: 'streaming' | 'complete' | 'error' | 'stopped'
}

export type ModelsResponse = {
  models: Model[]
  configured: boolean
}

export type StreamEvent = {
  type: 'delta' | 'done' | 'error'
  content?: string
  message?: string
}

