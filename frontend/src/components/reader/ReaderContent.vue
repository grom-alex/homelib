<template>
  <div
    ref="pageContainerRef"
    class="reader-page-container"
    :style="containerStyles"
    @click="handleContentClick"
  >
    <div class="reader-columns-viewport">
      <div
        ref="columnsRef"
        class="reader-columns reader-content"
        :class="animationClass"
        :style="contentStyles"
        v-html="store.currentChapterContent?.html ?? ''"
      />
    </div>

    <!-- Footnote popup -->
    <div
      v-if="footnotePopup.visible"
      class="footnote-popup"
      :style="{ top: footnotePopup.top + 'px', left: footnotePopup.left + 'px' }"
      v-html="footnotePopup.html"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useReaderStore } from '@/stores/reader'
import { usePagination } from '@/composables/usePagination'
import { useReaderGestures } from '@/composables/useReaderGestures'

const emit = defineEmits<{
  nextPage: []
  prevPage: []
  toggleUI: []
}>()

const store = useReaderStore()
const pageContainerRef = ref<HTMLElement | null>(null)
const columnsRef = ref<HTMLElement | null>(null)

const { translateX, calculateTotalPages, goToPage, nextPage, prevPage, recalculate, setupResizeObserver, cleanup } = usePagination(columnsRef)

useReaderGestures(pageContainerRef, {
  nextPage: () => emit('nextPage'),
  prevPage: () => emit('prevPage'),
  toggleUI: () => emit('toggleUI'),
})

const footnotePopup = reactive({
  visible: false,
  html: '',
  top: 0,
  left: 0,
})

const animationClass = computed(() => {
  return `animation-${store.settings.pageAnimation}`
})

const containerStyles = computed(() => {
  const s = store.settings
  return {
    '--margin-horizontal': s.marginHorizontal + 'vw',
    '--margin-vertical': s.marginVertical + 'vh',
  }
})

const contentStyles = computed(() => {
  const s = store.settings
  return {
    transform: `translateX(${translateX.value}px)`,
    '--font-size': s.fontSize + 'px',
    '--font-family': s.fontFamily === 'System' ? 'system-ui, sans-serif' : `"${s.fontFamily}", serif`,
    '--font-weight': String(s.fontWeight),
    '--line-height': String(s.lineHeight),
    '--paragraph-spacing': s.paragraphSpacing + 'em',
    '--letter-spacing': s.letterSpacing + 'em',
    '--first-line-indent': s.firstLineIndent + 'em',
    '--text-align': s.textAlign,
    '--hyphenation': s.hyphenation ? 'auto' : 'manual',
  }
})

function handleContentClick(e: MouseEvent) {
  const target = e.target as HTMLElement

  // Footnote ref click
  if (target.classList.contains('footnote-ref')) {
    e.preventDefault()
    e.stopPropagation()
    showFootnote(target)
    return
  }

  // Close popup on outside click
  if (footnotePopup.visible) {
    footnotePopup.visible = false
    return
  }
}

function showFootnote(anchor: HTMLElement) {
  const noteId = anchor.getAttribute('data-note-id')
  if (!noteId) return

  // Find footnote body in the current chapter content
  const container = columnsRef.value
  if (!container) return

  const body = container.querySelector(`#${CSS.escape(noteId)}`)
  if (!body) return

  footnotePopup.html = body.innerHTML

  // Position near the anchor
  const rect = anchor.getBoundingClientRect()
  const containerRect = pageContainerRef.value?.getBoundingClientRect()
  if (!containerRect) return

  footnotePopup.top = rect.bottom - containerRect.top + 4
  footnotePopup.left = Math.min(
    rect.left - containerRect.left,
    containerRect.width - 316,
  )
  footnotePopup.visible = true
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape' && footnotePopup.visible) {
    footnotePopup.visible = false
    e.stopPropagation()
  }
}

onMounted(() => {
  nextTick(() => {
    calculateTotalPages()
    setupResizeObserver()
  })
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  cleanup()
  document.removeEventListener('keydown', handleEscape)
})

watch(
  () => store.currentChapterContent,
  () => {
    footnotePopup.visible = false
    nextTick(() => {
      // Disable animation during chapter transition to prevent reverse slide
      const el = columnsRef.value
      if (el) el.style.transition = 'none'

      calculateTotalPages()

      // Navigate backward → open last page of previous chapter
      if (store.navigationDirection === 'backward') {
        goToPage(store.totalPages)
      } else {
        goToPage(1)
      }
      store.navigationDirection = 'forward'

      // Re-enable animation after layout settles
      requestAnimationFrame(() => {
        if (el) el.style.transition = ''
      })

      // Recalculate when images finish loading
      watchImageLoads()
    })
  },
)

function watchImageLoads() {
  const el = columnsRef.value
  if (!el) return
  const imgs = el.querySelectorAll('img')
  if (imgs.length === 0) return

  let pending = 0
  for (const img of imgs) {
    if (!img.complete) {
      pending++
      img.addEventListener('load', onImageLoad, { once: true })
      img.addEventListener('error', onImageLoad, { once: true })
    }
  }

  function onImageLoad() {
    pending--
    if (pending <= 0) {
      recalculate()
    }
  }
}

defineExpose({ nextPage, prevPage, recalculate, goToPage })
</script>
