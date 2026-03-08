"use client"

import CodeMirror from "@uiw/react-codemirror"
import { javascript } from "@codemirror/lang-javascript"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"

const codeExample = `// Your code snippet here
function Example() {
  const data = fetchData();
  return <div>{data}</div>;
}
`

export default function CodeSnippet() {
  return (
    <div className="rounded-lg overflow-hidden shadow-2xl shadow-purple-500/10 border border-purple-500/20">
      <div className="bg-[#1e1e1e] px-4 py-2 flex items-center justify-between">
        <div className="flex space-x-2">
          <div className="w-3 h-3 rounded-full bg-red-500"></div>
          <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
          <div className="w-3 h-3 rounded-full bg-green-500"></div>
        </div>
        <div className="text-xs text-white/60">example.jsx</div>
      </div>
      <CodeMirror
        value={codeExample}
        height="160px"
        theme={vscodeDark}
        extensions={[javascript({ jsx: true })]}
        editable={false}
        className="text-sm"
      />
    </div>
  )
}
