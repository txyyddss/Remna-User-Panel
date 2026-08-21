import { getTelegramWebApp, supportsTelegramVersion, tryTelegramCall } from './telegram'

export type HapticImpact = 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'
export type HapticIntent =
  | 'action'
  | 'confirm'
  | 'copy'
  | 'destructive'
  | 'dismiss'
  | 'navigate'
  | 'open'
  | 'refresh'
  | 'retry'
  | 'zoom'

const impactByIntent: Readonly<Record<HapticIntent, HapticImpact>> = {
  action: 'medium',
  confirm: 'rigid',
  copy: 'light',
  destructive: 'heavy',
  dismiss: 'soft',
  navigate: 'soft',
  open: 'soft',
  refresh: 'light',
  retry: 'light',
  zoom: 'light',
}

export function haptic(intent: HapticIntent = 'action'): void {
  const app = getTelegramWebApp()
  if (app?.HapticFeedback && supportsTelegramVersion('6.1')) {
    tryTelegramCall(() => app.HapticFeedback?.impactOccurred(impactByIntent[intent]))
  }
}

export function selectionHaptic(): void {
  const app = getTelegramWebApp()
  if (!app?.HapticFeedback || !supportsTelegramVersion('6.1')) return
  tryTelegramCall(() => {
    if (app.HapticFeedback?.selectionChanged) app.HapticFeedback.selectionChanged()
    else app.HapticFeedback?.impactOccurred('light')
  })
}

function hapticIntentFor(element: Element): HapticIntent | undefined {
  const value = element.getAttribute('data-haptic') as HapticIntent | null
  return value && value in impactByIntent ? value : undefined
}

function handleHapticClick(event: MouseEvent): void {
  if (!(event.target instanceof Element)) return
  const target = event.target.closest('[data-haptic]')
  if (!target || target.hasAttribute('disabled') || target.getAttribute('aria-disabled') === 'true') return
  const intent = hapticIntentFor(target)
  if (intent) haptic(intent)
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
    app.HapticFeedback?.impactOccurred(outcome === 'win' ? impactByIntent.confirm : impactByIntent.action)
    app.HapticFeedback?.notificationOccurred(outcome === 'win' ? 'success' : 'warning')
  })
}
