import { z } from 'zod'

export const usernameSchema = z.string().regex(/^[a-zA-Z0-9_-]{3,36}$/)
export const txbInputSchema = z.string().regex(/^(?:0|[1-9]\d*)(?:\.\d{1,2})?$/)
export const adminReasonSchema = z.string().trim().min(4).max(300)
export const identifierSchema = z.string().regex(/^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/)

export function isValid<T>(schema: z.ZodType<T>, value: unknown): value is T {
  return schema.safeParse(value).success
}
