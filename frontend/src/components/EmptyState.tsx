import { ArrowUpRight, KeyRound } from 'lucide-react'

type EmptyStateProps = {
  configured: boolean
  onPrompt: (prompt: string) => void
}

const prompts = [
  ['Make it clear', 'Explain why the sky changes color at sunset.'],
  ['Think it through', 'Plan a focused two-hour study session.'],
  ['Start creating', 'Write an opening line for a quiet sci-fi story.'],
]

export function EmptyState({ configured, onPrompt }: EmptyStateProps) {
  if (!configured) {
    return (
      <section className="setup-state">
        <span className="setup-icon"><KeyRound size={22} /></span>
        <p className="eyebrow">One step to begin</p>
        <h1>Give Allai a key<br />to the model room.</h1>
        <p>Add your OpenRouter API key to the project’s <code>.env</code> file, then restart the server.</p>
        <div className="env-snippet"><span>OPENROUTER_API_KEY</span><b>=</b><em>your_key_here</em></div>
      </section>
    )
  }

  return (
    <section className="empty-state">
      <p className="eyebrow">One conversation · many minds</p>
      <h1>What do you want<br />to think through?</h1>
      <p className="empty-intro">Choose a free model, ask anything, then switch models whenever another perspective would help.</p>
      <div className="prompt-grid">
        {prompts.map(([label, prompt]) => (
          <button type="button" key={prompt} onClick={() => onPrompt(prompt)}>
            <span>{label}</span>
            <p>{prompt}</p>
            <ArrowUpRight size={17} aria-hidden="true" />
          </button>
        ))}
      </div>
    </section>
  )
}

