import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// 引入 Bootstrap CSS
import 'bootstrap/dist/css/bootstrap.min.css'
// 引入 Bootstrap JS
import 'bootstrap/dist/js/bootstrap.bundle.min.js'

// 如果需要全局注册 Cytoscape 插件
import cytoscape from 'cytoscape';
import dagre from 'cytoscape-dagre';
import cose from 'cytoscape-cose-bilkent';
import popper from 'cytoscape-popper';

cytoscape.use(dagre);
cytoscape.use(cose);
cytoscape.use(popper);

createApp(App)
  .use(router)
  .mount('#app')
