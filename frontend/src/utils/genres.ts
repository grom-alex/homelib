import type { GenreTreeItem } from '@/api/books'

export function sortGenreTree(items: GenreTreeItem[]): GenreTreeItem[] {
  const sorted = [...items].sort((a, b) => a.name.localeCompare(b.name, 'ru'))
  return sorted.map(item => ({
    ...item,
    children: item.children ? sortGenreTree(item.children) : undefined,
  }))
}
