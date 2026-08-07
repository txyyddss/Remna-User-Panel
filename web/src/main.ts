import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { initializeTelegram } from './utils/telegram'
import './styles/main.css'

initializeTelegram()

createApp(App)
  .use(createPinia())
  .use(router)
  .mount('#app')
