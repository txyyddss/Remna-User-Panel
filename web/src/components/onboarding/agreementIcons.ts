export const agreementIconRegistry = {
  'link-break': 'i-ph-link-break',
  'shield-check': 'i-ph-shield-check',
  'users-three': 'i-ph-users-three',
  warning: 'i-ph-warning',
  'lock-key': 'i-ph-lock-key',
  heart: 'i-ph-heart',
  scales: 'i-ph-scales',
} as const satisfies Record<string, string>

export function agreementIcon(key: string): string {
  return agreementIconRegistry[key as keyof typeof agreementIconRegistry] ?? 'i-ph-shield-check'
}
