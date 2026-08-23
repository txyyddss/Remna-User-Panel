import type { DatabaseFilter, DatabaseFilterOperator } from '@/api/features'

export type DeepReadonly<T> = T extends readonly (infer Item)[]
  ? readonly DeepReadonly<Item>[]
  : T extends object ? { readonly [Key in keyof T]: DeepReadonly<T[Key]> } : T

export type TextDatabaseFilter = Omit<DatabaseFilter, 'value'> & { value?: string }

export interface DatabaseColumnOption {
  value: string
  label: string
}

export interface DatabaseOperatorOption {
  value: DatabaseFilterOperator
  label: string
}
