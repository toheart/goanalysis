import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n' // 引入i18n配置

// 引入 Bootstrap CSS
import 'bootstrap/dist/css/bootstrap.min.css'
// 引入 Bootstrap JS
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
// 引入全局自定义样式
import './assets/styles.css'

// 如果需要全局注册 Cytoscape 插件
import cytoscape from 'cytoscape';
import dagre from 'cytoscape-dagre';
import cose from 'cytoscape-cose-bilkent';
import popper from 'cytoscape-popper';

cytoscape.use(dagre);
cytoscape.use(cose);
cytoscape.use(popper);

// 创建应用实例
const app = createApp(App)

// 使用插件
app.use(router)
app.use(i18n)

// 挂载应用
app.mount('#app')
