import { createPinia } from 'pinia'
import { autoAnimatePlugin } from '@formkit/auto-animate/vue'
import ui from '@nuxt/ui/vue-plugin'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { initializeTelegram, markTelegramReady } from './utils/telegram'
import { t } from './i18n'
import './styles/main.css'

const disposeTelegram = initializeTelegram()

const app = createApp(App)
app.config.globalProperties.$t = t
app.use(createPinia()).use(router).use(ui).use(autoAnimatePlugin)
app.mount('#app')
markTelegramReady()
window.addEventListener('pagehide', disposeTelegram, { once: true })
