<template>
  <div class="set-language">
    <div class="card">
      <div class="card-header">
        <h5>{{ $t('common.language') }} </h5>
      </div>
      <div class="card-body">
        <p>当前语言: <strong>{{ currentLanguage }}</strong></p>
        <p>{{ $t('nav.title') }}</p>
        <p>{{ $t('staticAnalysis.title') }}</p>
        
        <div class="btn-group mt-3">
          <button class="btn btn-primary" @click="changeLanguage('zh')">切换到中文</button>
          <button class="btn btn-secondary" @click="changeLanguage('en')">Switch to English</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'

export default {
  name: 'SetLanguage',
  setup() {
    const { locale, t } = useI18n({ useScope: 'global' })
    
    const currentLanguage = computed(() => {
      return locale.value === 'zh' ? '中文' : 'English'
    })
    
    const changeLanguage = (lang) => {
      locale.value = lang
      localStorage.setItem('locale', lang)
      console.log('Language changed to:', lang)
    }
    
    return {
      currentLanguage,
      changeLanguage,
      t
    }
  }
}
</script>

<style scoped>
.set-language {
  margin: 20px;
}
</style> 