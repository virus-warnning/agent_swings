<script setup lang="ts">
interface Subject {
  title: string
  icon: string
  color: string
  method: string
  tips: string[]
  tools: { name: string; to: string; icon: string; desc: string }[]
}

const subjects: Subject[] = [
  {
    title: '語文',
    icon: 'i-lucide-book-open',
    color: 'primary',
    method: '以「積累＋理解」為主軸：先廣泛閱讀建立詞彙庫，再以語感與文意分析強化表達能力。每日少量生字、搭配例句記憶，效果遠勝一次性背誦。',
    tips: [
      '每天記 5-10 個生字，搭配音標與例句，使用間隔重複法複習',
      '閱讀文章後用翻譯機把難句轉為繁體中文，理解語意再回看原文',
      '寫作時先列出關鍵詞，再組句擴寫，訓練結構化表達'
    ],
    tools: [
      { name: '生字卡', to: '/vocab-card', icon: 'i-lucide-book', desc: '批量查詢生字音標、詞性、例句，可匯出 Anki 間隔重複記憶' },
      { name: '翻譯機', to: '/translator', icon: 'i-lucide-languages', desc: '將外語原文翻為繁體中文，快速理解語意、輔助閱讀' }
    ]
  },
  {
    title: '數學',
    icon: 'i-lucide-calculator',
    color: 'info',
    method: '以「引導式解題」取代背公式：透過提問與線索讓大脑自主推導，遇到卡點時逐步提示而非直接給答案，培養獨立思考與問題拆解能力。',
    tips: [
      '解題時先寫出「已知」與「求什麼」，明確問題方向',
      '卡住時請 AI 家教給提示或類似範例，而非直接要答案',
      '解完後回顧「關鍵轉折步驟」，記下思路而非只記答案'
    ],
    tools: [
      { name: '家教', to: '/jhs-tutor', icon: 'i-lucide-graduation-cap', desc: 'AI 導師以提問與線索引導你逐步解題，不直接給答案' }
    ]
  },
  {
    title: '自然',
    icon: 'i-lucide-flask-conical',
    color: 'success',
    method: '以「概念圖＋流程化」為主軸：將抽象的自然現象拆解為因果鏈與流程圖，視覺化呈現概念間的關聯，幫助建立完整心智模型。',
    tips: [
      '學新章節時先畫出概念間的箭頭圖，標出「因為…所以…」',
      '用流程圖呈現實驗步驟、生理循環、地質變化等流程',
      '把多個章節的圖合併成一張大的心智地圖，串聯知識'
    ],
    tools: [
      { name: '美人魚圖', to: '/mermaid', icon: 'i-lucide-git-branch', desc: '用 Mermaid 語法繪製概念圖、流程圖、因果鏈，直覺呈現邏輯' }
    ]
  },
  {
    title: '社會',
    icon: 'i-lucide-globe',
    color: 'warning',
    method: '以「時間軸＋比較分析」為主軸：歷史事件用時間軸排列因果，地理與公民用比較表格釐清異同。閱讀史料時先翻譯再分析觀點。',
    tips: [
      '整理歷史事件時依時間軸排列，標出「起因→經過→影響」',
      '比較不同文明、政體、經濟制度時用表格對照異同',
      '閱讀外語史料或新聞時用翻譯機轉為中文，再分析作者立場'
    ],
    tools: [
      { name: '美人魚圖', to: '/mermaid', icon: 'i-lucide-git-branch', desc: '繪製時間軸、因果圖、制度比較圖，視覺化整理脈絡' },
      { name: '翻譯機', to: '/translator', icon: 'i-lucide-languages', desc: '將外語史料、國際新聞翻為中文，輔助理解與分析' }
    ]
  }
]
</script>

