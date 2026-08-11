void bootstrap()

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
    app.mount('#app')
    return
  }

  const [
    { createPinia },
    { autoAnimatePlugin },
    { default: ui },
    { createApp },
    { default: App },
    { default: router },
    { t },
    { initializeTelegram, markTelegramReady },
  ] = await Promise.all([
    import('pinia'),
    import('@formkit/auto-animate/vue'),
    import('@nuxt/ui/vue-plugin'),
    import('vue'),
    import('./App.vue'),
    import('./router'),
    import('./i18n'),
    import('./utils/telegram'),
  ])
  const disposeTelegram = initializeTelegram()
  const app = createApp(App)
  app.config.globalProperties.$t = t
  app.use(createPinia()).use(router).use(ui).use(autoAnimatePlugin)
  app.mount('#app')
  markTelegramReady()
  window.addEventListener('pagehide', disposeTelegram, { once: true })
}
