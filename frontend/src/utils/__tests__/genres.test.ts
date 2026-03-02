import { describe, it, expect } from 'vitest'
import { sortGenreTree } from '../genres'
import type { GenreTreeItem } from '@/api/books'

function makeGenre(id: number, name: string, children?: GenreTreeItem[]): GenreTreeItem {
  return { id, code: `code_${id}`, name, position: `0.${id}`, books_count: 0, children }
}

describe('sortGenreTree', () => {
  it('sorts genres alphabetically by Russian locale', () => {
    const input = [
      makeGenre(1, 'Фантастика'),
      makeGenre(2, 'Детективы'),
      makeGenre(3, 'Приключения'),
    ]
    const result = sortGenreTree(input)
    expect(result.map(g => g.name)).toEqual(['Детективы', 'Приключения', 'Фантастика'])
  })

  it('sorts children recursively', () => {
    const input = [
      makeGenre(1, 'Родитель', [
        makeGenre(3, 'Яблоко'),
        makeGenre(2, 'Абрикос'),
      ]),
    ]
    const result = sortGenreTree(input)
    expect(result[0].children!.map(g => g.name)).toEqual(['Абрикос', 'Яблоко'])
  })

  it('returns empty array for empty input', () => {
    expect(sortGenreTree([])).toEqual([])
  })

  it('does not mutate original array', () => {
    const input = [makeGenre(2, 'Б'), makeGenre(1, 'А')]
    const original = [...input]
    sortGenreTree(input)
    expect(input.map(g => g.name)).toEqual(original.map(g => g.name))
  })

  it('handles genres without children', () => {
    const input = [makeGenre(1, 'Тест')]
    const result = sortGenreTree(input)
    expect(result[0].children).toBeUndefined()
  })
})
