import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Self-hosted fonts (no external Google Fonts dependency — works offline / behind firewalls).
import '@fontsource/inter/400.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import '@fontsource/jetbrains-mono/500.css'
import 'material-symbols/outlined.css'

import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
