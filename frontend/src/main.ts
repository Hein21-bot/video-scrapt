import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'
import { initTheme } from './utils/theme'
import { initLang } from './utils/i18n'

initTheme()
initLang()

createApp(App).use(router).mount('#app')

if ('serviceWorker' in navigator) {
  if (import.meta.env.PROD) {
    navigator.serviceWorker.register('/sw.js').catch(() => {})
  } else {
    // In dev: unregister any existing SW so Vite always serves fresh files
    navigator.serviceWorker.getRegistrations().then(regs => {
      regs.forEach(r => r.unregister())
    })
  }
}
