import { useRef, type ChangeEvent, type ReactNode } from 'react'

const jsonTokenPattern = /("(?:\\.|[^"\\])*"(?=\s*:)|"(?:\\.|[^"\\])*"|-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b|\b(?:true|false|null)\b|[{}[\],:])/g

function highlightJson(json: string) {
  const tokens: ReactNode[] = []
  let lastIndex = 0

  for (const match of json.matchAll(jsonTokenPattern)) {
    const token = match[0]
    const {index} = match
    if (index > lastIndex) tokens.push(json.slice(lastIndex, index))

    let className = 'json-token-punctuation'
    if (token.startsWith('"')) {
      className = json.slice(index + token.length).match(/^\s*:/) ? 'json-token-key' : 'json-token-string'
    } else if (/^(?:true|false|null)$/.test(token)) {
      className = 'json-token-literal'
    } else if (/^-?\d/.test(token)) {
      className = 'json-token-number'
    }
    tokens.push(<span className={className} key={index}>{token}</span>)
    lastIndex = index + token.length
  }

  if (lastIndex < json.length) tokens.push(json.slice(lastIndex))
  return tokens
}

export function JsonCodeEditor({ id, value, onChange, describedBy }: {
  id: string
  value: string
  onChange: (value: string) => void
  describedBy?: string
}) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const highlightRef = useRef<HTMLPreElement>(null)

  function handleChange(event: ChangeEvent<HTMLTextAreaElement>) {
    onChange(event.target.value)
  }

  function syncScroll() {
    if (!textareaRef.current || !highlightRef.current) return
    highlightRef.current.scrollTop = textareaRef.current.scrollTop
    highlightRef.current.scrollLeft = textareaRef.current.scrollLeft
  }

  return <div className="json-code-editor">
    <pre className="json-highlight-layer" aria-hidden="true" ref={highlightRef}>{highlightJson(value)}{'\n'}</pre>
    <textarea
      ref={textareaRef}
      id={id}
      className="input code-textarea json-code-input"
      spellCheck={false}
      value={value}
      onChange={handleChange}
      onScroll={syncScroll}
      aria-label="DynamoDB JSON"
      aria-describedby={describedBy}
      wrap="off"
    />
  </div>
}
