import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Copy, Check } from 'lucide-react'

interface AIAssistantProps {
  onApply?: (suggestion: string) => void
}

export const AIAssistant: React.FC<AIAssistantProps> = ({ onApply }) => {
  const [prompt, setPrompt] = useState('')
  const [includeContext, setIncludeContext] = useState(true)
  const [suggestion, setSuggestion] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)

  const handleSuggest = async () => {
    if (!prompt.trim()) {
      setError('Please enter a prompt')
      return
    }

    setLoading(true)
    setError('')
    setSuggestion('')

    try {
      const response = await fetch('/api/v1/ai/suggest', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt,
          include_context: includeContext,
          project_path: '.',
        }),
      })

      if (!response.ok) {
        throw new Error(`API error: ${response.statusText}`)
      }

      const data = await response.json()
      setSuggestion(data.result || '')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get suggestion')
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(suggestion)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleApply = () => {
    if (onApply) {
      onApply(suggestion)
    }
  }

  return (
    <div className="w-full max-w-2xl mx-auto space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>🤖 AI Assistant</CardTitle>
          <CardDescription>Get AI-powered suggestions for your code and architecture</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Input Section */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Prompt</label>
            <Textarea
              placeholder="E.g., 'Add rate limiting to the API' or 'Optimize this database query'"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="min-h-24"
            />
          </div>

          {/* Options */}
          <div className="space-y-2">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={includeContext}
                onChange={(e) => setIncludeContext(e.target.checked)}
                className="h-4 w-4"
              />
              <span className="text-sm">Include project context</span>
            </label>
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
              {error}
            </div>
          )}

          {/* Suggestion Display */}
          {suggestion && (
            <div className="space-y-2">
              <label className="text-sm font-medium">AI Suggestion</label>
              <div className="p-3 bg-slate-50 border border-slate-200 rounded text-sm whitespace-pre-wrap font-mono">
                {suggestion}
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleCopy}
                  className="flex items-center gap-2"
                >
                  {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  {copied ? 'Copied' : 'Copy'}
                </Button>
                {onApply && (
                  <Button size="sm" onClick={handleApply}>
                    Apply
                  </Button>
                )}
              </div>
            </div>
          )}

          {/* Action Button */}
          <Button
            onClick={handleSuggest}
            disabled={loading || !prompt.trim()}
            className="w-full"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Generating...
              </>
            ) : (
              '✨ Get Suggestion'
            )}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
