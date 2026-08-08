import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { initializeTelegram } from './utils/telegram'
import { t } from './i18n'
import './styles/main.css'

initializeTelegram()

const app = createApp(App)
app.config.globalProperties.$t = t
app.use(createPinia()).use(router)
app.mount('#app')
