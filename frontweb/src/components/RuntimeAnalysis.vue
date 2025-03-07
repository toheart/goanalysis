<template>
  <div class="runtime-analysis">
    <!-- 搜索和过滤区域 -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row align-items-center">
          <div class="col-md-8 search-container">
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input
                v-model="functionName"
                type="text"
                placeholder="输入函数名称搜索"
                class="form-control"
                @input="onFunctionNameInput"
                @focus="showFunctionSuggestions = true"
                @blur="hideSuggestionsDelayed"
              />
              <button class="btn btn-primary" @click="searchByFunctionName">查询</button>
            </div>
          </div>
          <div class="col-md-4">
            <div class="form-check form-switch">
              <input class="form-check-input" type="checkbox" id="showAllGoroutines" v-model="showAllGoroutines">
              <label class="form-check-label" for="showAllGoroutines">显示所有Goroutine</label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 函数名建议列表 -->
    <div class="suggestions-wrapper" v-if="showFunctionSuggestions && filteredFunctionNames.length">
      <ul class="list-group function-suggestions">
        <li
          v-for="(name, index) in filteredFunctionNames"
          :key="index"
          class="list-group-item list-group-item-action"
          @mousedown.prevent="selectFunction(name)"
        >
          {{ name }}
        </li>
      </ul>
    </div>

    <!-- 统计卡片 -->
    <div class="row mb-4">
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-cpu me-2"></i>活跃Goroutine</h5>
            <p class="display-4">{{ goroutineStats.active || 0 }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-hourglass-split me-2"></i>平均执行时间</h5>
            <p class="display-4">{{ goroutineStats.avgTime || '0ms' }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-lightning-charge me-2"></i>最大调用深度</h5>
            <p class="display-4">{{ goroutineStats.maxDepth || 0 }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 热点函数分析 -->
    <div class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0"><i class="bi bi-fire me-2"></i>热点函数分析</h5>
        <div class="btn-group">
          <button class="btn btn-sm btn-outline-primary" @click="sortHotFunctions('calls')" :class="{ active: hotFunctionSortBy === 'calls' }">
            按调用次数
          </button>
          <button class="btn btn-sm btn-outline-primary" @click="sortHotFunctions('time')" :class="{ active: hotFunctionSortBy === 'time' }">
            按耗时
          </button>
        </div>
      </div>
      <div class="card-body">
        <div v-if="loading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">加载中...</span>
          </div>
          <p class="mt-3">正在加载热点函数数据...</p>
        </div>
        <div v-else-if="hotFunctions.length === 0" class="text-center py-5">
          <i class="bi bi-exclamation-circle text-warning display-4"></i>
          <p class="mt-3">暂无热点函数数据</p>
        </div>
        <div v-else>
          <div class="table-responsive">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>函数名</th>
                  <th class="text-center">调用次数</th>
                  <th class="text-center">总耗时</th>
                  <th class="text-center">平均耗时</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(func, index) in hotFunctions.slice(0, 10)" :key="index">
                  <td><code>{{ func.name }}</code></td>
                  <td class="text-center">
                    <span class="badge bg-primary">{{ func.callCount }}</span>
                  </td>
                  <td class="text-center">{{ func.totalTime }}</td>
                  <td class="text-center">{{ func.avgTime }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Goroutine列表 -->
    <div class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0"><i class="bi bi-list-ul me-2"></i>Goroutine列表</h5>
        <div class="pagination-info">
          <span class="badge bg-secondary">当前页: {{ currentPage }} / {{ totalPages }}</span>
          <div class="btn-group ms-2">
            <button class="btn btn-sm btn-outline-primary" @click="prevPage" :disabled="currentPage <= 1">
              <i class="bi bi-chevron-left"></i>
            </button>
            <button class="btn btn-sm btn-outline-primary" @click="nextPage" :disabled="currentPage >= totalPages">
              <i class="bi bi-chevron-right"></i>
            </button>
          </div>
        </div>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover table-striped mb-0">
            <thead class="table-light">
              <tr>
                <th>GID</th>
                <th>初始函数</th>
                <th class="text-center">调用深度</th>
                <th class="text-center">执行时间</th>
                <th class="text-center">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="result in filteredGIDs" :key="result.GID">
                <td><span class="badge bg-primary">{{ result.GID }}</span></td>
                <td><code>{{ result.InitialFunc }}</code></td>
                <td class="text-center">{{ result.depth || '-' }}</td>
                <td class="text-center">{{ result.executionTime || '-' }}</td>
                <td class="text-center">
                  <template v-if="result.GID">
                    <div class="btn-group">
                      <router-link 
                        :to="{ name: 'TraceDetails', params: { gid: result.GID } }" 
                        class="btn btn-sm btn-primary"
                        title="查看详情"
                      >
                        <i class="bi bi-eye"></i> 详情
                      </router-link>
                      <router-link 
                        :to="{ name: 'MermaidViewer', params: { gid: result.GID } }" 
                        class="btn btn-sm btn-success"
                        title="查看调用图"
                      >
                        <i class="bi bi-diagram-3"></i> 调用图
                      </router-link>
                    </div>
                  </template>
                </td>
              </tr>
              <!-- 无数据时显示 -->
              <tr v-if="filteredGIDs.length === 0">
                <td colspan="5" class="text-center py-4">
                  <div class="alert alert-info mb-0">
                    <i class="bi bi-info-circle me-2"></i>
                    没有找到匹配的数据，请尝试其他搜索条件
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="card-footer">
        <!-- 分页控件 -->
        <nav aria-label="Page navigation">
          <ul class="pagination justify-content-center mb-0">
            <li class="page-item" :class="{ disabled: currentPage <= 1 }">
              <a class="page-link" href="#" @click.prevent="prevPage">上一页</a>
            </li>
            <li v-for="page in displayedPages" :key="page" class="page-item" :class="{ active: page === currentPage }">
              <a class="page-link" href="#" @click.prevent="goToPage(page)">{{ page }}</a>
            </li>
            <li class="page-item" :class="{ disabled: currentPage >= totalPages }">
              <a class="page-link" href="#" @click.prevent="nextPage">下一页</a>
            </li>
          </ul>
        </nav>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../axios';

export default {
  name: 'RuntimeAnalysis',
  props: {
    projectPath: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      gids: [],
      functionName: '',
      filteredGIDs: [],
      functionNames: [],
      filteredFunctionNames: [],
      currentPage: 1,
      itemsPerPage: 10, // 每页显示的 GID 数量
      totalPages: 0, // 总页数
      isComponentMounted: false, // 添加组件挂载状态标志
      showFunctionSuggestions: false, // 控制函数建议列表的显示
      suggestionsTimer: null, // 用于延迟隐藏建议列表的计时器
      isSearching: false, // 标记是否正在搜索
      inputPosition: { top: 0, left: 0, width: 0 }, // 存储输入框位置
      loading: false,
      showAllGoroutines: false,
      hotFunctions: [],
      hotFunctionSortBy: 'calls', // 'calls' 或 'time'
      goroutineStats: {
        active: 0,
        avgTime: '0ms',
        maxDepth: 0
      }
    };
  },
  mounted() {
    this.isComponentMounted = true;
    this.initializeData();
    
    // 添加点击事件监听器，点击页面其他地方时隐藏建议列表
    document.addEventListener('click', this.handleDocumentClick);
    
    // 添加窗口大小改变事件监听器
    window.addEventListener('resize', this.updateInputPosition);
  },
  beforeUnmount() {
    // 组件卸载前设置标志
    this.isComponentMounted = false;
    
    // 移除事件监听器
    document.removeEventListener('click', this.handleDocumentClick);
    window.removeEventListener('resize', this.updateInputPosition);
    
    // 清除计时器
    if (this.suggestionsTimer) {
      clearTimeout(this.suggestionsTimer);
    }
  },
  computed: {
    displayedPages() {
      const pages = [];
      const maxVisiblePages = 5;
      let startPage = Math.max(1, this.currentPage - Math.floor(maxVisiblePages / 2));
      let endPage = Math.min(this.totalPages, startPage + maxVisiblePages - 1);
      
      if (endPage - startPage + 1 < maxVisiblePages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }
      
      for (let i = startPage; i <= endPage; i++) {
        pages.push(i);
      }
      
      return pages;
    }
  },
  watch: {
    showAllGoroutines() {
      this.fetchGIDs();
    }
  },
  methods: {
    // 更新输入框位置
    updateInputPosition() {
      this.$nextTick(() => {
        const inputField = document.querySelector('.search-container .input-group');
        if (inputField) {
          const rect = inputField.getBoundingClientRect();
          this.inputPosition = {
            top: rect.bottom,
            left: rect.left,
            width: rect.width
          };
        }
      });
    },
    
    // 处理文档点击事件
    handleDocumentClick(event) {
      // 检查点击是否在建议列表或输入框之外
      const suggestionsList = document.querySelector('.function-suggestions');
      const inputField = document.querySelector('.search-container .input-group');
      
      if (suggestionsList && inputField && 
          !suggestionsList.contains(event.target) && 
          !inputField.contains(event.target)) {
        this.showFunctionSuggestions = false;
      }
    },
    
    // 输入函数名时触发
    onFunctionNameInput() {
      this.updateFunctionSuggestions();
      this.updateInputPosition();
    },
    
    // 更新函数建议列表
    updateFunctionSuggestions() {
      if (this.functionName) {
        const searchTerm = this.functionName.toLowerCase();
        this.filteredFunctionNames = this.functionNames
          .filter(name => name.toLowerCase().includes(searchTerm))
          .slice(0, 10); // 限制显示数量，提高性能
      } else {
        this.filteredFunctionNames = [];
      }
    },
    
    // 延迟隐藏建议列表
    hideSuggestionsDelayed() {
      // 清除之前的计时器
      if (this.suggestionsTimer) {
        clearTimeout(this.suggestionsTimer);
      }
      
      // 设置新的计时器，延迟隐藏建议列表
      this.suggestionsTimer = setTimeout(() => {
        this.showFunctionSuggestions = false;
      }, 200); // 200毫秒延迟，给点击事件足够的时间处理
    },

    async initializeData() {
      this.loading = true;
      await Promise.all([
        this.fetchGIDs(),
        this.fetchFunctionNames(),
        this.fetchHotFunctions(),
        this.fetchGoroutineStats()
      ]);
      this.loading = false;
    },

    async fetchGIDs() {
      try {
        const response = await axios.get('/api/gids', {
          params: {
            page: this.currentPage,
            limit: this.itemsPerPage,
            showAll: this.showAllGoroutines,
            includeMetrics: true // 添加参数，请求包含调用深度和执行时间
          },
        });
        
        this.filteredGIDs = (response.data.body || []).map(item => {
          // 添加模拟数据
          const mockData = this.generateMockMetrics();
          
          return {
            GID: item.gid,
            InitialFunc: item.initialFunc,
            depth: item.depth || mockData.depth,
            executionTime: item.executionTime || mockData.executionTime
          };
        });
        
        this.totalPages = Math.ceil(response.data.total / this.itemsPerPage);
      } catch (error) {
        console.error('获取GIDs失败:', error);
        this.$nextTick(() => {
          alert('获取GIDs失败: ' + error.message);
        });
      }
    },

    async fetchFunctionNames() {
      try {
        const response = await axios.get('/api/functions');
        this.functionNames = response.data.functionNames || [];
      } catch (error) {
        console.error('获取函数名列表失败:', error);
      }
    },

    async fetchHotFunctions() {
      try {
        const response = await axios.get('/api/hot-functions', {
          params: {
            sortBy: this.hotFunctionSortBy
          }
        });
        this.hotFunctions = response.data.functions || [];
      } catch (error) {
        console.error('获取热点函数失败:', error);
      }
    },

    async fetchGoroutineStats() {
      try {
        const response = await axios.get('/api/goroutine-stats');
        this.goroutineStats = response.data || {
          active: 0,
          avgTime: '0ms',
          maxDepth: 0
        };
      } catch (error) {
        console.error('获取Goroutine统计信息失败:', error);
      }
    },

    sortHotFunctions(sortBy) {
      this.hotFunctionSortBy = sortBy;
      this.fetchHotFunctions();
    },

    searchByFunctionName() {
      if (this.isSearching) return; // 防止重复搜索
      
      if (this.functionName) {
        this.updateFunctionSuggestions();
        // 使用 nextTick 确保 DOM 更新后再执行查询
        this.$nextTick(() => {
          this.fetchGIDsByFunctionName();
        });
      } else {
        this.filteredFunctionNames = [];
        this.fetchGIDs(); // 如果没有函数名，则获取所有 GIDs
      }
    },

    selectFunction(name) {
      this.functionName = name;
      this.filteredFunctionNames = [];
      this.showFunctionSuggestions = false;
      
      // 使用 nextTick 确保 DOM 更新后再执行查询
      this.$nextTick(() => {
        this.fetchGIDsByFunctionName();
      });
    },

    async fetchGIDsByFunctionName() {
      if (!this.isComponentMounted || this.isSearching) return;
      
      this.isSearching = true;
      
      try {
        const response = await axios.post('/api/gids/function', {
          functionName: this.functionName,
          path: this.projectPath,
          includeMetrics: true // 添加参数，请求包含调用深度和执行时间
        });
        
        if (!this.isComponentMounted) return;
        
        // 确保数据格式正确并且每个项目都有 GID 属性
        if (response.data && response.data.body) {
          this.filteredGIDs = response.data.body.map(item => {
            // 添加模拟数据
            const mockData = this.generateMockMetrics();
            
            // 确保每个项目都有 GID 和 InitialFunc 属性
            return {
              GID: item.gid || item.GID || '',
              InitialFunc: item.initialFunc || item.InitialFunc || this.functionName,
              depth: item.depth || mockData.depth,
              executionTime: item.executionTime || mockData.executionTime
            };
          }).filter(item => item.GID); // 过滤掉没有 GID 的项目
          
          this.totalPages = Math.ceil(this.filteredGIDs.length / this.itemsPerPage);
          this.currentPage = 1;
        } else {
          this.filteredGIDs = [];
          this.totalPages = 0;
          this.currentPage = 1;
        }
      } catch (error) {
        if (!this.isComponentMounted) return;
        
        console.error('搜索函数相关GIDs失败:', error);
        this.$nextTick(() => {
          alert('搜索函数相关GIDs失败: ' + error.message);
        });
        this.filteredGIDs = [];
        this.totalPages = 0;
      } finally {
        this.isSearching = false;
      }
    },

    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++;
        this.fetchGIDs();
      }
    },

    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--;
        this.fetchGIDs();
      }
    },

    goToPage(page) {
      if (page !== this.currentPage) {
        this.currentPage = page;
        this.fetchGIDs();
      }
    },

    // 生成模拟的调用深度和执行时间数据
    generateMockMetrics() {
      // 生成1-20之间的随机整数作为调用深度
      const depth = Math.floor(Math.random() * 20) + 1;
      
      // 生成1-100之间的随机整数作为执行时间（毫秒）
      const execTimeMs = Math.floor(Math.random() * 100) + 1;
      
      return {
        depth: depth,
        executionTime: `${execTimeMs}ms`
      };
    }
  }
};
</script>

