export const gameIconRegistry = {
  dice: 'i-ph-dice-five',
  coin: 'i-ph-coin-vertical',
  cards: 'i-ph-cards-three',
  target: 'i-ph-target',
  trophy: 'i-ph-trophy',
  lightning: 'i-ph-lightning',
  sparkle: 'i-ph-sparkle',
} as const satisfies Record<string, string>

export type GameIconKey = keyof typeof gameIconRegistry

export function gameIcon(key: string): string {
  return gameIconRegistry[key as GameIconKey] ?? gameIconRegistry.dice
}
