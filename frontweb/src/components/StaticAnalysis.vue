<template>
  <div class="static-analysis container mt-5">
    <div v-if="!dbFiles.length" class="guide-container">
      <h2>欢迎使用程序调用静态分析</h2>
      <div class="usage-guide">
        <h3>使用说明：</h3>
        <pre>首先需要运行程序收集调用数据</pre>
        <pre>运行以下命令开始收集：</pre>
        <pre>1. 登录到服务器，运行 goanalysis callgraph -d <code>your_project_path</code></pre>
        <pre>2. 运行完成后，刷新页面即可查看分析结果</pre>
        <pre>注意: 分析完成的数据库保存路径为配置文件中 biz.callgraphDB.source 配置的路径</pre>
      </div>
    </div>

    <div v-else class="analysis-content">
      <h2 class="text-center mb-4">程序调用静态分析</h2>
      
      <!-- 数据库文件选择 -->
      <div class="db-selector mb-4">
        <h4>选择数据库文件:</h4>
        <div class="list-group">
          <button
            v-for="file in dbFiles"
            :key="file.path"
            class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
            :class="{ active: selectedDb === file.path }"
            @click="selectDb(file.path)"
          >
            <div>
              <strong>{{ file.name }}</strong>
              <small class="text-muted ms-2">{{ formatDate(file.createTime) }}</small>
            </div>
            <span class="badge bg-primary rounded-pill">{{ formatSize(file.size) }}</span>
          </button>
        </div>
      </div>

      <!-- 分析结果展示 -->
      <div v-if="analysisResult" class="analysis-result">
        <h4>分析结果:</h4>
        <div class="card">
          <div class="card-body">
            <h5 class="card-title">基本信息</h5>
            <ul class="list-group list-group-flush">
              <li class="list-group-item">
                <strong>总函数数:</strong> {{ analysisResult.totalFunctions }}
              </li>
              <li class="list-group-item">
                <strong>调用关系数:</strong> {{ analysisResult.totalCalls }}
              </li>
              <li class="list-group-item">
                <strong>包数量:</strong> {{ analysisResult.totalPackages }}
              </li>
            </ul>
          </div>
        </div>

        <!-- 可以添加更多分析结果展示组件 -->
      </div>

      <!-- Loading 状态 -->
      <div v-if="loading" class="text-center mt-4">
        <div class="spinner-border text-primary" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../axios'

export default {
  name: 'StaticAnalysis',
  data() {
    return {
      dbFiles: [],
      selectedDb: '',
      analysisResult: null,
      loading: false
    }
  },
  mounted() {
    this.fetchDbFiles()
  },
  methods: {
    async fetchDbFiles() {
      try {
        const response = await axios.get('/api/static/dbfiles')
        this.dbFiles = response.data.files || []
      } catch (error) {
        console.error('获取数据库文件列表失败:', error)
      }
    },

    async selectDb(path) {
      this.selectedDb = path
      this.loading = true
      try {
        const response = await axios.post('/api/static/analyze', {
          dbPath: path
        })
        this.analysisResult = response.data
      } catch (error) {
        console.error('分析数据库失败:', error)
      } finally {
        this.loading = false
      }
    },

    formatDate(timestamp) {
      return new Date(timestamp).toLocaleString()
    },

    formatSize(bytes) {
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      if (bytes === 0) return '0 Byte'
      const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)))
      return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i]
    }
  }
}
</script>

<style scoped>
.static-analysis {
  padding: 20px;
}

.guide-container {
  max-width: 800px;
  margin: 0 auto;
}

.usage-guide {
  background-color: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-top: 20px;
}

.usage-guide pre {
  background-color: #2c3e50;
  color: #fff;
  padding: 10px;
  border-radius: 4px;
  margin: 10px 0;
}

.analysis-content {
  max-width: 1200px;
  margin: 0 auto;
}

.db-selector {
  background-color: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
}

.list-group-item {
  cursor: pointer;
  transition: all 0.2s;
}

.list-group-item:hover {
  background-color: #e9ecef;
}

.list-group-item.active {
  background-color: #007bff;
  border-color: #007bff;
}

.analysis-result {
  margin-top: 30px;
}

.card {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style> 