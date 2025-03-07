<template>
  <div class="static-analysis container mt-5">
    <h1 class="page-title text-center mb-4">程序调用静态分析</h1>
    
    <!-- 路径输入和验证部分 -->
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-folder2-open me-2"></i>项目路径</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-8">
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-folder"></i></span>
              <input 
                type="text" 
                class="form-control" 
                placeholder="请输入Go项目路径" 
                v-model="projectPath"
                :disabled="isAnalyzing"
              >
              <button 
                class="btn btn-primary" 
                @click="verifyAndAnalyze" 
                :disabled="!projectPath || isAnalyzing"
              >
                <span v-if="isAnalyzing" class="spinner-border spinner-border-sm me-2" role="status"></span>
                {{ isAnalyzing ? '分析中...' : '开始分析' }}
              </button>
            </div>
            <div v-if="pathError" class="alert alert-danger mt-2">
              <i class="bi bi-exclamation-triangle-fill me-2"></i>{{ pathError }}
            </div>
            <div v-if="analysisProgress > 0 && analysisProgress < 100" class="mt-3">
              <div class="progress">
                <div 
                  class="progress-bar progress-bar-striped progress-bar-animated" 
                  role="progressbar" 
                  :style="{width: analysisProgress + '%'}" 
                  :aria-valuenow="analysisProgress" 
                  aria-valuemin="0" 
                  aria-valuemax="100"
                >
                  {{ analysisProgress }}%
                </div>
              </div>
              <p class="text-muted mt-2">正在分析项目，请稍候...</p>
            </div>
          </div>
          <div class="col-md-4">
            <div class="alert alert-info mb-0">
              <h6><i class="bi bi-info-circle me-2"></i>提示</h6>
              <p class="mb-0 small">输入Go项目的完整路径，系统将自动分析项目的调用关系并生成数据库。</p>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="!dbFiles.length && !isAnalyzing" class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-info-circle me-2"></i>使用指南</h5>
      </div>
      <div class="card-body">
        <div class="alert alert-info">
          <h4><i class="bi bi-lightbulb me-2"></i>欢迎使用程序调用静态分析</h4>
          <p>本工具可以帮助您分析Go程序的调用关系，生成静态调用图。</p>
        </div>
        
        <div class="usage-steps">
          <h5><i class="bi bi-1-circle me-2"></i>输入项目路径</h5>
          <p>在上方输入框中输入您要分析的Go项目路径，点击"开始分析"按钮。</p>
          
          <h5 class="mt-4"><i class="bi bi-2-circle me-2"></i>等待分析完成</h5>
          <p>系统将自动分析项目并生成数据库，分析完成后可以查看结果。</p>
          
          <h5 class="mt-4"><i class="bi bi-3-circle me-2"></i>查看分析结果</h5>
          <p>分析完成后，您可以选择数据库查看详细的分析结果。</p>
          <button class="btn btn-primary" @click="fetchDbFiles">
            <i class="bi bi-arrow-clockwise me-2"></i>刷新数据库列表
          </button>
        </div>
      </div>
    </div>

    <div v-else-if="dbFiles.length > 0" class="row">
      <!-- 数据库文件选择 -->
      <div class="col-md-4">
        <div class="card mb-4">
          <div class="card-header">
            <h5 class="mb-0"><i class="bi bi-database me-2"></i>数据库文件</h5>
          </div>
          <div class="card-body p-0">
            <div class="list-group list-group-flush">
              <button
                v-for="file in dbFiles"
                :key="file.path"
                class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
                :class="{ active: selectedDb === file.path }"
                @click="selectDb(file.path)"
              >
                <div>
                  <i class="bi bi-file-earmark-code me-2"></i>
                  <strong>{{ file.name }}</strong>
                  <small class="d-block text-muted mt-1">{{ formatDate(file.createTime) }}</small>
                </div>
                <span class="badge bg-primary rounded-pill">{{ formatSize(file.size) }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 分析结果展示 -->
      <div class="col-md-8">
        <div v-if="loading" class="card mb-4">
          <div class="card-body text-center py-5">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">Loading...</span>
            </div>
            <p class="mt-3">正在分析数据库，请稍候...</p>
          </div>
        </div>
        
        <div v-else-if="analysisResult" class="analysis-result">
          <!-- 基本信息卡片 -->
          <div class="card mb-4">
            <div class="card-header">
              <h5 class="mb-0"><i class="bi bi-bar-chart-line me-2"></i>基本统计</h5>
            </div>
            <div class="card-body">
              <div class="row">
                <div class="col-md-4">
                  <div class="stat-card">
                    <div class="stat-icon">
                      <i class="bi bi-code-square"></i>
                    </div>
                    <div class="stat-value">{{ analysisResult.totalFunctions }}</div>
                    <div class="stat-label">总函数数</div>
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="stat-card">
                    <div class="stat-icon">
                      <i class="bi bi-arrow-left-right"></i>
                    </div>
                    <div class="stat-value">{{ analysisResult.totalCalls }}</div>
                    <div class="stat-label">调用关系数</div>
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="stat-card">
                    <div class="stat-icon">
                      <i class="bi bi-folder2"></i>
                    </div>
                    <div class="stat-value">{{ analysisResult.totalPackages }}</div>
                    <div class="stat-label">包数量</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 包依赖关系 -->
          <div class="card mb-4" v-if="analysisResult.packageDependencies">
            <div class="card-header">
              <h5 class="mb-0"><i class="bi bi-diagram-3 me-2"></i>包依赖关系</h5>
            </div>
            <div class="card-body">
              <div class="table-responsive">
                <table class="table table-hover">
                  <thead>
                    <tr>
                      <th>源包</th>
                      <th>目标包</th>
                      <th class="text-center">调用次数</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(dep, index) in analysisResult.packageDependencies.slice(0, 10)" :key="index">
                      <td><code>{{ dep.source }}</code></td>
                      <td><code>{{ dep.target }}</code></td>
                      <td class="text-center">
                        <span class="badge bg-primary">{{ dep.count }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-if="analysisResult.packageDependencies.length > 10" class="text-center mt-3">
                <button class="btn btn-sm btn-outline-primary" @click="showMoreDeps = !showMoreDeps">
                  {{ showMoreDeps ? '显示更少' : '显示更多' }}
                </button>
              </div>
            </div>
          </div>
          
          <!-- 热点函数 -->
          <div class="card mb-4" v-if="analysisResult.hotFunctions">
            <div class="card-header">
              <h5 class="mb-0"><i class="bi bi-fire me-2"></i>热点函数</h5>
            </div>
            <div class="card-body">
              <div class="table-responsive">
                <table class="table table-hover">
                  <thead>
                    <tr>
                      <th>函数名</th>
                      <th class="text-center">被调用次数</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(func, index) in analysisResult.hotFunctions.slice(0, 10)" :key="index">
                      <td><code>{{ func.name }}</code></td>
                      <td class="text-center">
                        <span class="badge bg-danger">{{ func.callCount }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
        
        <div v-else-if="selectedDb" class="card mb-4">
          <div class="card-body text-center py-5">
            <i class="bi bi-exclamation-circle text-warning display-1"></i>
            <h4 class="mt-3">暂无分析结果</h4>
            <p>选择的数据库文件未返回分析结果，请尝试其他文件。</p>
          </div>
        </div>
        
        <div v-else class="card mb-4">
          <div class="card-body text-center py-5">
            <i class="bi bi-arrow-left-circle text-primary display-1"></i>
            <h4 class="mt-3">请选择数据库文件</h4>
            <p>从左侧列表中选择一个数据库文件进行分析。</p>
          </div>
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
      projectPath: '',
      pathError: '',
      isAnalyzing: false,
      analysisProgress: 0,
      progressInterval: null,
      dbFiles: [],
      selectedDb: '',
      analysisResult: null,
      loading: false,
      showMoreDeps: false,
      copiedCommand: false
    }
  },
  mounted() {
    this.fetchDbFiles()
  },
  methods: {
    async verifyAndAnalyze() {
      this.pathError = '';
      this.isAnalyzing = true;
      this.analysisProgress = 0;
      
      try {
        // 开始分析，启动进度条模拟
        this.startProgressSimulation();
        
        // 调用后端执行callgraph分析
        const response = await axios.post('/api/static/analyze/path', {
          path: this.projectPath
        });
        
        // 检查分析结果
        if (!response.data.success) {
          this.pathError = response.data.message || '分析失败';
          clearInterval(this.progressInterval);
          this.isAnalyzing = false;
          this.analysisProgress = 0;
          return;
        }
        
        // 分析完成，停止进度条
        clearInterval(this.progressInterval);
        this.analysisProgress = 100;
        
        // 刷新数据库列表
        setTimeout(() => {
          this.fetchDbFiles();
          this.isAnalyzing = false;
          this.analysisProgress = 0;
        }, 1000);
        
      } catch (error) {
        this.pathError = '分析过程出错: ' + (error.response?.data?.message || error.message);
        clearInterval(this.progressInterval);
        this.isAnalyzing = false;
        this.analysisProgress = 0;
      }
    },
    
    startProgressSimulation() {
      // 模拟进度条，实际进度应该由后端提供
      this.analysisProgress = 5;
      this.progressInterval = setInterval(() => {
        if (this.analysisProgress < 90) {
          this.analysisProgress += Math.floor(Math.random() * 5) + 1;
        }
      }, 1000);
    },
    
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

.list-group-item.active {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.stat-card {
  text-align: center;
  padding: 1.5rem;
  border-radius: var(--border-radius);
  background-color: #f8f9fa;
  transition: var(--transition);
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: var(--box-shadow);
}

.stat-icon {
  font-size: 2rem;
  color: var(--primary-color);
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--secondary-color);
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.9rem;
  color: #6c757d;
}

@media (max-width: 768px) {
  .row {
    flex-direction: column;
  }
  
  .stat-card {
    margin-bottom: 1rem;
  }
}
</style> 