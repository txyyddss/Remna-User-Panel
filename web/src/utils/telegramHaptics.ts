import { getTelegramWebApp, supportsTelegramVersion, tryTelegramCall } from './telegram'

export type HapticImpact = 'light' | 'medium' | 'heavy'

export function haptic(type: HapticImpact = 'light'): void {
  const app = getTelegramWebApp()
  if (app?.HapticFeedback && supportsTelegramVersion('6.1')) tryTelegramCall(() => app.HapticFeedback?.impactOccurred(type))
}

export function selectionHaptic(): void {
  const app = getTelegramWebApp()
  if (!app?.HapticFeedback || !supportsTelegramVersion('6.1')) return
  tryTelegramCall(() => {
    if (app.HapticFeedback?.selectionChanged) app.HapticFeedback.selectionChanged()
    else app.HapticFeedback?.impactOccurred('light')
  })
}

function hapticImpactFor(element: Element): HapticImpact {
  const value = element.getAttribute('data-haptic')
  return value === 'medium' || value === 'heavy' ? value : 'light'
}

function handleHapticClick(event: MouseEvent): void {
  if (!(event.target instanceof Element)) return
  const target = event.target.closest('[data-haptic]')
  if (!target || target.hasAttribute('disabled') || target.getAttribute('aria-disabled') === 'true') return
  haptic(hapticImpactFor(target))
}

export function installHapticClickFeedback(): () => void {
  document.addEventListener('click', handleHapticClick, true)
  return () => document.removeEventListener('click', handleHapticClick, true)
}

export function notifyHaptic(type: 'error' | 'success' | 'warning'): void {
  const app = getTelegramWebApp()
  if (app?.HapticFeedback && supportsTelegramVersion('6.1')) tryTelegramCall(() => app.HapticFeedback?.notificationOccurred(type))
}

export function notifyBetOutcome(outcome: 'win' | 'loss'): void {
  const app = getTelegramWebApp()
  if (!app?.HapticFeedback || !supportsTelegramVersion('6.1')) return
  tryTelegramCall(() => {
    app.HapticFeedback?.impactOccurred(outcome === 'win' ? 'heavy' : 'light')
    app.HapticFeedback?.notificationOccurred(outcome === 'win' ? 'success' : 'warning')
  })
}
