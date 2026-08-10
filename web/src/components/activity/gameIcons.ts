import type { Component } from 'vue'
import {
  PhCardsThree,
  PhCoinVertical,
  PhDiceFive,
  PhLightning,
  PhSparkle,
  PhTarget,
  PhTrophy,
} from '@phosphor-icons/vue'

export const gameIconRegistry = {
  dice: PhDiceFive,
  coin: PhCoinVertical,
  cards: PhCardsThree,
  target: PhTarget,
  trophy: PhTrophy,
  lightning: PhLightning,
  sparkle: PhSparkle,
} satisfies Record<string, Component>

export type GameIconKey = keyof typeof gameIconRegistry

export function gameIcon(key: string): Component {
  return gameIconRegistry[key as GameIconKey] ?? gameIconRegistry.dice
}