<template>
  <div class="p-4 space-y-6">
    <!-- Hero -->
    <div class="text-center space-y-2">
      <h1 class="text-2xl font-bold tracking-tight">如何高效學習</h1>
      <p class="text-sm text-dimmed">四大科目學習方法 × 工具推薦</p>
    </div>

    <!-- Subject cards -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <UCard
        v-for="s in subjects"
        :key="s.title"
        :ui="{ body: 'p-4 space-y-3' }"
        class="transition-shadow hover:shadow-md"
      >
        <!-- Card header -->
        <div class="flex items-center gap-2">
          <UIcon :name="s.icon" class="w-5 h-5" :color="s.color" />
          <h2 class="text-lg font-semibold">{{ s.title }}</h2>
        </div>

        <!-- Method -->
        <div class="text-sm leading-relaxed text-muted">
          <UIcon name="i-lucide-lightbulb" class="w-4 h-4 inline-block mr-1 -mt-0.5 text-warning" />
          {{ s.method }}
        </div>

        <!-- Tips -->
        <div class="space-y-1.5">
          <p class="text-xs font-semibold text-dimmed uppercase tracking-wide">學習技巧</p>
          <ul class="space-y-1">
            <li
              v-for="(tip, i) in s.tips"
              :key="i"
              class="text-xs text-muted flex items-start gap-1.5"
            >
              <UIcon name="i-lucide-check" class="w-3.5 h-3.5 mt-0.5 shrink-0 text-success" />
              <span>{{ tip }}</span>
            </li>
          </ul>
        </div>

        <!-- Tools -->
        <div class="space-y-1.5">
          <p class="text-xs font-semibold text-dimmed uppercase tracking-wide">推薦工具</p>
          <div class="flex flex-wrap gap-2">
            <ULink
              v-for="t in s.tools"
              :key="t.name"
              :to="t.to"
              class="group flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-lg border border-default bg-default hover:bg-elevated transition-colors"
            >
              <UIcon :name="t.icon" class="w-3.5 h-3.5 text-primary" />
              <span class="font-medium">{{ t.name }}</span>
              <UIcon name="i-lucide-arrow-right" class="w-3 h-3 text-dimmed group-hover:text-primary group-hover:translate-x-0.5 transition-all" />
            </ULink>
          </div>
          <!-- Tool descriptions -->
          <div class="space-y-1 mt-1">
            <div
              v-for="t in s.tools"
              :key="'desc-' + t.name"
              class="text-[11px] text-dimmed flex items-start gap-1"
            >
              <UIcon :name="t.icon" class="w-3 h-3 mt-px shrink-0" />
              <span>{{ t.desc }}</span>
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- General study tips -->
    <UCard :ui="{ body: 'p-4' }">
      <div class="flex items-center gap-2 mb-2">
        <UIcon name="i-lucide-sparkles" class="w-5 h-5 text-primary" />
        <h2 class="text-lg font-semibold">通用學習原則</h2>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div class="flex items-start gap-2 p-2 rounded-lg bg-default">
          <UIcon name="i-lucide-timer" class="w-4 h-4 text-primary shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-medium">間隔重複</p>
            <p class="text-[11px] text-dimmed">1-3-7-14 天節奏複習，對抗遺忘曲線</p>
          </div>
        </div>
        <div class="flex items-start gap-2 p-2 rounded-lg bg-default">
          <UIcon name="i-lucide-puzzle" class="w-4 h-4 text-info shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-medium">主動回想</p>
            <p class="text-[11px] text-dimmed">合上書本嘗試重述，比反覆閱讀有效 3 倍</p>
          </div>
        </div>
        <div class="flex items-start gap-2 p-2 rounded-lg bg-default">
          <UIcon name="i-lucide-git-compare" class="w-4 h-4 text-success shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-medium">交錯練習</p>
            <p class="text-[11px] text-dimmed">混合不同題型練習，提升辨識與遷移能力</p>
          </div>
        </div>
        <div class="flex items-start gap-2 p-2 rounded-lg bg-default">
          <UIcon name="i-lucide-brain" class="w-4 h-4 text-warning shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-medium">費曼技巧</p>
            <p class="text-[11px] text-dimmed">用簡單語言教別人，暴露理解盲區</p>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>
