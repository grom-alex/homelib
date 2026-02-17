import api from './client'

export interface BookAuthorRef {
  id: number
  name: string
}

export interface BookGenreRef {
  id: number
  name: string
}

export interface BookSeriesRef {
  id: number
  name: string
  num?: number
}

export interface BookListItem {
  id: number
  title: string
  lang: string
  year?: number
  format: string
  file_size?: number
  lib_rate?: number
  is_deleted: boolean
  authors: BookAuthorRef[]
  genres: BookGenreRef[]
  series?: BookSeriesRef
}

export interface BookDetail {
  id: number
  title: string
  lang: string
  year?: number
  format: string
  file_size?: number
  lib_rate?: number
  is_deleted: boolean
  description?: string
  keywords?: string[]
  date_added?: string
  authors: BookAuthorRef[]
  genres: { id: number; code: string; name: string }[]
  series?: { id: number; name: string; num?: number; type?: string }
  collection?: { id: number; name: string }
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface BookFilters {
  q?: string
  author_id?: number
  genre_id?: number
  series_id?: number
  lang?: string
  format?: string
  page?: number
  limit?: number
  sort?: string
  order?: string
}

export interface AuthorListItem {
  id: number
  name: string
  books_count: number
}

export interface AuthorDetail {
  id: number
  name: string
  books: BookListItem[]
  books_count: number
}

export interface GenreTreeItem {
  id: number
  code: string
  name: string
  meta_group?: string
  books_count: number
  children?: GenreTreeItem[]
}

export interface SeriesListItem {
  id: number
  name: string
  books_count: number
}

export interface CatalogStats {
  books_count: number
  authors_count: number
  genres_count: number
  series_count: number
  languages: string[]
  formats: string[]
}

export async function getBooks(filters: BookFilters = {}, signal?: AbortSignal): Promise<PaginatedResponse<BookListItem>> {
  const params = Object.fromEntries(
    Object.entries(filters).filter(([, v]) => v !== undefined && v !== '' && v !== null),
  )
  const { data } = await api.get<PaginatedResponse<BookListItem>>('/books', { params, signal })
  return data
}

export async function getBook(id: number): Promise<BookDetail> {
  const { data } = await api.get<BookDetail>(`/books/${id}`)
  return data
}

export async function downloadBook(id: number): Promise<void> {
  const response = await api.get(`/books/${id}/download`, { responseType: 'blob' })
  const disposition = response.headers['content-disposition'] || ''

  let filename = `book_${id}`
  // Try RFC 6266 filename*=UTF-8''... first, then plain filename="..."
  const utf8Match = disposition.match(/filename\*=UTF-8''(.+?)(?:;|$)/i)
  if (utf8Match) {
    filename = decodeURIComponent(utf8Match[1])
  } else {
    const plainMatch = disposition.match(/filename="(.+?)"/)
    if (plainMatch) {
      filename = plainMatch[1]
    }
  }

  const url = window.URL.createObjectURL(response.data)
  try {
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  } finally {
    window.URL.revokeObjectURL(url)
  }
}

export async function getAuthors(params: { q?: string; page?: number; limit?: number } = {}): Promise<PaginatedResponse<AuthorListItem>> {
  const { data } = await api.get<PaginatedResponse<AuthorListItem>>('/authors', { params })
  return data
}

export async function getAuthor(id: number): Promise<AuthorDetail> {
  const { data } = await api.get<AuthorDetail>(`/authors/${id}`)
  return data
}

export async function getGenres(): Promise<GenreTreeItem[]> {
  const { data } = await api.get<GenreTreeItem[]>('/genres')
  return data
}

export async function getSeries(params: { q?: string; page?: number; limit?: number } = {}): Promise<PaginatedResponse<SeriesListItem>> {
  const { data } = await api.get<PaginatedResponse<SeriesListItem>>('/series', { params })
  return data
}

export async function getStats(): Promise<CatalogStats> {
  const { data } = await api.get<CatalogStats>('/stats')
  return data
}
