<template>
  <div class="trace-viewer">
    <h1 class="page-title text-center mb-4">程序运行分析</h1>

    <!-- 文件路径输入界面 -->
    <div v-if="!isPathVerified" class="container">
      <div class="row justify-content-center">
        <div class="col-md-8">
          <div class="card shadow">
            <div class="card-header">
              <h3 class="mb-0 text-center">请输入项目路径</h3>
            </div>
            <div class="card-body p-4">
              <div class="mb-4">
                <label for="projectPath" class="form-label">项目路径</label>
                <div class="input-group">
                  <span class="input-group-text"><i class="bi bi-folder2-open"></i></span>
                  <input
                    id="projectPath"
                    v-model="projectPath"
                    type="text"
                    class="form-control"
                    placeholder="例如: /path/to/your/project"
                    :class="{'is-invalid': pathError}"
                  />
                  <div class="invalid-feedback" v-if="pathError">
                    {{ pathError }}
                  </div>
                </div>
                <small class="text-muted">输入要分析的Go项目的完整路径</small>
              </div>
              <div class="text-center">
                <button 
                  class="btn btn-primary btn-lg"
                  @click="verifyPath"
                  :disabled="isVerifying"
                >
                  <i class="bi" :class="isVerifying ? 'bi-hourglass-split' : 'bi-search'"></i>
                  {{ isVerifying ? '验证中...' : '开始分析' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 分析界面 -->
    <div v-else class="container">
      <!-- 项目信息 -->
      <div class="card mb-4">
        <div class="card-body">
          <div class="row align-items-center">
            <div class="col-md-8">
              <h5 class="mb-0"><i class="bi bi-folder me-2"></i>当前项目路径</h5>
              <p class="mb-0 text-muted">{{ projectPath }}</p>
            </div>
            <div class="col-md-4 text-end">
              <button class="btn btn-outline-secondary" @click="changePath">
                <i class="bi bi-arrow-repeat"></i> 更换项目
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 标签页导航 -->
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <router-link to="/runtime-analysis" class="nav-link" active-class="active">
            <i class="bi bi-activity me-1"></i>运行时分析大盘
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/function-analysis" class="nav-link" active-class="active">
            <i class="bi bi-search me-1"></i>函数查询分析
          </router-link>
        </li>
      </ul>

      <!-- 子路由视图 -->
      <router-view :project-path="projectPath"></router-view>
    </div>
  </div>
</template>

<script>
import axios from '../axios';

export default {
  data() {
    return {
      projectPath: '',
      isPathVerified: false,
      isVerifying: false,
      pathError: '',
    };
  },
  mounted() {
    // 检查本地存储中是否有已验证的路径
    const savedPath = localStorage.getItem('verifiedProjectPath');
    if (savedPath) {
      this.projectPath = savedPath;
      this.isPathVerified = true;
    }
  },
  methods: {
    async verifyPath() {
      if (!this.projectPath.trim()) {
        this.pathError = '请输入项目路径';
        return;
      }

      this.isVerifying = true;
      this.pathError = '';

      try {
        // 发送验证请求
        const response = await axios.post('/api/verify/path', {
          path: this.projectPath
        });

        if (response.data.verified) {
          this.isPathVerified = true;
          // 保存验证通过的路径
          localStorage.setItem('verifiedProjectPath', this.projectPath);
          // 导航到运行时分析页面
          this.$router.push('/runtime-analysis');
        } else {
          this.pathError = response.data.message || '项目路径验证失败';
        }
      } catch (error) {
        this.pathError = '验证过程出错: ' + (error.response?.data?.message || error.message);
      } finally {
        this.isVerifying = false;
      }
    },

    changePath() {
      this.isPathVerified = false;
      localStorage.removeItem('verifiedProjectPath');
    }
  }
};
</script>

<style scoped>
.nav-tabs .nav-link {
  font-weight: 500;
  color: #6c757d;
  padding: 0.75rem 1.25rem;
  border-radius: 0;
  transition: all 0.2s ease;
}

.nav-tabs .nav-link.active {
  color: #0d6efd;
  border-bottom: 2px solid #0d6efd;
  background-color: transparent;
}

.nav-tabs .nav-link:hover:not(.active) {
  background-color: rgba(13, 110, 253, 0.05);
  border-color: transparent;
}
</style>