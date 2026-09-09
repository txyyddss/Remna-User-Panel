import { showBootstrapFailure } from './utils/bootstrapFallback'

void bootstrap().catch(showBootstrapFailure)

async function bootstrap(): Promise<void> {
  const compatibility = await import('./utils/browserCompatibility')
  compatibility.installBrowserCompatibility()
  await import('./styles/main.css')

  if (compatibility.missingBrowserCapabilities().length > 0) {
    const [{ createApp }, { default: BrowserCapabilityGate }, { t }] = await Promise.all([
      import('vue'),
      import('./components/session/BrowserCapabilityGate.vue'),
      import('./i18n'),
    ])

    const app = createApp(BrowserCapabilityGate)
    app.config.globalProperties.$t = t
    app.config.errorHandler = () => showBootstrapFailure()
    app.mount('#app')
    return
  }

  const { initializeTelegram, markTelegramReady, waitForTelegramContext } = await import('./utils/telegram')
  await waitForTelegramContext()
  const disposeTelegram = initializeTelegram()

  const [
    { createPinia },
    { default: ui },
    { createApp },
    { default: App },
    { default: router },
    { t },
  ] = await Promise.all([
    import('pinia'),
    import('@nuxt/ui/vue-plugin'),
    import('vue'),
    import('./App.vue'),
    import('./router'),
    import('./i18n'),
  ])

  // Do not import AutoAnimate until its actual runtime prerequisites exist.
  // @formkit/auto-animate touches observer/animation APIs internally; a
  // partially capable Telegram WebView must still be able to boot the app.
  let autoAnimatePlugin: Awaited<
    ReturnType<typeof importAutoAnimatePlugin>
  > | null = null

  if (compatibility.supportsAutoAnimate()) {
    try {
      autoAnimatePlugin = await importAutoAnimatePlugin()
    } catch {
      autoAnimatePlugin = null
    }
  }

  const app = createApp(App)
  app.config.globalProperties.$t = t
  app.config.errorHandler = () => showBootstrapFailure()

  app.use(createPinia())
  app.use(router)
  app.use(ui)

  if (autoAnimatePlugin) {
    app.use(autoAnimatePlugin)
  } else {
    // Keep v-auto-animate templates valid while degrading to static layout.
    app.directive('auto-animate', {})
  }

  app.mount('#app')
  markTelegramReady()

  window.addEventListener('pagehide', disposeTelegram, { once: true })
}

async function importAutoAnimatePlugin() {
  const module = await import('@formkit/auto-animate/vue')
  return module.autoAnimatePlugin
}
