<script setup lang="ts">
import type { SplitterItem } from '@nuxt/ui'
import { ref } from 'vue'
import { aiModels } from '~/composables/useAiModels'

const card = 'border border-default rounded-xl overflow-hidden p-[5px] bg-default'

const innerItems: SplitterItem[] = [
  { slot: 'main', minSize: 30, defaultSize: 35, class: card },
  { slot: 'right', minSize: 30, defaultSize: 65, class: card }
]

const model = ref('chat')

const wordList = ref(`breathtaking
meticulous
ephemeral
resilient
paradox`)

interface VocabEntry {
  word: string
  phonetic: string
  pos: string
  meaning: string
  example: string
}

const results = ref<VocabEntry[]>([])
const loading = ref(false)
const hasError = ref(false)
const error = ref('')
const elapsed = ref(0)

const SYSTEM_PROMPT = `你是一位專業英語辭典查詢助手。使用者會給你一組英文生字（每行一個），你需要為每個生字提供：
- 國際音標
- 詞性（n. / v. / adj. / adv. 等）
- 簡短中文翻譯（一個最常見的意思）
- 一個簡短英文例句

請以 JSON 陣列回傳，格式如下（只回傳 JSON，不要其他文字）：
[{"word":"breathtaking","phonetic":"/ˈbreɪtkeɪkɪŋ/","pos":"adj.","meaning":"令人驚嘆的","example":"The breathtaking view from the top was unforgettable."}]`

async function lookup() {
  const words = wordList.value.trim().split('\n').map(w => w.trim()).filter(Boolean)
  if (words.length === 0) return

  loading.value = true
  hasError.value = false
  error.value = ''
  results.value = []

  const start = performance.now()
  try {
    const resp = await fetch('http://localhost:9002/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt: SYSTEM_PROMPT,
        content: wordList.value,
        model: model.value
      })
    })
    elapsed.value = Math.round(performance.now() - start)

    const data = await resp.json()
    if (!resp.ok || data.error) {
      hasError.value = true
      error.value = data.detail || data.error || `HTTP ${resp.status}`
      return
    }

    // Parse JSON from AI response
    let jsonStr = data.content
    // Try to extract JSON array if AI wrapped it in markdown
    const match = jsonStr.match(/\[[\s\S]*\]/)
    if (match) jsonStr = match[0]

    const parsed = JSON.parse(jsonStr) as VocabEntry[]
    results.value = Array.isArray(parsed) ? parsed : []
  } catch (e: any) {
    elapsed.value = Math.round(performance.now() - start)
    hasError.value = true
    error.value = e.message || String(e)
  } finally {
    loading.value = false
  }
}

function downloadAnki() {
  if (results.value.length === 0) return
  // Anki CSV: Front,Back
  const lines = ['Front,Back']
  for (const r of results.value) {
    const front = escapeCsv(`${r.word}  ${r.phonetic}`)
    const back = escapeCsv(`${r.pos} ${r.meaning}\n${r.example}`)
    lines.push(`${front},${back}`)
  }
  const csv = lines.join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'vocab-anki.csv'
  a.click()
  URL.revokeObjectURL(url)
}

function escapeCsv(s: string): string {
  if (s.includes(',') || s.includes('"') || s.includes('\n')) {
    return '"' + s.replace(/"/g, '""') + '"'
  }
  return s
}
</script>

<template>
  <div class="w-full h-full p-2 bg-muted">
    <USplitter id="vocab-splitter" :items="innerItems">
      <template #main>
        <div class="w-full h-full flex flex-col">
          <!-- Left toolbar -->
          <div class="flex items-center gap-1 p-1 border-b border-default">
            <USelect
              v-model="model"
              :items="aiModels"
              size="xs"
              class="w-28"
            />
            <span class="text-xs text-dimmed ml-2">生字</span>
            <UButton
              size="xs"
              icon="i-lucide-search"
              label="查詢"
              :loading="loading"
              :disabled="!wordList.trim()"
              class="ml-auto"
              @click="lookup"
            />
          </div>
          <!-- Word list editor -->
          <div class="flex-1 min-h-0 p-3">
            <textarea
              v-model="wordList"
              class="w-full h-full resize-none border-none outline-none bg-transparent text-sm font-mono"
              placeholder="每行輸入一個英文生字..."
              @keydown.ctrl.enter="lookup"
            />
          </div>
          <div class="flex items-center justify-between px-3 py-1 border-t border-default">
            <span class="text-[10px] text-dimmed">每行一個生字，Ctrl+Enter 查詢</span>
          </div>
        </div>
      </template>
      <template #right>
        <div class="w-full h-full flex flex-col">
          <!-- Right toolbar -->
          <div class="flex items-center gap-1 p-1 border-b border-default">
            <span class="text-xs text-dimmed">查詢結果</span>
            <span v-if="elapsed > 0" class="text-xs text-dimmed ml-2">{{ elapsed }}ms</span>
            <UButton
              size="xs"
              icon="i-lucide-download"
              label="下載 Anki"
              :disabled="results.length === 0"
              class="ml-auto"
              @click="downloadAnki"
            />
          </div>
          <!-- Result table -->
          <div class="flex-1 min-h-0 overflow-y-auto p-3">
            <div v-if="loading" class="flex items-center justify-center h-full text-dimmed">
              <UIcon name="i-lucide-loader-circle" class="animate-spin mr-2" />
              查詢中...
            </div>
            <div v-else-if="hasError" class="text-red-500 text-sm whitespace-pre-wrap font-mono">{{ error }}</div>
            <div v-else-if="results.length > 0">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left border-b border-default">
                    <th class="py-1.5 pr-2 font-medium">Word</th>
                    <th class="py-1.5 pr-2 font-medium">Phonetic</th>
                    <th class="py-1.5 pr-2 font-medium">POS</th>
                    <th class="py-1.5 pr-2 font-medium">Meaning</th>
                    <th class="py-1.5 font-medium">Example</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(r, i) in results" :key="i" class="border-b border-default/50">
                    <td class="py-1.5 pr-2 font-semibold">{{ r.word }}</td>
                    <td class="py-1.5 pr-2 font-mono text-dimmed">{{ r.phonetic }}</td>
                    <td class="py-1.5 pr-2 text-dimmed">{{ r.pos }}</td>
                    <td class="py-1.5 pr-2">{{ r.meaning }}</td>
                    <td class="py-1.5 text-dimmed italic">{{ r.example }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="flex flex-col items-center justify-center h-full text-dimmed text-sm gap-2">
              <UIcon name="i-lucide-book-open" class="w-8 h-8" />
              <span>在左邊輸入生字，點擊查詢</span>
            </div>
          </div>
        </div>
      </template>
    </USplitter>
  </div>
</template>
