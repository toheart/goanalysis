import { createRouter, createWebHistory } from 'vue-router';
import TraceViewer from '../components/TraceViewer.vue';
import TraceDetails from '../components/TraceDetails.vue';
import MermaidViewer from '../components/TraceGraph.vue';
import StaticAnalysis from '../components/StaticAnalysis.vue';
import WelcomePage from '../components/Welcome.vue';
import RuntimeAnalysis from '../components/RuntimeAnalysis.vue';
import FunctionAnalysis from '../components/FunctionAnalysis.vue';

const routes = [
  {
    path: '/',
    name: 'WelcomePage',
    component: WelcomePage,
  },
  {
    path: '/allgids',
    name: 'TraceViewer',
    component: TraceViewer,
    redirect: '/runtime-analysis',
    children: [
      {
        path: '/runtime-analysis',
        name: 'RuntimeAnalysis',
        component: RuntimeAnalysis,
      },
      {
        path: '/function-analysis',
        name: 'FunctionAnalysis',
        component: FunctionAnalysis,
      }
    ]
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
  {
    path: '/static-analysis',
    name: 'StaticAnalysis',
    component: StaticAnalysis
  }
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router; 