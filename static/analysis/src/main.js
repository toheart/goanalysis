import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// 引入 Bootstrap CSS
import 'bootstrap/dist/css/bootstrap.min.css'
// 引入 Bootstrap JS
import 'bootstrap/dist/js/bootstrap.bundle.min.js'



createApp(App)
  .use(router)
  .mount('#app')
