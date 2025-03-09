<template>
  <div class="welcome-page">
    <div class="container py-5">
      <div class="text-center mb-5">
        <h1 class="display-4 fw-bold text-primary">{{ $t('welcome.title') }}</h1>
        <p class="lead">{{ $t('welcome.subtitle') }}</p>
      </div>

      <div class="row justify-content-center mb-5">
        <div class="col-md-8">
          <div class="card shadow-lg border-0">
            <div class="card-body p-5">
              <h2 class="card-title text-center mb-4"><i class="bi bi-info-circle me-2"></i>{{ $t('welcome.about.title') }}</h2>
              <p class="card-text">
                {{ $t('welcome.about.content') }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <h2 class="text-center mb-4"><i class="bi bi-tools me-2"></i>{{ $t('welcome.features') }}</h2>
      
      <!-- 程序运行分析功能区 -->
      <div class="card mb-5">
        <div class="card-header bg-primary text-white">
          <h3 class="h5 mb-0"><i class="bi bi-activity me-2"></i>{{ $t('welcome.runtime.title') }}</h3>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-7">
              <p class="card-text">
                {{ $t('welcome.runtime.description') }}
              </p>
              
              <!-- 项目插桩区域 -->
              <div class="card mt-3 mb-3">
                <div class="card-header bg-light">
                  <h5 class="mb-0"><i class="bi bi-code-square me-2"></i>{{ $t('welcome.runtime.instrumentation.title') }}</h5>
                </div>
                <div class="card-body">
                  <div class="input-group mb-3">
                    <span class="input-group-text"><i class="bi bi-folder"></i></span>
                    <input 
                      type="text" 
                      class="form-control" 
                      :placeholder="$t('welcome.runtime.instrumentation.placeholder')" 
                      v-model="projectPath"
                      :disabled="isInstrumenting"
                    >
                    <button 
                      class="btn btn-primary" 
                      @click="instrumentProject"
                      :disabled="!projectPath || isInstrumenting"
                    >
                      <span v-if="isInstrumenting" class="spinner-border spinner-border-sm me-2" role="status"></span>
                      <i class="bi bi-code-square me-1" v-else></i>{{ isInstrumenting ? $t('welcome.runtime.instrumentation.processing') : $t('welcome.runtime.instrumentation.button') }}
                    </button>
                  </div>
                  <small class="text-muted">{{ $t('welcome.runtime.instrumentation.hint') }}</small>
                  
                  <!-- 插桩结果提示 -->
                  <div v-if="instrumentError" class="alert alert-danger mt-3">
                    <i class="bi bi-exclamation-triangle-fill me-2"></i>{{ instrumentError }}
                  </div>
                  <div v-if="instrumentSuccess" class="alert alert-success mt-3">
                    <i class="bi bi-check-circle-fill me-2"></i>{{ instrumentSuccess }}
                    <div class="mt-2">
                      <router-link to="/allgids" class="btn btn-sm btn-primary">
                        <i class="bi bi-arrow-right me-1"></i>{{ $t('welcome.runtime.instrumentation.viewResults') }}
                      </router-link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="col-md-5 d-flex align-items-center justify-content-center">
              <div class="text-center">
                <i class="bi bi-activity display-1 text-primary mb-3"></i>
                <div>
                  <router-link to="/allgids" class="btn btn-primary btn-lg">
                    <i class="bi bi-arrow-right me-1"></i>{{ $t('welcome.runtime.viewAnalysis') }}
                  </router-link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 程序调用静态分析功能区 -->
      <div class="card mb-5">
        <div class="card-header bg-success text-white">
          <h3 class="h5 mb-0"><i class="bi bi-diagram-3 me-2"></i>{{ $t('welcome.static.title') }}</h3>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-7">
              <p class="card-text">
                {{ $t('welcome.static.description') }}
              </p>
            </div>
            <div class="col-md-5 d-flex align-items-center justify-content-center">
              <div class="text-center">
                <i class="bi bi-diagram-3 display-1 text-success mb-3"></i>
                <div>
                  <router-link to="/static-analysis" class="btn btn-success btn-lg">
                    <i class="bi bi-arrow-right me-1"></i>{{ $t('welcome.static.startAnalysis') }}
                  </router-link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../axios';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

export default {
  name: 'WelcomePage',
  setup() {
    const { t } = useI18n({ useScope: 'global' });
    
    const projectPath = ref('');
    const isInstrumenting = ref(false);
    const instrumentError = ref('');
    const instrumentSuccess = ref('');
    
    const instrumentProject = async () => {
      if (!projectPath.value) {
        instrumentError.value = '请输入项目路径';
        return;
      }
      
      instrumentError.value = '';
      instrumentSuccess.value = '';
      isInstrumenting.value = true;
      
      try {
        const response = await axios.post('/api/runtime/instrument', {
          path: projectPath.value
        });
        
        if (response.data.success) {
          instrumentSuccess.value = response.data.message || '项目插桩成功，现在可以运行您的程序进行分析';
        } else {
          instrumentError.value = response.data.message || '插桩失败';
        }
      } catch (error) {
        instrumentError.value = '插桩过程出错: ' + (error.response?.data?.message || error.message);
      } finally {
        isInstrumenting.value = false;
      }
    };
    
    return {
      projectPath,
      isInstrumenting,
      instrumentError,
      instrumentSuccess,
      instrumentProject,
      t
    };
  }
}
</script>

<style scoped>
.welcome-page {
  padding-top: 2rem;
  padding-bottom: 4rem;
}

.feature-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  font-size: 1.5rem;
}

.card {
  transition: all 0.3s ease;
  border-radius: 0.5rem;
  overflow: hidden;
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
}

.card:hover {
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
  transform: translateY(-5px);
}

.card-header {
  padding: 1rem 1.5rem;
}

.card-body {
  padding: 1.5rem;
}

.display-1 {
  font-size: 4rem;
  opacity: 0.8;
}
</style> 