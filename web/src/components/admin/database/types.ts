import type { DatabaseFilter, DatabaseFilterOperator } from '@/api/features'

export type TextDatabaseFilter = Omit<DatabaseFilter, 'value'> & { value?: string }

export interface DatabaseColumnOption {
  value: string
  label: string
}

export interface DatabaseOperatorOption {
  value: DatabaseFilterOperator
  label: string
}
