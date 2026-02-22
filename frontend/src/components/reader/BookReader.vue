<template>
  <div
    ref="readerRef"
    class="reader"
    :class="themeClass"
    :style="customColorVars"
  >
    <ReaderHeader />

    <ReaderContent
      ref="contentRef"
      @next-page="handleNextPage"
      @prev-page="handlePrevPage"
      @toggle-u-i="store.toggleUI()"
    />

    <ReaderFooter @navigate-to-progress="handleNavigateToProgress" />
    <ReaderTOC @navigate="handleNavigate" />
    <ReaderSettings />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useReaderStore } from '@/stores/reader'
import { useBookContent } from '@/composables/useBookContent'
import { useReaderKeyboard } from '@/composables/useReaderKeyboard'
import { useReadingProgress } from '@/composables/useReadingProgress'
import { useReaderSettings } from '@/composables/useReaderSettings'
import ReaderContent from './ReaderContent.vue'
import ReaderHeader from './ReaderHeader.vue'
import ReaderFooter from './ReaderFooter.vue'
import ReaderTOC from './ReaderTOC.vue'
import ReaderSettings from './ReaderSettings.vue'

const props = defineProps<{ bookId: number }>()

const store = useReaderStore()
const router = useRouter()
const { navigateToChapter, nextChapter, prevChapter, prefetchAdjacentChapters } = useBookContent()
const { loadProgress, scheduleSave } = useReadingProgress(props.bookId)
const { loadSettings, watchSettings } = useReaderSettings()
const contentRef = ref<InstanceType<typeof ReaderContent> | null>(null)
const readerRef = ref<HTMLElement | null>(null)

const themeClass = computed(() => `theme-${store.settings.theme}`)

const customColorVars = computed(() => {
  if (store.settings.theme !== 'custom' || !store.settings.customColors) return {}
  const c = store.settings.customColors
  return {
    '--custom-bg': c.background,
    '--custom-text': c.text,
    '--custom-link': c.link,
    '--custom-selection': c.selection,
  }
})

function handleNextPage() {
  if (!contentRef.value) return
  if (store.currentPage < store.totalPages) {
    contentRef.value.nextPage()
  } else {
    nextChapter(props.bookId)
  }
  scheduleSave()
}

function handlePrevPage() {
  if (!contentRef.value) return
  if (store.currentPage > 1) {
    contentRef.value.prevPage()
  } else {
    prevChapter(props.bookId)
  }
  scheduleSave()
}

function handleNavigate(chapterId: string) {
  navigateToChapter(props.bookId, chapterId)
  scheduleSave()
}

async function handleNavigateToProgress(percent: number) {
  if (!store.bookContent) return
  const chapters = store.bookContent.chapters
  const totalPages = store.bookTotalPages

  // Convert percent to target book page
  const targetPage = Math.max(1, Math.round((percent / 100) * (totalPages - 1)) + 1)

  // Find which chapter contains this page
  let accumulated = 0
  for (const chId of chapters) {
    const chPages = store.chapterPageCounts.get(chId) ?? 1
    if (accumulated + chPages >= targetPage) {
      const pageInChapter = targetPage - accumulated
      await navigateToChapter(props.bookId, chId)
      // Wait for DOM update before navigating to a specific page
      await nextTick()
      if (contentRef.value) {
        contentRef.value.goToPage(pageInChapter)
      }
      scheduleSave()
      return
    }
    accumulated += chPages
  }

  // Fallback: navigate to last chapter
  const lastChapter = chapters[chapters.length - 1]
  await navigateToChapter(props.bookId, lastChapter)
  scheduleSave()
}

useReaderKeyboard({
  nextPage: handleNextPage,
  prevPage: handlePrevPage,
  nextChapter: () => nextChapter(props.bookId),
  prevChapter: () => {
    if (store.currentPage > 1 && contentRef.value) {
      contentRef.value.goToPage(1)
    } else if (store.bookContent && store.hasPrevChapter) {
      // Navigate to START of previous chapter (not end)
      store.navigationDirection = 'forward'
      const prevId = store.bookContent.chapters[store.currentChapterIndex - 1]
      navigateToChapter(props.bookId, prevId)
    }
  },
  goToStart: () => {
    if (store.bookContent?.chapters.length) {
      navigateToChapter(props.bookId, store.bookContent.chapters[0])
    }
  },
  goToEnd: () => {
    if (store.bookContent?.chapters.length) {
      const chapters = store.bookContent.chapters
      navigateToChapter(props.bookId, chapters[chapters.length - 1])
    }
  },
  changeFontSize: (delta: number) => {
    const newSize = Math.max(12, Math.min(36, store.settings.fontSize + delta))
    store.updateSettings({ fontSize: newSize })
  },
  exitReader: () => {
    router.back()
  },
})

// Load settings, apply CSS variables, and restore reading progress
onMounted(async () => {
  await loadSettings()
  if (readerRef.value) {
    watchSettings(readerRef.value)
  }

  // Restore saved reading progress after settings are loaded
  const saved = await loadProgress()
  if (saved && saved.chapterId) {
    await navigateToChapter(props.bookId, saved.chapterId)
  }
})

// Prefetch adjacent chapters when current chapter changes
watch(
  () => store.currentChapterId,
  () => {
    prefetchAdjacentChapters(props.bookId)
  },
)
</script>
