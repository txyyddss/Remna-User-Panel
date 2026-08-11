import { createMemoryHistory, createWebHistory, type RouterHistory } from 'vue-router'

import { isTelegramWebAppDetected } from '@/utils/telegram'
import { pendingRouteRecoveryPath } from './recovery'

function browserRoute(): string {
  return `${window.location.pathname}${window.location.search}`
}

export function createAppHistory(): RouterHistory {
  if (!isTelegramWebAppDetected()) return createWebHistory()

  const history = createMemoryHistory()
  const initialRoute = pendingRouteRecoveryPath() ?? browserRoute()
  if (initialRoute !== '/') history.replace(initialRoute)
  return history
}
