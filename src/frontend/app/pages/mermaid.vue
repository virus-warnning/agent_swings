<script setup lang="ts">
import type { SplitterItem } from '@nuxt/ui'
import { ref, nextTick, watch, onMounted } from 'vue'
import { Codemirror } from 'vue-codemirror'
import { EditorView } from '@codemirror/view'

const card = 'border border-default rounded-xl overflow-hidden p-[5px] bg-default'

const innerItems: SplitterItem[] = [
  { slot: 'main', minSize: 25, defaultSize: 50, class: card },
  { slot: 'right', minSize: 25, defaultSize: 50, class: card }
]

const flowchartTemplate = `flowchart TD
  A1(["👩🏻‍💻 Human"])
  A2(["🤖 Agent"])

  B1("HTTP<br>(Fiber v3)")

  subgraph backend handlers
    B2("API handler")
    B3("MCP HTTP")
    B4("Streamable Handler")
    B5("SSE handler")
    B6("MCP Server")
  end

  subgraph backend services
    C1(Mermaid to SVG)
    C2(Mermaid to PNG)
    C3(XML Formatter)
  end

  subgraph frontend
    B7("Nuxt.js")
  end

  subgraph resources
    D1(Mermaid Container<br>yuzutech/kroki-mermaid)
  end

  %% human path
  A1 --> B1
  B1 --> |the rest URI| B7
  linkStyle 0,1 stroke:#0000ff,stroke-width:2px

  %% agent path
  A2 --> B1
  B1 --> |/mcp or /sse| B3
  linkStyle 2,3 stroke:#00aa00,stroke-width:2px

  %% api path
  B7 --> |axios| B1
  B1 --> |/api| B2
  linkStyle 4,5 stroke:#9900cc,stroke-width:2px

  B3 --> |/mcp| B4
  B3 --> |/sse| B5
  B4 & B5 --> B6
  B6 --> C1 & C2 & C3
  B2 --> D1
  C1 & C2 --> D1`

const mindmapTemplate = `mindmap
  root((Swings))
    Frontend
      Nuxt.js
      Vue Codemirror
      SVG Preview
    Backend
      Fiber HTTP
      MCP Server
      Services
        Mermaid to SVG
        Mermaid to PNG
    Resources
      kroki-mermaid`

const timelineTemplate = `timeline
    title Project Timeline
    2026-09 : Init repo
           : Setup Fiber API
    2026-10 : MCP server
            : Mermaid to SVG
    2026-11 : Frontend editor
            : SVG preview & zoom`

const diagramTypes = [
  { label: '流程圖', value: 'flowchart' },
  { label: '心智圖', value: 'mindmap' },
  { label: '時間軸', value: 'timeline' }
]

const diagramType = ref('flowchart')

const templates: Record<string, string> = {
  flowchart: flowchartTemplate,
  mindmap: mindmapTemplate,
  timeline: timelineTemplate
}

const code = ref(flowchartTemplate)

watch(diagramType, (type) => {
  code.value = templates[type] ?? code.value
})

// 頁面載入完成時自動生成 SVG
onMounted(() => {
  if (code.value.trim()) generateSVG()
})

// Debounced auto-generate: 停止輸入 800ms 後自動生成
const AUTO_GEN_DELAY = 800
let autoGenTimer: ReturnType<typeof setTimeout> | undefined

watch(code, () => {
  if (autoGenTimer) clearTimeout(autoGenTimer)
  autoGenTimer = setTimeout(() => {
    if (code.value.trim()) generateSVG()
  }, AUTO_GEN_DELAY)
})

const svg = ref('')
const loading = ref(false)
const hasError = ref(false)
const translatingError = ref(false)
const fixingCode = ref(false)
const lastError = ref("")

// Zoom & pan state
const zoom = ref(1)
const tx = ref(0)
const ty = ref(0)
const dragging = ref(false)
let dragStart = { x: 0, y: 0 }

const ZOOM_STEP = 0.15
const MIN_ZOOM = 0.05
const MAX_ZOOM = 5

const canvasRef = ref<HTMLElement>()
const svgWrapperRef = ref<HTMLElement>()

