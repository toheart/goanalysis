<template>
  <div class="static-analysis container mt-5">
    <h1 class="page-title text-center mb-4">{{ $t('staticAnalysis.title') }}</h1>
    
    <!-- Tab导航 -->
    <ul class="nav nav-tabs mb-4">
      <li class="nav-item">
        <a class="nav-link" :class="{ active: activeTab === 'static' }" href="#" @click.prevent="activeTab = 'static'">
          <i class="bi bi-diagram-3 me-2"></i>{{ $t('staticAnalysis.tabs.static') }}
        </a>
      </li>
      <li class="nav-item">
        <a class="nav-link" :class="{ active: activeTab === 'gitlab' }" href="#" @click.prevent="activeTab = 'gitlab'">
          <i class="bi bi-git me-2"></i>{{ $t('staticAnalysis.tabs.gitlab') }}
        </a>
      </li>
    </ul>
    
    <!-- 静态调用分析Tab -->
    <div v-if="activeTab === 'static'">
      <!-- 如果选择了数据库文件，显示分析详情 -->
      <div v-if="selectedDb && !showDbList">
        <DbAnalysisDetail 
          :dbFilePath="selectedDb"
          :dbFileName="getSelectedDbFileName()"
          :dbFileSize="getSelectedDbFileSize()"
          :dbFileCreateTime="getSelectedDbCreateTime()"
          @back="showDbList = true"
        />
      </div>
      
      <!-- 否则显示数据库文件列表和项目路径输入 -->
      <div v-else>
        <!-- 使用新的合并组件替换原来的表单和监控组件 -->
        <StaticAnalysisWithMonitor
          :initialProjectPath="projectPath"
          :initialAnalysisOptions="analysisOptions"
          @analysis-started="handleAnalysisStarted"
          @task-completed="handleTaskCompleted"
          @status-updated="updateTaskStatus"
          @refresh-db-files="fetchDbFiles"
        />
        
        <!-- 使用指南 -->
        <div v-if="!dbFiles.length && !isAnalyzing && !currentTaskId" class="card mb-4">
          <div class="card-header">
            <h5 class="mb-0"><i class="bi bi-info-circle me-2"></i>{{ $t('staticAnalysis.guide.title') }}</h5>
          </div>
          <div class="card-body">
            <div class="alert alert-info">
              <h4><i class="bi bi-lightbulb me-2"></i>{{ $t('staticAnalysis.guide.welcome') }}</h4>
              <p>{{ $t('staticAnalysis.guide.description') }}</p>
            </div>
            
            <div class="usage-steps">
              <h5><i class="bi bi-1-circle me-2"></i>{{ $t('staticAnalysis.guide.step1') }}</h5>
              <p>{{ $t('staticAnalysis.guide.step1Desc') }}</p>
              
              <h5 class="mt-4"><i class="bi bi-2-circle me-2"></i>{{ $t('staticAnalysis.guide.step2') }}</h5>
              <p>{{ $t('staticAnalysis.guide.step2Desc') }}</p>
              
              <h5 class="mt-4"><i class="bi bi-3-circle me-2"></i>{{ $t('staticAnalysis.guide.step3') }}</h5>
              <p>{{ $t('staticAnalysis.guide.step3Desc') }}</p>
              
              <h5 class="mt-4"><i class="bi bi-4-circle me-2"></i>{{ $t('staticAnalysis.guide.step4') }}</h5>
              <p>{{ $t('staticAnalysis.guide.step4Desc') }}</p>
              <button class="btn btn-primary" @click="fetchDbFiles">
                <i class="bi bi-arrow-clockwise me-2"></i>{{ $t('staticAnalysis.guide.refreshDb') }}
              </button>
            </div>
          </div>
        </div>

        <!-- 数据库文件列表 -->
        <div v-if="dbFiles.length > 0" class="card mb-4">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h5 class="mb-0"><i class="bi bi-database me-2"></i>{{ $t('staticAnalysis.dbList.title') }}</h5>
            <button class="btn btn-sm btn-outline-primary" @click="fetchDbFiles">
              <i class="bi bi-arrow-clockwise me-2"></i>{{ $t('staticAnalysis.dbList.refresh') }}
            </button>
          </div>
          <div class="card-body p-0">
            <div class="table-responsive">
              <table class="table table-hover mb-0">
                <thead>
                  <tr>
                    <th>{{ $t('staticAnalysis.dbList.fileName') }}</th>
                    <th>{{ $t('staticAnalysis.dbList.createTime') }}</th>
                    <th>{{ $t('staticAnalysis.dbList.size') }}</th>
                    <th class="text-center">{{ $t('staticAnalysis.dbList.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="file in dbFiles" :key="file.path">
                    <td>
                      <i class="bi bi-file-earmark-code me-2"></i>
                      {{ file.name }}
                    </td>
                    <td>{{ formatDate(file.createTime) }}</td>
                    <td>{{ formatSize(file.size) }}</td>
                    <td class="text-center">
                      <button 
                        class="btn btn-sm btn-primary" 
                        @click="viewDbAnalysis(file.path)"
                      >
                        <i class="bi bi-search me-1"></i>{{ $t('staticAnalysis.dbList.viewAnalysis') }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- GitLab改动影响分析Tab -->
    <div v-if="activeTab === 'gitlab'">
      <div class="card">
        <div class="card-body text-center py-5">
          <i class="bi bi-tools display-1 text-muted mb-3"></i>
          <h3 class="text-muted">{{ $t('staticAnalysis.gitlab.notAvailable') }}</h3>
          <p class="text-muted">{{ $t('staticAnalysis.gitlab.comingSoon') }}</p>
          <button class="btn btn-outline-primary" @click="activeTab = 'static'">
            <i class="bi bi-arrow-left me-2"></i>{{ $t('staticAnalysis.gitlab.backToStatic') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../../axios'
import DbAnalysisDetail from './DbAnalysisDetail.vue'
import StaticAnalysisWithMonitor from './StaticAnalysisWithMonitor.vue'

export default {
  name: 'StaticAnalysis',
  components: {
    DbAnalysisDetail,
    StaticAnalysisWithMonitor
  },
  data() {
    return {
      activeTab: 'static',
      projectPath: '',
      pathError: '',
      isAnalyzing: false,
      currentTaskId: '',
      dbFiles: [],
      selectedDb: '',
      showDbList: true,
      copiedCommand: false,
      showOptions: true,
      analysisOptions: {
        algo: 'vta',         // 默认使用VTA算法
        isCache: true,       // 默认启用缓存
        outputPath: '',      // 默认输出路径
        cachePath: '',       // 默认缓存路径
        onlyMethod: '',      // 默认分析所有方法
      },
      taskStatus: {
        status: 'processing',
        progress: 0,
        message: '分析任务已启动'
      },
    }
  },
  mounted() {
    this.fetchDbFiles()
  },
  methods: {
    handleAnalysisStarted(taskId) {
      this.currentTaskId = taskId;
      this.isAnalyzing = true;
      console.log('Analysis task started with ID:', taskId);
    },
    
    handleTaskCompleted(status) {
      // 任务完成后的处理
      if (status.status === 'completed') {
        this.fetchDbFiles();
        this.isAnalyzing = false;
        // 可以选择在一段时间后清除任务ID
        setTimeout(() => {
          this.currentTaskId = '';
        }, 10000); // 10秒后清除
      } else if (status.status === 'failed') {
        this.isAnalyzing = false;
        this.pathError = status.message || '分析失败';
      }
    },

    async fetchDbFiles() {
      try {
        const response = await axios.get('/api/static/dbfiles')
        this.dbFiles = response.data.files || []
      } catch (error) {
        console.error('获取数据库文件列表失败:', error)
      }
    },

    viewDbAnalysis(path) {
      // 获取文件信息
      const file = this.dbFiles.find(f => f.path === path);
      if (!file) return;
      
      // 使用路由导航到数据库分析详情页面
      this.$router.push({
        name: 'DbAnalysisDetail',
        params: { path: path },
        query: {
          name: file.name,
          size: file.size,
          createTime: file.createTime
        }
      });
    },

    getSelectedDbFileName() {
      const file = this.dbFiles.find(f => f.path === this.selectedDb);
      return file ? file.name : '';
    },

    getSelectedDbFileSize() {
      const file = this.dbFiles.find(f => f.path === this.selectedDb);
      return file ? file.size : 0;
    },

    getSelectedDbCreateTime() {
      const file = this.dbFiles.find(f => f.path === this.selectedDb);
      return file ? file.createTime : '';
    },

    formatDate(timestamp) {
      return new Date(timestamp).toLocaleString()
    },

    formatSize(bytes) {
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      if (bytes === 0) return '0 Byte'
      const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)))
      return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i]
    },

    // 内部组件方式显示分析详情
    selectDb(path) {
      this.selectedDb = path;
      this.showDbList = false;
    },

    // 更新任务状态，包括进度信息
    updateTaskStatus(status) {
      this.taskStatus = status;
      console.log('Task status updated:', status);
    }
  }
}
</script>

<style scoped>
.usage-steps {
  padding: 1.5rem;
  background-color: #f8f9fa;
  border-radius: var(--border-radius);
}

.code-block {
  position: relative;
  background-color: #2c3e50;
  color: #fff;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
}

.code-block pre {
  margin-bottom: 0;
  white-space: pre-wrap;
}

.copy-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background-color: rgba(255, 255, 255, 0.1);
  color: #fff;
  border: none;
}

.copy-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.analysis-logs {
  max-height: 300px;
  overflow-y: auto;
  background-color: #1e1e1e;
  color: #fff;
  padding: 1rem;
  border-radius: 4px;
  font-family: monospace;
}

.log-entry {
  margin-bottom: 0.5rem;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.nav-tabs .nav-link {
  font-weight: 500;
  color: #495057;
  border-bottom-width: 3px;
}

.nav-tabs .nav-link.active {
  color: #007bff;
  border-bottom-color: #007bff;
}
</style> 