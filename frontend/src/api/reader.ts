import api from './client'
import type { BookContent, ChapterContent, ReadingPosition, ReaderSettings } from '@/types/reader'

export async function getBookContent(bookId: number): Promise<BookContent> {
  const { data } = await api.get<BookContent>(`/books/${bookId}/content`)
  return data
}

export async function getChapter(bookId: number, chapterId: string): Promise<ChapterContent> {
  const { data } = await api.get<ChapterContent>(`/books/${bookId}/chapter/${chapterId}`)
  return data
}

export function getBookImageUrl(bookId: number, imageId: string): string {
  return `/api/books/${bookId}/image/${imageId}`
}

export async function getReadingProgress(bookId: number): Promise<ReadingPosition | null> {
  const response = await api.get(`/me/books/${bookId}/progress`)
  if (response.status === 204) return null
  return response.data as ReadingPosition
}

export async function saveReadingProgress(
  bookId: number,
  position: Omit<ReadingPosition, 'updatedAt'>,
): Promise<ReadingPosition> {
  const { data } = await api.put<ReadingPosition>(`/me/books/${bookId}/progress`, position)
  return data
}

export async function getUserSettings(): Promise<{ reader?: Partial<ReaderSettings> }> {
  const { data } = await api.get<{ reader?: Partial<ReaderSettings> }>('/me/settings')
  return data
}

export async function updateUserSettings(
  settings: { reader: Partial<ReaderSettings> },
): Promise<{ reader: ReaderSettings }> {
  const { data } = await api.put<{ reader: ReaderSettings }>('/me/settings', settings)
  return data
}
