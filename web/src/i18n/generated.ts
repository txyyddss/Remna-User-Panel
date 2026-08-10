import enAdminActivity from '../../locales/en/admin-activity.json'
import enAdminCatalog from '../../locales/en/admin-catalog.json'
import enAdminCommunity from '../../locales/en/admin-community.json'
import enAdminCore from '../../locales/en/admin-core.json'
import enAdminData from '../../locales/en/admin-data.json'
import enAdminOperations from '../../locales/en/admin-operations.json'
import enActivity from '../../locales/en/activity.json'
import enCommerce from '../../locales/en/commerce.json'
import enCommunity from '../../locales/en/community.json'
import enCore from '../../locales/en/core.json'
import enMember from '../../locales/en/member.json'
import zhAdminActivity from '../../locales/zh-CN/admin-activity.json'
import zhAdminCatalog from '../../locales/zh-CN/admin-catalog.json'
import zhAdminCommunity from '../../locales/zh-CN/admin-community.json'
import zhAdminCore from '../../locales/zh-CN/admin-core.json'
import zhAdminData from '../../locales/zh-CN/admin-data.json'
import zhAdminOperations from '../../locales/zh-CN/admin-operations.json'
import zhActivity from '../../locales/zh-CN/activity.json'
import zhCommerce from '../../locales/zh-CN/commerce.json'
import zhCommunity from '../../locales/zh-CN/community.json'
import zhCore from '../../locales/zh-CN/core.json'
import zhMember from '../../locales/zh-CN/member.json'

const en = {
  ...enCore, ...enActivity, ...enCommerce, ...enCommunity, ...enMember, ...enAdminCore, ...enAdminData,
  ...enAdminActivity, ...enAdminCommunity, ...enAdminCatalog, ...enAdminOperations,
}
const zhCN: typeof en = {
  ...zhCore, ...zhActivity, ...zhCommerce, ...zhCommunity, ...zhMember, ...zhAdminCore, ...zhAdminData,
  ...zhAdminActivity, ...zhAdminCommunity, ...zhAdminCatalog, ...zhAdminOperations,
}

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
