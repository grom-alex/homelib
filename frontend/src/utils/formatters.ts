interface NamedItem {
  id: number
  name: string
}

export function formatAuthorsSummary(authors: NamedItem[]): string {
  if (!authors || authors.length === 0) return '—'
  if (authors.length === 1) return authors[0].name
  return `${authors[0].name} и др.`
}

export function formatAuthorsFull(authors: NamedItem[]): string {
  if (!authors || authors.length === 0) return '—'
  return authors.map((a) => a.name).join(', ')
}

export function formatGenres(genres: NamedItem[]): string {
  if (!genres || genres.length === 0) return '—'
  return genres[0].name
}

export function formatGenresFull(genres: NamedItem[]): string {
  if (!genres || genres.length === 0) return '—'
  return genres.map((g) => g.name).join(', ')
}

export function formatSeries(book: { series?: { name: string; num?: number } }): string {
  if (!book.series) return '—'
  return book.series.num
    ? `${book.series.name} #${book.series.num}`
    : book.series.name
}

export function formatFileSize(bytes?: number): string {
  if (!bytes) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
