import type { Message, ModelsResponse, StreamEvent } from '../types'

export async function getModels(signal?: AbortSignal): Promise<ModelsResponse> {
  const response = await fetch('/api/models', { signal })
  if (!response.ok) {
    throw new Error(await readError(response))
  }
  return response.json() as Promise<ModelsResponse>
}

type StreamChatOptions = {
  model: string
  messages: Message[]
  signal: AbortSignal
  onEvent: (event: StreamEvent) => void
}

export async function streamChat({ model, messages, signal, onEvent }: StreamChatOptions) {
  const response = await fetch('/api/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      model,
      messages: messages.map(({ role, content }) => ({ role, content })),
    }),
    signal,
  })

  if (!response.ok) {
    throw new Error(await readError(response))
  }
  if (!response.body) {
    throw new Error('The browser could not read the model response stream.')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })
    const blocks = buffer.replaceAll('\r\n', '\n').split('\n\n')
    buffer = blocks.pop() ?? ''

    for (const block of blocks) {
      const dataLine = block.split('\n').find((line) => line.startsWith('data:'))
      if (!dataLine) continue
      const event = JSON.parse(dataLine.slice(5).trim()) as StreamEvent
      onEvent(event)
    }

    if (done) break
  }
}

async function readError(response: Response) {
  try {
    const payload = (await response.json()) as { error?: string }
    return payload.error || `Request failed with status ${response.status}.`
  } catch {
    return `Request failed with status ${response.status}.`
  }
}

