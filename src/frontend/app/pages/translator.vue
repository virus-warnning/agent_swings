<script setup lang="ts">
import type { SplitterItem } from '@nuxt/ui'
import { ref } from 'vue'
import { aiModels } from '~/composables/useAiModels'

const card = 'border border-default rounded-xl overflow-hidden p-[5px] bg-default'

const innerItems: SplitterItem[] = [
  { slot: 'main', minSize: 30, defaultSize: 45, class: card },
  { slot: 'right', minSize: 30, defaultSize: 55, class: card }
]

const model = ref('chat')

const sourceText = ref(`SyntaxError: Lexical error on line 1. Unrecognized text.
flowchart TD2  A1(["👩🏻‍💻`)
const result = ref('')
const loading = ref(false)
const hasError = ref(false)
const error = ref('')
const elapsed = ref(0)

const SYSTEM_PROMPT = '你是一位專業翻譯人員。將使用者的文字翻譯為繁體中文。只負責翻譯，不添加解釋、備註或額外說明。直接輸出翻譯結果，盡可能用最短時間完成。'

async function translate() {
  if (!sourceText.value.trim()) return
  loading.value = true
  hasError.value = false
  error.value = ''
  result.value = ''

  const start = performance.now()
  try {
    const resp = await fetch('http://localhost:9002/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt: SYSTEM_PROMPT,
        content: sourceText.value,
        model: model.value
      })
    })
    elapsed.value = Math.round(performance.now() - start)

    const data = await resp.json()
    if (!resp.ok || data.error) {
      hasError.value = true
      error.value = data.detail || data.error || `HTTP ${resp.status}`
    } else {
      result.value = data.content
    }
  } catch (e: any) {
    elapsed.value = Math.round(performance.now() - start)
    hasError.value = true
    error.value = e.message || String(e)
  } finally {
    loading.value = false
  }
}

function clearAll() {
  sourceText.value = ''
  result.value = ''
  error.value = ''
  hasError.value = false
  elapsed.value = 0
}
</script>

<template>
  <div class="w-full h-full p-2 bg-muted">
    <USplitter id="translator-splitter" :items="innerItems">
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
            <span class="text-xs text-dimmed ml-2">原文</span>
            <UButton
              size="xs"
              icon="i-lucide-languages"
              label="翻譯"
              :loading="loading"
              :disabled="!sourceText.trim()"
              class="ml-auto"
              @click="translate"
            />
          </div>
          <!-- Source editor -->
          <div class="flex-1 min-h-0 p-3">
            <textarea
              v-model="sourceText"
              class="w-full h-full resize-none border-none outline-none bg-transparent text-sm font-mono"
              placeholder="Paste text here to translate..."
              @keydown.ctrl.enter="translate"
            />
          </div>
          <div class="flex items-center justify-between px-3 py-1 border-t border-default">
            <span class="text-[10px] text-dimmed">Ctrl+Enter 翻譯</span>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-lucide-eraser"
              label="清除"
              @click="clearAll"
            />
          </div>
        </div>
      </template>
      <template #right>
        <div class="w-full h-full flex flex-col">
          <!-- Right toolbar -->
          <div class="flex items-center gap-1 p-1 border-b border-default">
            <span class="text-xs text-dimmed">翻譯結果</span>
            <span v-if="elapsed > 0" class="text-xs text-dimmed ml-2">{{ elapsed }}ms</span>
          </div>
          <!-- Result area -->
          <div class="flex-1 min-h-0 overflow-y-auto p-3">
            <div v-if="loading" class="flex items-center justify-center h-full text-dimmed">
              <UIcon name="i-lucide-loader-circle" class="animate-spin mr-2" />
              翻譯中...
            </div>
            <div v-else-if="hasError" class="text-red-500 text-sm whitespace-pre-wrap font-mono">{{ error }}</div>
            <div v-else-if="result" class="text-sm whitespace-pre-wrap leading-relaxed">{{ result }}</div>
            <div v-else class="flex flex-col items-center justify-center h-full text-dimmed text-sm gap-2">
              <UIcon name="i-lucide-languages" class="w-8 h-8" />
              <span>在左邊貼上文字，點擊翻譯</span>
            </div>
          </div>
        </div>
      </template>
    </USplitter>
  </div>
</template>
