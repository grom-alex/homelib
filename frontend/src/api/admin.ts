import api from './client'

export interface ImportStats {
  books_added: number
  books_updated: number
  books_deleted: number
  authors_added: number
  genres_added: number
  series_added: number
  errors: number
  duration_ms: number
}

export interface ImportStatus {
  status: 'idle' | 'running' | 'completed' | 'failed'
  started_at?: string
  finished_at?: string
  stats?: ImportStats
  error?: string
  total_records?: number
  processed_batch?: number
  total_batches?: number
}

export async function startImport(): Promise<ImportStatus> {
  const { data } = await api.post<ImportStatus>('/admin/import')
  return data
}

export async function getImportStatus(): Promise<ImportStatus> {
  const { data } = await api.get<ImportStatus>('/admin/import/status')
  return data
}
