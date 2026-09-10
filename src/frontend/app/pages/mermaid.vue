<script setup lang="ts">
import type { SplitterItem } from '@nuxt/ui'
import { ref } from 'vue'
import { Codemirror } from 'vue-codemirror'
import { EditorView } from '@codemirror/view'

const card = 'border border-default rounded-xl overflow-hidden p-[5px] bg-default'

const items: SplitterItem[] = [
  { slot: 'main', minSize: 25, defaultSize: 50, class: card },
  { slot: 'right', minSize: 25, defaultSize: 50, class: card }
]

const code = ref(`graph TD
    A[Start] --> B{Is it?}
    B -->|Yes| C[OK]
    B -->|No| D[End]`)

const extensions = [
  EditorView.theme({
    '&': { height: '100%', fontSize: '14px' },
    '.cm-scroller': { overflow: 'auto' }
  })
]
</script>

<template>
  <div class="w-full h-full p-2 bg-muted">
    <USplitter id="splitter-example" :items="items">
      <template #main>
        <Codemirror
          v-model="code"
          :extensions="extensions"
          :style="{ width: '100%', height: '100%' }"
        />
      </template>
      <template #right>
        Right
      </template>
    </USplitter>
  </div>
</template>