function zoomIn() {
  if (hasError.value) return
  zoom.value = Math.min(MAX_ZOOM, zoom.value + ZOOM_STEP)
}
function zoomOut() {
  if (hasError.value) return
  zoom.value = Math.max(MIN_ZOOM, zoom.value - ZOOM_STEP)
}

function fitToView() {
  const canvas = canvasRef.value
  const svgEl = svgWrapperRef.value?.querySelector('svg')
  if (!canvas || !svgEl) {
    zoom.value = 1
    tx.value = 0
    ty.value = 0
    return
  }
  const svgW = (svgEl as SVGSVGElement).viewBox.baseVal.width || (svgEl as SVGElement).getBoundingClientRect().width
  const svgH = (svgEl as SVGSVGElement).viewBox.baseVal.height || (svgEl as SVGElement).getBoundingClientRect().height
  if (!svgW || !svgH) return
  const padding = 20
  const fitZoom = Math.min(
    (canvas.clientWidth - padding) / svgW,
    (canvas.clientHeight - padding) / svgH
  )
  zoom.value = Math.min(Math.max(fitZoom, MIN_ZOOM), MAX_ZOOM)
  tx.value = 0
  ty.value = 0
}

function centerView() {
  fitToView()
}

function onWheel(e: WheelEvent) {
  e.preventDefault()
  if (hasError.value) return
  const delta = e.deltaY > 0 ? -ZOOM_STEP : ZOOM_STEP
  zoom.value = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom.value + delta))
}

function onPointerDown(e: PointerEvent) {
  if (hasError.value) return
  dragging.value = true
  dragStart = { x: e.clientX - tx.value, y: e.clientY - ty.value }
  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
}
function onPointerMove(e: PointerEvent) {
  if (!dragging.value) return
  tx.value = e.clientX - dragStart.x
  ty.value = e.clientY - dragStart.y
}
function onPointerUp(e: PointerEvent) {
  dragging.value = false
  ;(e.target as HTMLElement).releasePointerCapture(e.pointerId)
}

const extensions = [
  EditorView.theme({
    '&': { height: '100%', fontSize: '14px' },
    '.cm-scroller': { overflow: 'auto' }
  })
]

function cleanError(msg: string): string {
  // 移除 kroki 錯誤前綴
  let s = msg.replace(/^kroki returned \d+:\s*/i, '').replace(/^Error \d+:\s*/i, '')
  // 從 cursor 行 (--------^) 開始截斷後續內容
  const lines = s.split('\n')
  const idx = lines.findIndex(line => /[-\s]*\^/.test(line))
  if (idx !== -1) {
    s = lines.slice(0, idx).join('\n')
  }
  return s.trim() || msg
}

async function translateError(rawError: string) {
  translatingError.value = true
  try {
    const resp = await fetch('http://localhost:9002/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt: '你是一位專業翻譯人員。將錯誤訊息翻譯為繁體中文。只翻譯，不添加解釋。直接輸出翻譯結果。',
        content: rawError,
        model: 'chat'
      })
    })
    const data = await resp.json()
    if (resp.ok && !data.error && data.content) {
      svg.value = `<pre class="text-red-500 text-sm p-2 whitespace-pre-wrap">${data.content}</pre>`
    }
  } catch {
    // keep the raw error if translation fails
  } finally {
    translatingError.value = false
  }
}

