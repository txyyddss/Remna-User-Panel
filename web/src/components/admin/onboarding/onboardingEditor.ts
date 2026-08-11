import type { OnboardingAgreement, OnboardingLocalizedContent, OnboardingWelcomeMessage } from '@/api/features'
import { agreementIconRegistry } from '@/components/onboarding/agreementIcons'

export const agreementIconKeys = Object.keys(agreementIconRegistry) as Array<keyof typeof agreementIconRegistry>
export const agreementColors = ['accent', 'success', 'warning', 'danger', 'neutral'] as const

export type AgreementColor = typeof agreementColors[number]
export type EditableAgreement = Omit<OnboardingAgreement, 'icon'> & {
  icon: keyof typeof agreementIconRegistry
  color?: AgreementColor
  pageTitle?: string
}

export interface BilingualWelcome {
  id: string
  english: OnboardingWelcomeMessage
  chinese: OnboardingWelcomeMessage
}

export interface BilingualAgreement {
  id: string
  english: EditableAgreement
  chinese: EditableAgreement
}

let editorIDSequence = 0

function newContentID(prefix: 'welcome' | 'agreement'): string {
  editorIDSequence += 1
  return `${prefix}-${Date.now().toString(36)}-${editorIDSequence.toString(36)}`
}

export function normalizeContentID(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9-]+/g, '-').replace(/^-+/, '').slice(0, 64)
}

function welcomeMessages(content: OnboardingLocalizedContent, locale: 'en' | 'zh-CN'): OnboardingWelcomeMessage[] {
  return content[locale] as OnboardingWelcomeMessage[]
}

function agreements(content: OnboardingLocalizedContent, locale: 'en' | 'zh-CN'): EditableAgreement[] {
  return content[locale] as EditableAgreement[]
}

function acceptedColor(value: unknown): AgreementColor {
  return agreementColors.includes(value as AgreementColor) ? value as AgreementColor : 'warning'
}

function acceptedIcon(value: unknown): keyof typeof agreementIconRegistry {
  return typeof value === 'string' && value in agreementIconRegistry ? value as keyof typeof agreementIconRegistry : 'shield-check'
}

function cleanAgreement(value: EditableAgreement, id: string, icon: keyof typeof agreementIconRegistry, color: AgreementColor, keepPageTitle: boolean): EditableAgreement {
  const { pageTitle, ...agreement } = value
  return {
    ...agreement,
    id,
    icon,
    color,
    ...(keepPageTitle && pageTitle?.trim() ? { pageTitle } : {}),
  }
}

export function welcomePairs(content: OnboardingLocalizedContent): BilingualWelcome[] {
  const english = welcomeMessages(content, 'en')
  const chinese = welcomeMessages(content, 'zh-CN')
  return english.map((message, index) => ({
    id: message.id,
    english: { ...message },
    chinese: { ...(chinese[index] ?? { id: message.id, text: '' }), id: message.id },
  }))
}

export function agreementPairs(content: OnboardingLocalizedContent): BilingualAgreement[] {
  const english = agreements(content, 'en')
  const chinese = agreements(content, 'zh-CN')
  return english.map((agreement, index) => {
    const translated = chinese[index] ?? { id: agreement.id, icon: agreement.icon, title: '', body: '' }
    const icon = acceptedIcon(agreement.icon)
    const color = acceptedColor(agreement.color ?? translated.color)
    return {
      id: agreement.id,
      english: { ...agreement, id: agreement.id, icon, color },
      chinese: { ...translated, id: agreement.id, icon, color },
    }
  })
}

export function welcomeContent(pairs: readonly BilingualWelcome[]): OnboardingLocalizedContent {
  return {
    en: pairs.map(({ id, english }) => ({ ...english, id })),
    'zh-CN': pairs.map(({ id, chinese }) => ({ ...chinese, id })),
  }
}

export function agreementContent(pairs: readonly BilingualAgreement[]): OnboardingLocalizedContent {
  return {
    en: pairs.map(({ id, english }, index) => cleanAgreement(english, id, acceptedIcon(english.icon), acceptedColor(english.color), index === 0)),
    'zh-CN': pairs.map(({ id, chinese }, index) => cleanAgreement(chinese, id, acceptedIcon(chinese.icon), acceptedColor(chinese.color), index === 0)),
  } as OnboardingLocalizedContent
}

export function newWelcomePair(): BilingualWelcome {
  const id = newContentID('welcome')
  return { id, english: { id, text: '' }, chinese: { id, text: '' } }
}

export function newAgreementPair(): BilingualAgreement {
  const id = newContentID('agreement')
  const agreement: EditableAgreement = { id, icon: 'shield-check', color: 'warning', title: '', body: '' }
  return { id, english: { ...agreement }, chinese: { ...agreement } }
}
