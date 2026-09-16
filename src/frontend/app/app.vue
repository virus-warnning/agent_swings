<script setup>
const colorMode = useColorMode()

if (process.client && !localStorage.getItem('nuxt-color-mode')) {
  colorMode.value = 'light'
}
function toggleColorMode() {
  colorMode.value = colorMode.value === 'dark' ? 'light' : 'dark'
}
useHead({
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ],
  link: [
    { rel: 'icon', href: '/favicon.ico' }
  ],
  htmlAttrs: {
    lang: 'en'
  }
})

const title = 'Nuxt Starter Template'
const description = 'A production-ready starter template powered by Nuxt UI. Build beautiful, accessible, and performant applications in minutes, not hours.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  ogImage: 'https://ui.nuxt.com/assets/templates/nuxt/starter-light.png',
  twitterCard: 'summary_large_image'
})

const sidebarOpen = ref(true)

const navLinks = [
  { label: 'Mermaid', icon: 'i-lucide-git-branch', to: '/mermaid' },
  { label: 'Projects', icon: 'i-lucide-folder-open', to: '/projects' },
  { label: 'Settings', icon: 'i-lucide-settings', to: '/settings' }
]
</script>

<template>
  <UApp>
    <div class="flex h-dvh flex-col">
      <UHeader
        title="Nuxt UI"
        class="h-12"
        :ui="{ base: 'h-8 px-2 justify-between', container: 'max-w-full px-0' }"
      >
        <template #left>
          <div>
            Agent Swings
          </div>
        </template>
        <template #right>
          <div class="p-1">
            <UButton
              :icon="colorMode === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'"
              :ui="{ base: 'transition-colors' }"
              color="neutral"
              variant="ghost"
              :aria-label="`Switch to ${colorMode === 'dark' ? 'light' : 'dark'} mode`"
              @click="toggleColorMode"
            />
          </div>
        </template>
      </UHeader>
      <div class="flex flex-1 overflow-hidden">
        <USidebar v-model:open="sidebarOpen"
          variant="inset"
          collapsible="none"
          class="border-r border-default">
          <ULink to="/">如何高效學習</ULink>
          <ULink to="/mermaid">美人魚圖</ULink>
          <ULink to="/jhs-tutor">家教</ULink>
          <ULink to="/translator">翻譯機</ULink>
          <ULink to="/vocab-card">生字卡</ULink>
        </USidebar>
        <UMain class="flex-1 overflow-y-auto">
          <NuxtPage />
        </UMain>
      </div>
    </div>
  </UApp>
</template>
