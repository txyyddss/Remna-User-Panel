import en from '../../locales/en.json'
import zhCN from '../../locales/zh-CN.json'

export const localeMessages = {
  en,
  'zh-CN': zhCN,
} as const

export type Locale = keyof typeof localeMessages
export type LocaleMessages = typeof en

type LeafKeys<Value, Prefix extends string = ''> = Value extends object
  ? { [Key in keyof Value & string]: LeafKeys<Value[Key], Prefix extends '' ? Key : `${Prefix}.${Key}`> }[keyof Value & string]
  : Prefix

export type LocaleKey = LeafKeys<LocaleMessages>
