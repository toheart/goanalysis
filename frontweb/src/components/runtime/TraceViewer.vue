<template>
  <div class="trace-viewer">
    <h1 class="page-title text-center mb-4">{{ $t('runtimeAnalysis.title') }}</h1>

    <!-- 文件路径输入界面 -->
    <div v-if="!isPathVerified" class="container">
      <!-- 初始验证加载状态 -->
      <div v-if="isInitialVerifying" class="row justify-content-center mb-4">
        <div class="col-md-8">
          <div class="card shadow">
            <div class="card-body p-4 text-center">
              <div class="spinner-border text-primary mb-3" role="status">
                <span class="visually-hidden">加载中...</span>
              </div>
              <h4>{{ $t('runtimeAnalysis.projectPath.starting') }}</h4>
              <p class="text-muted">{{ $t('runtimeAnalysis.projectPath.tip') }}</p>
            </div>
          </div>
        </div>
      </div>
      
      <div v-else class="row justify-content-center">
        <div class="col-md-8">
          <div class="card shadow">
            <div class="card-header">
              <h3 class="mb-0 text-center">{{ $t('runtimeAnalysis.projectPath.title') }}</h3>
            </div>
            <div class="card-body p-4">
              <div class="mb-4">
                <label for="projectPath" class="form-label">{{ $t('runtimeAnalysis.projectPath.label') }}</label>
                <div class="input-group">
                  <span class="input-group-text"><i class="bi bi-folder2-open"></i></span>
                  <input
                    id="projectPath"
                    v-model="projectPath"
                    type="text"
                    class="form-control"
                    :placeholder="$t('runtimeAnalysis.projectPath.placeholder')"
                    :class="{'is-invalid': pathError}"
                  />
                  <div class="invalid-feedback" v-if="pathError">
                    {{ pathError }}
                  </div>
                </div>
                <small class="text-muted">{{ $t('runtimeAnalysis.projectPath.description') }}</small>
              </div>
              <div class="text-center">
                <button 
                  class="btn btn-primary btn-lg"
                  @click="verifyPath"
                  :disabled="isVerifying"
                >
                  <i class="bi" :class="isVerifying ? 'bi-hourglass-split' : 'bi-search'"></i>
                  {{ isVerifying ? $t('runtimeAnalysis.projectPath.verifying') : $t('runtimeAnalysis.projectPath.startAnalysis') }}
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
              <h5 class="mb-0"><i class="bi bi-folder me-2"></i>{{ $t('runtimeAnalysis.projectPath.currentProjectPath') }}</h5>
              <p class="mb-0 text-muted">{{ projectPath }}</p>
            </div>
            <div class="col-md-4 text-end">
              <button class="btn btn-outline-secondary" @click="changePath">
                <i class="bi bi-arrow-repeat"></i> {{ $t('runtimeAnalysis.projectPath.changeProjectPath') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 标签页导航 -->
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <router-link to="/runtime-analysis" class="nav-link" active-class="active">
            <i class="bi bi-activity me-1"></i> {{ $t('runtimeAnalysis.tabs.runtimeAnalysis') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/function-analysis" class="nav-link" active-class="active">
            <i class="bi bi-search me-1"></i> {{ $t('runtimeAnalysis.tabs.functionAnalysis') }}
          </router-link>
        </li>
      </ul>

      <!-- 子路由视图 -->
      <router-view :project-path="projectPath"></router-view>
    </div>
  </div>
</template>

<script>
import axios from '../../axios';
import { useI18n } from 'vue-i18n';

export default {
  data() {
    return {
      projectPath: '',
      isPathVerified: false,
      isVerifying: false,
      pathError: '',
      isInitialVerifying: false,
    };
  },
  setup() {
    const { t, locale } = useI18n({ useScope: 'global' });
    return { t, locale };
  },
  mounted() {
    // 检查本地存储中是否有已验证的路径
    const savedPath = localStorage.getItem('verifiedProjectPath');
    if (savedPath) {
      this.projectPath = savedPath;
      // 在设置路径为已验证之前，先调用API验证路径是否有效
      this.isInitialVerifying = true;
      this.verifyPathSilently(savedPath);
    }
    // 添加语言变化监听
    window.addEventListener('languageChanged', this.handleLanguageChange);
  },
  beforeUnmount() {
    // 移除语言变化监听
    window.removeEventListener('languageChanged', this.handleLanguageChange);
  },
  methods: {
    async verifyPathSilently(path) {
      try {
        // 静默验证路径
        const response = await axios.post('/api/runtime/verify/path', {
          path: path
        });

        console.log('验证路径响应:', response.data);

        if (response.data && response.data.verified) {
          this.isPathVerified = true;
        } else {
          // 如果验证失败，清除本地存储并要求用户重新输入
          localStorage.removeItem('verifiedProjectPath');
          this.isPathVerified = false;
          this.pathError = '保存的项目路径已失效，请重新输入';
        }
      } catch (error) {
        console.error('验证路径失败:', error);
        // 验证出错时，清除本地存储并要求用户重新输入
        localStorage.removeItem('verifiedProjectPath');
        this.isPathVerified = false;
        this.pathError = '验证保存的路径时出错，请重新输入';
      } finally {
        this.isInitialVerifying = false;
      }
    },

    async verifyPath() {
      if (!this.projectPath.trim()) {
        this.pathError = '请输入项目路径';
        return;
      }

      this.isVerifying = true;
      this.pathError = '';

      try {
        // 发送验证请求
        const response = await axios.post('/api/runtime/verify/path', {
          path: this.projectPath
        });

        console.log('验证路径响应:', response.data);

        if (response.data && response.data.verified) {
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