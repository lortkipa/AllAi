import { Check, ChevronDown, Search, Sparkles } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { Model } from '../types'

type ModelPickerProps = {
  models: Model[]
  selected?: Model
  loading: boolean
  disabled?: boolean
  onSelect: (model: Model) => void
}

export function ModelPicker({ models, selected, loading, disabled, onSelect }: ModelPickerProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const close = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [open])

  const filteredModels = useMemo(() => {
    const normalized = query.trim().toLowerCase()
    if (!normalized) return models
    return models.filter((model) =>
      `${model.name} ${model.provider} ${model.id}`.toLowerCase().includes(normalized),
    )
  }, [models, query])

  const choose = (model: Model) => {
    onSelect(model)
    setOpen(false)
    setQuery('')
  }

  return (
    <div className="model-picker" ref={containerRef}>
      <button
        className="model-trigger"
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        disabled={disabled || loading || models.length === 0}
        onClick={() => setOpen((value) => !value)}
      >
        <span className="model-trigger-icon"><Sparkles size={15} strokeWidth={1.8} /></span>
        <span className="model-trigger-copy">
          <span className="model-trigger-label">Current model</span>
          <strong>{loading ? 'Finding free models…' : selected?.name ?? 'Choose a model'}</strong>
        </span>
        <ChevronDown size={17} className={open ? 'chevron-open' : ''} />
      </button>

      {open ? (
        <div className="model-menu" role="dialog" aria-label="Choose a model">
          <div className="model-search">
            <Search size={16} aria-hidden="true" />
            <input
              autoFocus
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search free models"
              aria-label="Search free models"
            />
          </div>
          <div className="model-list" role="listbox" aria-label="Free models">
            {filteredModels.map((model) => (
              <button
                type="button"
                role="option"
                aria-selected={model.id === selected?.id}
                className="model-option"
                key={model.id}
                onClick={() => choose(model)}
              >
                <span className="provider-monogram" aria-hidden="true">
                  {model.provider.slice(0, 2).toUpperCase()}
                </span>
                <span className="model-option-copy">
                  <strong>{model.name}</strong>
                  <small>{model.provider}{model.contextLength ? ` · ${formatContext(model.contextLength)} context` : ''}</small>
                </span>
                {model.id === selected?.id ? <Check size={17} aria-hidden="true" /> : null}
              </button>
            ))}
            {filteredModels.length === 0 ? (
              <p className="model-empty">No free models match “{query}”.</p>
            ) : null}
          </div>
          <div className="model-menu-footer">Free models via OpenRouter</div>
        </div>
      ) : null}
    </div>
  )
}

function formatContext(tokens: number) {
  return tokens >= 1_000_000
    ? `${Math.round(tokens / 100_000) / 10}m`
    : `${Math.round(tokens / 1000)}k`
}