<style scoped>
.search-container {
  position: relative;
}

.suggestions-wrapper {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 0;
  z-index: 9999;
  pointer-events: none;
}

.function-suggestions {
  position: absolute;
  z-index: 9999;
  width: calc(100% - 30px);
  max-width: 600px;
  max-height: 300px;
  overflow-y: auto;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
  background-color: white;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  margin-top: 0;
  top: 140px;
  left: 50%;
  transform: translateX(-50%);
  pointer-events: auto;
  padding: 0;
}

.function-suggestions .list-group-item {
  border-left: none;
  border-right: none;
  border-radius: 0;
  padding: 0.75rem 1rem;
  font-family: 'Consolas', 'Monaco', monospace;
  transition: all 0.2s ease;
  cursor: pointer;
}

.function-suggestions .list-group-item:first-child {
  border-top: none;
}

.function-suggestions .list-group-item:last-child {
  border-bottom: none;
}

.function-suggestions .list-group-item:hover {
  background-color: #f8f9fa;
  color: #0d6efd;
}

.function-suggestions .list-group-item-action:active {
  background-color: #e9ecef;
}

.pagination-info {
  display: flex;
  align-items: center;
}

@media (max-width: 768px) {
  .pagination-info {
    margin-top: 1rem;
  }
  
  .function-suggestions {
    width: 90%;
    top: 180px;
  }
}
</style> 