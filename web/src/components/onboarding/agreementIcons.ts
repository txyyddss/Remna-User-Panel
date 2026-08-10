import type { Component } from 'vue'
import {
  PhHeart,
  PhLinkBreak,
  PhLockKey,
  PhScales,
  PhShieldCheck,
  PhUsersThree,
  PhWarning,
} from '@phosphor-icons/vue'

export const agreementIconRegistry = {
  'link-break': PhLinkBreak,
  'shield-check': PhShieldCheck,
  'users-three': PhUsersThree,
  warning: PhWarning,
  'lock-key': PhLockKey,
  heart: PhHeart,
  scales: PhScales,
} satisfies Record<string, Component>

export function agreementIcon(key: string): Component {
  return agreementIconRegistry[key as keyof typeof agreementIconRegistry] ?? PhShieldCheck
}
