import enAdminActivity from '../../locales/en/admin-activity.json'
import enAdminCatalog from '../../locales/en/admin-catalog.json'
import enAdminCommunity from '../../locales/en/admin-community.json'
import enAdminCore from '../../locales/en/admin-core.json'
import enAdminData from '../../locales/en/admin-data.json'
import enAdminOnboarding from '../../locales/en/admin-onboarding.json'
import enAdminOperations from '../../locales/en/admin-operations.json'
import enAdminWorkflows from '../../locales/en/admin-workflows.json'
import enActivity from '../../locales/en/activity.json'
import enAffiliates from '../../locales/en/affiliates.json'
import enCommerce from '../../locales/en/commerce.json'
import enCommunity from '../../locales/en/community.json'
import enConnections from '../../locales/en/connections.json'
import enCore from '../../locales/en/core.json'
import enHome from '../../locales/en/home.json'
import enMember from '../../locales/en/member.json'
import enSquadProfile from '../../locales/en/squad-profile.json'
import enStatistics from '../../locales/en/statistics.json'
import zhAdminActivity from '../../locales/zh-CN/admin-activity.json'
import zhAdminCatalog from '../../locales/zh-CN/admin-catalog.json'
import zhAdminCommunity from '../../locales/zh-CN/admin-community.json'
import zhAdminCore from '../../locales/zh-CN/admin-core.json'
import zhAdminData from '../../locales/zh-CN/admin-data.json'
import zhAdminOnboarding from '../../locales/zh-CN/admin-onboarding.json'
import zhAdminOperations from '../../locales/zh-CN/admin-operations.json'
import zhAdminWorkflows from '../../locales/zh-CN/admin-workflows.json'
import zhActivity from '../../locales/zh-CN/activity.json'
import zhAffiliates from '../../locales/zh-CN/affiliates.json'
import zhCommerce from '../../locales/zh-CN/commerce.json'
import zhCommunity from '../../locales/zh-CN/community.json'
import zhConnections from '../../locales/zh-CN/connections.json'
import zhCore from '../../locales/zh-CN/core.json'
import zhHome from '../../locales/zh-CN/home.json'
import zhMember from '../../locales/zh-CN/member.json'
import zhSquadProfile from '../../locales/zh-CN/squad-profile.json'
import zhStatistics from '../../locales/zh-CN/statistics.json'

const en = {
  ...enCore, ...enAffiliates, ...enActivity, ...enCommerce, ...enCommunity, ...enMember, ...enConnections, ...enHome, ...enAdminCore, ...enAdminOnboarding, ...enAdminData,
  ...enAdminActivity, ...enAdminCommunity, ...enAdminCatalog, ...enAdminOperations, ...enAdminWorkflows, ...enSquadProfile, ...enStatistics,
}
const zhCN: typeof en = {
  ...zhCore, ...zhAffiliates, ...zhActivity, ...zhCommerce, ...zhCommunity, ...zhMember, ...zhConnections, ...zhHome, ...zhAdminCore, ...zhAdminOnboarding, ...zhAdminData,
  ...zhAdminActivity, ...zhAdminCommunity, ...zhAdminCatalog, ...zhAdminOperations, ...zhAdminWorkflows, ...zhSquadProfile, ...zhStatistics,
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
