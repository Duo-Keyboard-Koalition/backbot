"use client"

import { useState } from "react"
import CodeMirror from "@uiw/react-codemirror"
import { javascript } from "@codemirror/lang-javascript"
import { html } from "@codemirror/lang-html"
import { css } from "@codemirror/lang-css"
import { json } from "@codemirror/lang-json"
import { markdown } from "@codemirror/lang-markdown"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"
import { xcodeLight } from "@uiw/codemirror-theme-xcode"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Download, Copy, Check } from "lucide-react"

const languageOptions = [
  { value: "javascript", label: "JavaScript", extension: "js", language: javascript },
  { value: "html", label: "HTML", extension: "html", language: html },
  { value: "css", label: "CSS", extension: "css", language: css },
  { value: "json", label: "JSON", extension: "json", language: json },
  { value: "markdown", label: "Markdown", extension: "md", language: markdown },
]

export default function CodeEditor() {
  const [code, setCode] = useState("// Start coding here")
  const [language, setLanguage] = useState(languageOptions[0])
  const [darkMode, setDarkMode] = useState(true)
  const [copied, setCopied] = useState(false)

  const handleLanguageChange = (value: string) => {
    const selectedLanguage = languageOptions.find((lang) => lang.value === value)
    if (selectedLanguage) {
      setLanguage(selectedLanguage)
    }
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    const blob = new Blob([code], { type: "text/plain" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `code.${language.extension}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader className="space-y-1">
        <div className="flex items-center justify-between">
          <CardTitle className="text-2xl">Code Editor</CardTitle>
          <div className="flex items-center space-x-2">
            <Label htmlFor="dark-mode" className="text-sm">
              Dark Mode
            </Label>
            <Switch id="dark-mode" checked={darkMode} onCheckedChange={setDarkMode} />
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Label htmlFor="language-select" className="text-sm">
            Language:
          </Label>
          <Select value={language.value} onValueChange={handleLanguageChange}>
            <SelectTrigger id="language-select" className="w-[180px]">
              <SelectValue placeholder="Select language" />
            </SelectTrigger>
            <SelectContent>
              {languageOptions.map((lang) => (
                <SelectItem key={lang.value} value={lang.value}>
                  {lang.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent>
        <div className="border rounded-md overflow-hidden">
          <CodeMirror
            value={code}
            height="400px"
            theme={darkMode ? vscodeDark : xcodeLight}
            extensions={[language.language()]}
            onChange={(value) => setCode(value)}
            className="text-sm"
          />
        </div>
      </CardContent>
      <CardFooter className="flex justify-between">
        <div className="text-sm text-muted-foreground">{code.length} characters</div>
        <div className="flex space-x-2">
          <Button variant="outline" size="sm" onClick={handleCopy}>
            {copied ? <Check className="h-4 w-4 mr-1" /> : <Copy className="h-4 w-4 mr-1" />}
            {copied ? "Copied" : "Copy"}
          </Button>
          <Button variant="default" size="sm" onClick={handleDownload}>
            <Download className="h-4 w-4 mr-1" />
            Download
          </Button>
        </div>
      </CardFooter>
    </Card>
  )
}
