import type { Components } from 'react-markdown'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

const components: Components = {
  a: ({ children, ...props }) => <a {...props} target="_blank" rel="noreferrer">{children}</a>,
  pre: ({ children }) => <pre className="code-block">{children}</pre>,
}

export function MessageContent({ content }: { content: string }) {
  return (
    <div className="message-markdown">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>{content}</ReactMarkdown>
    </div>
  )
}

