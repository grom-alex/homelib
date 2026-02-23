// Reader types per §8.5 architecture

export interface BookMetadata {
  title: string
  author: string
  cover: string | null
  language: string
  format: string
}

export interface TOCEntry {
  id: string
  title: string
  level: number
}

export interface BookContent {
  metadata: BookMetadata
  toc: TOCEntry[]
  chapters: string[]
  totalChapters: number
  chapterSizes?: Record<string, number>
}

export interface ChapterContent {
  id: string
  title: string
  html: string
}

export interface ReadingPosition {
  chapterId: string
  chapterProgress: number // 0-100
  totalProgress: number // 0-100
  device: string
  updatedAt?: string
}

export interface CustomColors {
  background: string
  text: string
  link: string
  selection: string
}

export interface ReaderSettings {
  // Font
  fontSize: number // 12-36 px
  fontFamily: string // 'Georgia' | 'PT Serif' | 'Literata' | 'OpenDyslexic' | 'System'
  fontWeight: 400 | 500

  // Spacing
  lineHeight: number // 1.0-2.5
  paragraphSpacing: number // 0-2 em
  letterSpacing: number // -0.05 — 0.1 em

  // Margins
  marginHorizontal: number // 0-20 % width
  marginVertical: number // 0-10 % height
  firstLineIndent: number // 0-3 em

  // Text
  textAlign: 'left' | 'justify'
  hyphenation: boolean

  // Theme
  theme: 'light' | 'sepia' | 'dark' | 'night' | 'custom'
  customColors?: CustomColors | null

  // View mode
  viewMode: 'paginated' | 'scroll'
  pageAnimation: 'slide' | 'fade' | 'none'

  // Extra
  showProgress: boolean
  showClock: boolean
  tapZones: 'lr' | 'lrc'
}

export const defaultSettings: ReaderSettings = {
  fontSize: 18,
  fontFamily: 'Georgia',
  fontWeight: 400,
  lineHeight: 1.6,
  paragraphSpacing: 0.5,
  letterSpacing: 0,
  marginHorizontal: 5,
  marginVertical: 3,
  firstLineIndent: 1.5,
  textAlign: 'justify',
  hyphenation: true,
  theme: 'light',
  customColors: null,
  viewMode: 'paginated',
  pageAnimation: 'slide',
  showProgress: true,
  showClock: false,
  tapZones: 'lrc',
}

export const fontFamilies = [
  'Georgia',
  'PT Serif',
  'Literata',
  'OpenDyslexic',
  'System',
] as const

export type FontFamily = (typeof fontFamilies)[number]