async function letAIFix() {
  if (!code.value.trim() || !lastError.value) return
  fixingCode.value = true
  try {
    const resp = await fetch('http://localhost:9002/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt: '你是一位 Mermaid 語法專家。使用者會給你一段有語法錯誤的 Mermaid 程式碼和錯誤訊息。請修正錯誤並只輸出修正後的完整 Mermaid 程式碼，不要包含任何解釋、備註或 markdown 圍欄。',
        content: '程式碼:\n' + code.value + '\n\n錯誤訊息: ' + lastError.value,
        model: 'chat'
      })
    })
    const data = await resp.json()
    if (resp.ok && !data.error && data.content) {
      let fixed = data.content.trim()
      fixed = fixed.replace(/^\`mermaid\n?/, '').replace(/\`\s*$/, '').trim()
      code.value = fixed
      hasError.value = false
      lastError.value = ''
      generateSVG()
    }
  } catch {
    // ignore fix failure
  } finally {
    fixingCode.value = false
  }
}

async function generateSVG() {
  loading.value = true
  hasError.value = false
  try {
    const resp = await fetch('http://localhost:9002/api/mermaid-to-svg', {
      method: 'POST',
      headers: { 'Content-Type': 'text/plain' },
      body: code.value
    })
    if (!resp.ok) {
      const text = await resp.text()
      try {
        const data = JSON.parse(text)
        throw new Error(data.error || text)
      } catch (e: any) {
        if (e instanceof SyntaxError) throw new Error(`API ${resp.status}: ${text}`)
        throw e
      }
    }
    svg.value = await resp.text()
    await nextTick()
    fitToView()
  } catch (e: any) {
    hasError.value = true
    const rawError = cleanError(e.message)
    lastError.value = rawError
    svg.value = `<pre class="text-red-500 text-sm p-2 whitespace-pre-wrap">${rawError}</pre>`
    translateError(rawError)
  } finally {
    loading.value = false
  }
}

function downloadSVG() {
  if (!svg.value) return
  const blob = new Blob([svg.value], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'diagram.svg'
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="w-full h-full p-2 bg-muted">
    <USplitter id="splitter-example" :items="innerItems">
      <template #main>
        <div class="w-full h-full flex flex-col">
          <!-- Left toolbar -->
          <div class="flex items-center gap-1 p-1 border-b border-default">
            <USelect
              v-model="diagramType"
              :items="diagramTypes"
              size="xs"
              class="w-32"
            />
            <UButton
              v-if="hasError"
              size="xs"
              color="warning"
              icon="i-lucide-wand-2"
              label="讓 AI 修正"
              :loading="fixingCode"
              @click="letAIFix"
            />
            <UButton
              size="xs"
              icon="i-lucide-file-code"
              label="生成"
              :loading="loading"
              class="ml-auto"
              @click="generateSVG"
            />
          </div>
          <!-- Editor -->
          <div class="flex-1 min-h-0">
            <Codemirror
              v-model="code"
              :extensions="extensions"
              :style="{ width: '100%', height: '100%' }"
            />
          </div>
        </div>
      </template>
      <template #right>
        <div class="w-full h-full flex flex-col">
          <!-- Zoom toolbar -->
          <div class="relative flex items-center gap-1 p-1 border-b border-default">
            <UButton
              size="xs"
              color="neutral"
              icon="i-lucide-zoom-in"
              :disabled="hasError"
              @click="zoomIn"
            />
            <UButton
              size="xs"
              color="neutral"
              icon="i-lucide-zoom-out"
              :disabled="hasError"
              @click="zoomOut"
            />
            <UButton
              size="xs"
              color="neutral"
              icon="i-lucide-focus"
              :disabled="hasError"
              @click="centerView"
            />
            <span class="absolute left-1/2 -translate-x-1/2 text-xs text-muted">{{ Math.round(zoom * 100) }}%</span>
            <UButton
              size="xs"
              color="neutral"
              icon="i-lucide-download"
              label="下載"
              :disabled="!svg || hasError"
              class="ml-auto"
              @click="downloadSVG"
            />
          </div>
          <!-- Preview canvas -->
          <div
            ref="canvasRef"
            class="flex-1 overflow-hidden bg-white flex items-center justify-center"
            :style="{ cursor: hasError ? 'default' : (dragging ? 'grabbing' : 'grab') }"
            @wheel="onWheel"
            @pointerdown="onPointerDown"
            @pointermove="onPointerMove"
            @pointerup="onPointerUp"
          >
            <div
              ref="svgWrapperRef"
              v-if="svg"
              v-html="svg"
              class="max-w-none"
              :style="{ transform: `scale(${zoom}) translate(${tx}px, ${ty}px)` }"
            ></div>
            <span v-else-if="translatingError" class="text-dimmed text-sm">翻譯錯誤中...</span>
            <span v-else class="text-muted text-sm">點擊「生成」來渲染圖形</span>
          </div>
        </div>
      </template>
    </USplitter>
  </div>
</template>


