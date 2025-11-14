import React, { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Send, Copy, Check, Settings2 } from 'lucide-react'
import { kthuluApi } from '@/services/kthuluApi'

interface Message {
  id: string
  type: 'user' | 'assistant'
  content: string
  timestamp: Date
  model?: string
  provider?: string
  usage?: {
    prompt_tokens?: number
    completion_tokens?: number
    total_tokens?: number
  }
}

interface AIProvider {
  id: string
  name: string
  enabled: boolean
}

export const AIChat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [providers, setProviders] = useState<AIProvider[]>([])
  const [selectedProvider, setSelectedProvider] = useState('litellm')
  const [models, setModels] = useState<string[]>([])
  const [selectedModel, setSelectedModel] = useState('gpt-4')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Load providers and models on mount
  useEffect(() => {
    loadProviders()
    loadModels()
  }, [])

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const loadProviders = async () => {
    try {
      const response = await kthuluApi.getAIProviders()
      setProviders(response.providers)
    } catch (err) {
      console.error('Failed to load providers:', err)
    }
  }

  const loadModels = async () => {
    try {
      const response = await kthuluApi.getAIProviders()
      // For now, show default models
      // TODO: Get actual models from API
      setModels([
        'gpt-4',
        'gpt-4-turbo',
        'gpt-3.5-turbo',
        'claude-3-opus',
        'claude-3-sonnet',
      ])
    } catch (err) {
      console.error('Failed to load models:', err)
    }
  }

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || loading) return

    // Add user message
    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: input,
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setLoading(true)
    setError('')

    try {
      const response = await kthuluApi.suggestAI({
        prompt: input,
        include_context: true,
        project_path: '.',
        model: selectedModel,
        provider: selectedProvider,
      })

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: response.result,
        timestamp: new Date(),
        model: response.model,
        provider: response.provider,
        usage: response.usage,
      }

      setMessages((prev) => [...prev, assistantMessage])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get response')
      console.error('AI Chat error:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = (content: string) => {
    navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="w-full h-full flex flex-col bg-kthulu-surface1">
      <Card className="flex-1 flex flex-col h-full border-primary/20">
        <CardHeader className="border-b border-primary/20">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">🤖 AI Assistant Chat</CardTitle>
              <CardDescription className="text-xs">
                Provider: {selectedProvider} • Model: {selectedModel}
              </CardDescription>
            </div>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setShowSettings(!showSettings)}
              className="hover:bg-primary/10"
            >
              <Settings2 className="h-4 w-4" />
            </Button>
          </div>

          {/* Settings Panel */}
          {showSettings && (
            <div className="mt-4 p-3 bg-kthulu-surface2 rounded border border-primary/20 space-y-3">
              <div>
                <label className="text-sm font-medium block mb-2">Provider</label>
                <select
                  value={selectedProvider}
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  className="w-full px-3 py-2 bg-kthulu-surface1 border border-primary/30 rounded text-sm"
                >
                  {providers.map((p) => (
                    <option
                      key={p.id}
                      value={p.id}
                      disabled={!p.enabled}
                    >
                      {p.name} {!p.enabled ? '(disabled)' : ''}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium block mb-2">Model</label>
                <select
                  value={selectedModel}
                  onChange={(e) => setSelectedModel(e.target.value)}
                  className="w-full px-3 py-2 bg-kthulu-surface1 border border-primary/30 rounded text-sm"
                >
                  {models.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          )}
        </CardHeader>

        {/* Messages Area */}
        <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 && (
            <div className="h-full flex items-center justify-center text-center">
              <div className="space-y-2">
                <div className="text-4xl">🤖</div>
                <p className="text-muted-foreground">Start a conversation...</p>
              </div>
            </div>
          )}

          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-xs lg:max-w-md p-3 rounded-lg ${
                  msg.type === 'user'
                    ? 'bg-gradient-cyber text-white'
                    : 'bg-kthulu-surface2 border border-primary/20 text-foreground'
                }`}
              >
                <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                {msg.usage && (
                  <p className="text-xs opacity-70 mt-1 font-mono">
                    Tokens: {msg.usage.total_tokens}
                  </p>
                )}
                <p className="text-xs opacity-70 mt-1">
                  {msg.timestamp.toLocaleTimeString()}
                </p>

                {msg.type === 'assistant' && (
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleCopy(msg.content)}
                    className="mt-2 h-auto p-1 text-xs hover:bg-primary/10"
                  >
                    {copied ? (
                      <>
                        <Check className="h-3 w-3 mr-1" />
                        Copied
                      </>
                    ) : (
                      <>
                        <Copy className="h-3 w-3 mr-1" />
                        Copy
                      </>
                    )}
                  </Button>
                )}
              </div>
            </div>
          ))}

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
              Error: {error}
            </div>
          )}

          {loading && (
            <div className="flex justify-start">
              <div className="p-3 bg-kthulu-surface2 border border-primary/20 rounded-lg">
                <Loader2 className="h-4 w-4 animate-spin" />
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </CardContent>

        {/* Input Area */}
        <div className="border-t border-primary/20 p-4">
          <form onSubmit={handleSendMessage} className="flex gap-2">
            <Input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Type your message..."
              disabled={loading}
              className="flex-1 bg-kthulu-surface2 border-primary/30"
            />
            <Button
              type="submit"
              disabled={loading || !input.trim()}
              className="bg-gradient-cyber hover:opacity-90"
              size="sm"
            >
              {loading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </form>
        </div>
      </Card>
    </div>
  )
}
