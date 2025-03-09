import { createRouter, createWebHistory } from 'vue-router';
import TraceViewer from '../components/runtime/TraceViewer.vue';
import TraceDetails from '../components/runtime/TraceDetails.vue';
import StaticAnalysis from '../components/callgraph/StaticAnalysis.vue';
import DbAnalysisDetail from '../components/callgraph/DbAnalysisDetail.vue';
import WelcomePage from '../components/Welcome.vue';
import RuntimeAnalysis from '../components/runtime/RuntimeAnalysis.vue';
import FunctionAnalysis from '../components/runtime/FunctionAnalysis.vue';
import SetLanguage from '../components/Language.vue';

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
        path: '/runtime-analysis/:projectPath?',
        name: 'RuntimeAnalysis',
        component: RuntimeAnalysis,
        props: true
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
    path: '/static-analysis',
    name: 'StaticAnalysis',
    component: StaticAnalysis
  },
  {
    path: '/db-analysis/:path(.*)',
    name: 'DbAnalysisDetail',
    component: DbAnalysisDetail,
    props: route => ({ 
      dbFilePath: route.params.path,
      dbFileName: route.query.name || '',
      dbFileSize: parseInt(route.query.size || '0'),
      dbFileCreateTime: route.query.createTime || ''
    })
  },
  {
    path: '/language',
    name: 'SetLanguage',
    component: SetLanguage
  }
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router; 