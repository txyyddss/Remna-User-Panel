import { onScopeDispose, shallowRef } from 'vue'
import type { NavigationMenuItem } from '@nuxt/ui'
import type { Router } from 'vue-router'

import { mediaQueryList } from '@/utils/browserCompatibility'

function navigationPaths(items: NavigationMenuItem[]): string[] {
  return items.flatMap(item => [
    ...(typeof item.to === 'string' ? [item.to] : []),
    ...navigationPaths(item.children ?? []),
  ])
}

/** Serialize Vue/CSS slides, including Telegram WebViews without View Transitions. */
export function usePageTransition(router: Router, items: () => NavigationMenuItem[]) {
  const name = shallowRef('page-forward')
  const reducedMotion = mediaQueryList('(prefers-reduced-motion: reduce)')
  let pending = Promise.resolve()
  let release: (() => void) | undefined
  let destination = ''
  let version = 0
  let disposed = false

  function finish(): void {
    release?.()
    release = undefined
  }
  function leave(element: Element): void {
    if (!(element instanceof HTMLElement)) return
    element.inert = true
    element.setAttribute('aria-hidden', 'true')
    element.querySelector('#main-content')?.removeAttribute('id')
  }
  const removeGuard = router.beforeResolve(async (to, from) => {
    const request = ++version
    await pending
    if (disposed || request !== version) return false
    if (to.path === from.path || to.meta.immersive || from.meta.immersive
      || reducedMotion?.matches) return
    const paths = navigationPaths(items())
    const rank = (path: string) => paths.indexOf(path.startsWith('/admin/users/') ? '/admin/users' : path)
    name.value = rank(to.path) >= 0 && rank(to.path) < rank(from.path) ? 'page-backward' : 'page-forward'
    destination = to.fullPath
    pending = new Promise<void>(resolve => { release = resolve })
  })
  const removeAfter = router.afterEach((to, _from, failure) => {
    if (failure && to.fullPath === destination) finish()
  })
  const removeError = router.onError(finish)
  onScopeDispose(() => {
    disposed = true
    removeGuard()
    removeAfter()
    removeError()
    finish()
  })
  return { name, finish, leave }
}
