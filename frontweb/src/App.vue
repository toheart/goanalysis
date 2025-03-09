<template>
  <div id="app">
    <div>
      <nav class="navbar navbar-expand-lg navbar-dark">
        <div class="container-fluid">
          <router-link to="/" class="navbar-brand">
            <i class="bi bi-code-square me-2"></i>{{ $t('nav.title', 'Go代码分析平台') }}
          </router-link>
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav"
            aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
          </button>
          <div class="collapse navbar-collapse" id="navbarNav">
            <ul class="navbar-nav">
              <li class="nav-item">
                <router-link to="/" class="nav-link" exact-active-class="active">
                  <i class="bi bi-house-door me-1"></i>{{ $t('nav.home') }}
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/allgids" class="nav-link" active-class="active">
                  <i class="bi bi-activity me-1"></i>{{ $t('nav.runtimeAnalysis') }}
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/static-analysis" class="nav-link" active-class="active">
                  <i class="bi bi-diagram-3 me-1"></i>{{ $t('nav.staticAnalysis') }}
                </router-link>
              </li>
            </ul>
            
            <!-- 语言切换按钮 - 靠右显示 -->
            <ul class="navbar-nav ms-auto">
              <router-link to="/language" class="nav-link" active-class="active">
                  <i class="bi bi-diagram-3 me-1"></i>{{ $t('welcome.setlanguage.button') }}
                </router-link>
             
            </ul>
          </div>
        </div>
      </nav>
      <main class="container mt-4 fade-in">
        <router-view></router-view> <!-- 动态内容区域 -->
      </main>
      <footer class="footer mt-5 py-3 bg-light">
        <div class="container text-center">
          <span class="text-muted">{{ $t('nav.title', 'Go代码分析平台') }} &copy; 2024</span>
        </div>
      </footer>
    </div>
  </div>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { ref, onMounted, watch, nextTick } from 'vue'

export default {
  setup() {
    const { locale, t } = useI18n({ useScope: 'global' })
    const currentLanguage = ref(locale.value)
    
    // 监听语言变化
    onMounted(() => {
      // 从localStorage读取语言设置
      const savedLocale = localStorage.getItem('locale')
      if (savedLocale) {
        changeLanguage(savedLocale)
      }
    })
    
    // 监听locale变化，确保currentLanguage同步更新
    watch(locale, (newVal) => {
      currentLanguage.value = newVal
    })
    
    // 切换语言
    const changeLanguage = async (lang) => {
      console.log('App.vue - Changing language to:', lang)
      
      // 先更新状态
      locale.value = lang
      currentLanguage.value = lang
      localStorage.setItem('locale', lang)
      
      // 设置HTML lang属性
      document.documentElement.setAttribute('lang', lang)
      
      // 等待下一个DOM更新周期
      await nextTick()
      
      // 触发自定义事件，通知所有组件语言已更改
      window.dispatchEvent(new CustomEvent('languageChanged', { detail: { locale: lang } }))
      
      // 强制刷新部分组件
      const event = new Event('resize')
      window.dispatchEvent(event)
    }
    
    return {
      currentLanguage,
      changeLanguage,
      t
    }
  }
};
</script>

<style>
#app {
  font-family: 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.footer {
  margin-top: 60px;
  border-top: 1px solid #eaeaea;
}

.dropdown-item.active {
  background-color: #f8f9fa;
  color: #212529;
}
</style>
