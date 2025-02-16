import { createRouter, createWebHistory } from 'vue-router';
import TraceViewer from '../components/TraceViewer.vue';
import TraceDetails from '../components/TraceDetails.vue';
import MermaidViewer from '../components/MermaidViewer.vue';

const routes = [
  {
    path: '/allgids',
    name: 'TraceViewer',
    component: TraceViewer,
  },
  {
    path: '/trace/:gid',
    name: 'TraceDetails',
    component: TraceDetails,
  },
  {
    path: '/mermaid/:gid',
    name: 'MermaidViewer',
    component: MermaidViewer,
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router; 