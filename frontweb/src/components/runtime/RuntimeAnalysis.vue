<template>
  <div class="runtime-analysis">
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
      <div class="col-md-3">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-cpu me-2"></i>{{ $t('runtimeAnalysis.statistics.activeGoroutines') }}</h5>
            <p class="display-4">{{ goroutineStats.active || 0 }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-hourglass-split me-2"></i>{{ $t('runtimeAnalysis.statistics.avgExecutionTime') }}</h5>
            <p class="display-4">{{ goroutineStats.avgTime || '0ms' }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-lightning-charge me-2"></i>{{ $t('runtimeAnalysis.statistics.maxCallDepth') }}</h5>
            <p class="display-4">{{ goroutineStats.maxDepth || 0 }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-exclamation-triangle me-2"></i>未完成函数</h5>
            <p class="display-4">{{ unfinishedFunctions.length || 0 }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 未完成函数列表 -->
    <div class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0"><i class="bi bi-exclamation-circle me-2"></i>未完成函数列表</h5>
        <div class="d-flex align-items-center">
          <div class="threshold-control me-3">
            <label class="form-label mb-0 me-2 fw-bold">阻塞阈值：</label>
            <div class="input-group">
              <input 
                type="number" 
                class="form-control" 
                v-model="blockingThreshold" 
                min="100"
                step="100"
                style="width: 100px;"
              >
              <span class="input-group-text">ms</span>
              <button class="btn btn-primary" @click="updateBlockingThreshold" title="应用阈值">
                <i class="bi bi-check-lg"></i> 应用
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="card-body">
        <div v-if="loadingUnfinishedFunctions" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">加载中...</span>
          </div>
          <p class="mt-3">正在加载未完成函数数据...</p>
        </div>
        <div v-else-if="unfinishedFunctions.length === 0" class="text-center py-5">
          <i class="bi bi-check-circle text-success display-4"></i>
          <p class="mt-3">没有检测到未完成函数</p>
        </div>
        <div v-else>
          <div class="alert alert-info mb-3">
            <i class="bi bi-info-circle me-2"></i>
            当函数运行时间超过 <strong>{{ blockingThreshold }}ms</strong> 时会被标记为阻塞状态
          </div>
          <div class="table-responsive">
            <table class="table table-hover">
              <thead class="table-light">
                <tr>
                  <th>函数名称</th>
                  <th class="text-center">GID</th>
                  <th class="text-center">已运行时间</th>
                  <th class="text-center">状态</th>
                  <th class="text-center">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(func, index) in unfinishedFunctions" :key="index" :class="{'table-warning': func.isBlocking}">
                  <td><code>{{ func.name }}</code></td>
                  <td class="text-center">
                    <span class="badge bg-primary">{{ func.gid }}</span>
                  </td>
                  <td class="text-center">{{ func.runningTime || '未知' }}</td>
                  <td class="text-center">
                    <span v-if="func.isBlocking" class="badge bg-danger">
                      <i class="bi bi-exclamation-triangle me-1"></i>阻塞
                    </span>
                    <span v-else class="badge bg-secondary">
                      <i class="bi bi-hourglass-split me-1"></i>运行中
                    </span>
                  </td>
                  <td class="text-center">
                    <button 
                      class="btn btn-sm btn-primary"
                      @click="viewFunctionCallChain(func.functionId, func.gid)"
                      title="查看调用链并高亮显示"
                    >
                      <i class="bi bi-eye me-1"></i>查看
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          
          <!-- 分页控件 -->
          <div class="d-flex justify-content-between align-items-center mt-3">
            <div>
              <span class="text-muted">共 {{ unfinishedFunctionsTotal }} 个未完成函数</span>
            </div>
            <nav aria-label="未完成函数分页">
              <ul class="pagination mb-0">
                <li class="page-item" :class="{ disabled: unfinishedFunctionsPage <= 1 }">
                  <a class="page-link" href="#" @click.prevent="prevUnfinishedFunctionsPage">
                    <i class="bi bi-chevron-left"></i>
                  </a>
                </li>
                <li v-for="page in displayedUnfinishedFunctionsPages" :key="page" 
                    class="page-item" :class="{ active: page === unfinishedFunctionsPage }">
                  <a class="page-link" href="#" @click.prevent="goToUnfinishedFunctionsPage(page)">{{ page }}</a>
                </li>
                <li class="page-item" :class="{ disabled: unfinishedFunctionsPage >= unfinishedFunctionsTotalPages }">
                  <a class="page-link" href="#" @click.prevent="nextUnfinishedFunctionsPage">
                    <i class="bi bi-chevron-right"></i>
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </div>
    </div>

    <!-- 热点函数分析 -->
    <div class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0"><i class="bi bi-fire me-2"></i>{{ $t('runtimeAnalysis.hotFunctions.title') }}</h5>
        <div class="btn-group">
          <button class="btn btn-sm btn-outline-primary" @click="sortHotFunctions('calls')" :class="{ active: hotFunctionSortBy === 'calls' }">
            {{ $t('runtimeAnalysis.hotFunctions.sortByCalls') }}
          </button>
          <button class="btn btn-sm btn-outline-primary" @click="sortHotFunctions('time')" :class="{ active: hotFunctionSortBy === 'time' }">
            {{ $t('runtimeAnalysis.hotFunctions.sortByTime') }}
          </button>
        </div>
      </div>
      <div class="card-body">
        <div v-if="loading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">{{ $t('runtimeAnalysis.hotFunctions.loading') }}</span>
          </div>
          <p class="mt-3">{{ $t('runtimeAnalysis.hotFunctions.loadingData') }}</p>
        </div>
        <div v-else-if="hotFunctions.length === 0" class="text-center py-5">
          <i class="bi bi-exclamation-circle text-warning display-4"></i>
          <p class="mt-3">{{ $t('runtimeAnalysis.hotFunctions.noData') }}</p>
        </div>
        <div v-else>
          <div class="table-responsive">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>{{ $t('runtimeAnalysis.hotFunctions.functionName') }}</th>
                  <th class="text-center">{{ $t('runtimeAnalysis.hotFunctions.callCount') }}</th>
                  <th class="text-center">{{ $t('runtimeAnalysis.hotFunctions.totalTime') }}</th>
                  <th class="text-center">{{ $t('runtimeAnalysis.hotFunctions.avgTime') }}</th>
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
        <h5 class="mb-0"><i class="bi bi-list-ul me-2"></i>{{ $t('runtimeAnalysis.goroutineList.title') }}</h5>
        <div class="pagination-info">
          <span class="badge bg-secondary">{{ $t('runtimeAnalysis.goroutineList.currentPage') }}: {{ currentPage }} / {{ totalPages }}</span>
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
                <th>{{ $t('runtimeAnalysis.goroutineList.gid') }}</th>
                <th>{{ $t('runtimeAnalysis.goroutineList.initialFunction') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.goroutineList.callDepth') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.goroutineList.executionTime') }}</th>
                <th class="text-center">状态</th>
                <th class="text-center">{{ $t('runtimeAnalysis.goroutineList.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="result in filteredGIDs" :key="result.GID">
                <td><span class="badge bg-primary">{{ result.GID }}</span></td>
                <td><code>{{ result.InitialFunc }}</code></td>
                <td class="text-center">{{ result.depth || '-' }}</td>
                <td class="text-center">{{ result.executionTime || '-' }}</td>
                <td class="text-center">
                  <span v-if="result.isFinished" class="badge bg-success">已完成</span>
                  <span v-else class="badge bg-warning">运行中</span>
                </td>
                <td class="text-center">
                  <template v-if="result.GID">
                    <div class="btn-group">
                      <router-link 
                        :to="{ name: 'TraceDetails', params: { gid: result.GID } }" 
                        class="btn btn-sm btn-primary"
                        :title="$t('runtimeAnalysis.goroutineList.details')"
                      >
                        <i class="bi bi-eye"></i> {{ $t('runtimeAnalysis.goroutineList.details') }}
                      </router-link>
                      <button 
                        class="btn btn-sm btn-success"
                        :title="$t('runtimeAnalysis.goroutineList.callGraph')"
                        @click="showFunctionCallGraph(result.GID)"
                      >
                        <i class="bi bi-graph-up"></i> {{ $t('runtimeAnalysis.goroutineList.callGraph') }}
                      </button>
                    </div>
                  </template>
                </td>
              </tr>
              <!-- 无数据时显示 -->
              <tr v-if="filteredGIDs.length === 0">
                <td colspan="6" class="text-center py-4">
                  <div class="alert alert-info mb-0">
                    <i class="bi bi-info-circle me-2"></i>
                    {{ $t('runtimeAnalysis.goroutineList.noData') }}
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
              <a class="page-link" href="#" @click.prevent="prevPage">{{ $t('runtimeAnalysis.goroutineList.prevPage') }}</a>
            </li>
            <li v-for="page in displayedPages" :key="page" class="page-item" :class="{ active: page === currentPage }">
              <a class="page-link" href="#" @click.prevent="goToPage(page)">{{ page }}</a>
            </li>
            <li class="page-item" :class="{ disabled: currentPage >= totalPages }">
              <a class="page-link" href="#" @click.prevent="nextPage">{{ $t('runtimeAnalysis.goroutineList.nextPage') }}</a>
            </li>
          </ul>
        </nav>
      </div>
    </div>

    <!-- 函数调用关系图组件 -->
    <function-call-graph
      v-if="showChart"
      :visible="showChart"
      :gid="currentGid"
      :dbpath="getCurrentDbPath()"
      @update:visible="showChart = $event"
      @error="handleChartError"
      :key="`chart-${currentGid}-${chartRenderCount}`"
      :use-mock-data="testMode"
    />
  
  </div>
</template>

<script>
import axios from '@/axios';
import FunctionCallGraph from '../charts/FunctionCallGraph.vue';
import { useI18n } from 'vue-i18n';

export default {
  name: 'RuntimeAnalysis',
  components: {
    FunctionCallGraph
  },
  props: {
    projectPath: {
      type: String,
      required: false,
      default: ''
    }
  },
  setup() {
    const { t, locale } = useI18n({ useScope: 'global' });
    return { t, locale };
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
      },
      showChart: false, // 是否显示图表
      currentGid: '', // 当前选中的GID
      showDatabaseError: false, // 添加数据库错误标志
      gidLimit: 10, // 添加gidLimit属性
      hotFunctionLimit: 10, // 添加hotFunctionLimit属性
      dbpath: '', // 当前使用的数据库路径
      chartRenderCount: 0, // 图表渲染计数器，用于强制重新创建组件
      testMode: false, // 测试模式，用于在API请求失败时使用模拟数据
      searchTimeout: null, // 用于防抖
      unfinishedFunctions: [], // 未完成函数列表
      unfinishedFunctionsTotal: 0, // 未完成函数总数
      unfinishedFunctionsPage: 1, // 当前页码
      unfinishedFunctionsLimit: 10, // 每页数量
      unfinishedFunctionsTotalPages: 1, // 总页数
      blockingThreshold: 1000, // 阻塞时间阈值（毫秒），默认1秒
      loadingUnfinishedFunctions: false, // 加载状态
      highlightedFunctionId: null, // 高亮显示的函数ID
    };
  },
  mounted() {
    this.isComponentMounted = true;
    
    // 从本地存储中恢复阻塞阈值设置
    const savedThreshold = localStorage.getItem('blockingThreshold');
    if (savedThreshold) {
      const threshold = parseInt(savedThreshold);
      if (!isNaN(threshold) && threshold >= 100) {
        this.blockingThreshold = threshold;
      }
    }
    
    this.initializeData();
    
    document.addEventListener('click', this.handleDocumentClick);
    window.addEventListener('resize', this.updateInputPosition);
    
    // 添加语言变化监听
    window.addEventListener('languageChanged', this.handleLanguageChange);
    
    // 添加路由变化监听，确保组件在重新激活时能正确加载数据
    this.$router.afterEach(() => {
      if (this.isComponentMounted) {
        this.initializeData();
      }
    });
  },
  beforeUnmount() {
    this.isComponentMounted = false;
    
    document.removeEventListener('click', this.handleDocumentClick);
    window.removeEventListener('resize', this.updateInputPosition);
    window.removeEventListener('languageChanged', this.handleLanguageChange);
    
    if (this.suggestionsTimer) {
      clearTimeout(this.suggestionsTimer);
    }
    
    // 清除定时器
    if (this.refreshUnfinishedFunctionsInterval) {
      clearInterval(this.refreshUnfinishedFunctionsInterval);
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
    },
    displayedUnfinishedFunctionsPages() {
      const pages = [];
      const maxVisiblePages = 5;
      let startPage = Math.max(1, this.unfinishedFunctionsPage - Math.floor(maxVisiblePages / 2));
      let endPage = Math.min(this.unfinishedFunctionsTotalPages, startPage + maxVisiblePages - 1);
      
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
    
    handleDocumentClick(event) {
      const suggestionsList = document.querySelector('.function-suggestions');
      const inputField = document.querySelector('.search-container .input-group');
      
      if (suggestionsList && inputField && 
          !suggestionsList.contains(event.target) && 
          !inputField.contains(event.target)) {
        this.showFunctionSuggestions = false;
      }
    },
    
    // 防抖函数
    debouncedUpdateSuggestions() {
      if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
      }
      
      this.searchTimeout = setTimeout(() => {
        this.updateFunctionSuggestions();
      }, 300); // 300毫秒延迟
    },
    
    // 输入函数名时触发
    onFunctionNameInput() {
      // 使用防抖处理，避免频繁触发搜索
      this.debouncedUpdateSuggestions();
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
        this.fetchGoroutineStats(),
        this.fetchUnfinishedFunctions()
      ]);
      this.loading = false;
    },

    async fetchGIDs() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.filteredGIDs = [];
          this.totalPages = 0;
          return;
        }
        
        const response = await axios.post('/api/runtime/gids', {
          page: this.currentPage,
          limit: this.itemsPerPage,
          includeMetrics: true,
          dbpath: dbpath
        });
        
        this.filteredGIDs = (response.data.body || []).map(item => {
          const mockData = this.generateMockMetrics();
          
          return {
            GID: item.gid,
            InitialFunc: item.initialFunc,
            depth: item.depth || mockData.depth,
            executionTime: item.executionTime || mockData.executionTime,
            isFinished: item.isFinished || this.isExecutionFinished(item.executionTime)
          };
        });
        
        this.totalPages = Math.ceil(response.data.total / this.itemsPerPage);
      } catch (error) {
        console.error('获取GIDs失败:', error);
        // 不显示弹窗，只在控制台输出错误
        this.filteredGIDs = [];
        this.totalPages = 0;
      }
    },

    async fetchFunctionNames() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.functionNames = [];
          return;
        }

        const response = await axios.post('/api/runtime/functions', {
          dbpath: dbpath
        });
        this.functionNames = response.data.functionNames || [];
      } catch (error) {
        console.error('获取函数名列表失败:', error);
        this.functionNames = [];
      }
    },

    async fetchHotFunctions() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.hotFunctions = [];
          return;
        }

        const response = await axios.post('/api/runtime/hot-functions', {
          sortBy: this.hotFunctionSortBy,
          dbpath: dbpath
        });
        this.hotFunctions = response.data.functions || [];
      } catch (error) {
        console.error('获取热点函数失败:', error);
        this.hotFunctions = [];
      }
    },

    async fetchGoroutineStats() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.goroutineStats = {
            active: 0,
            avgTime: '0ms',
            maxDepth: 0
          };
          return;
        }

        const response = await axios.post('/api/runtime/goroutine-stats', {
          dbpath: dbpath
        });
        this.goroutineStats = response.data || {
          active: 0,
          avgTime: '0ms',
          maxDepth: 0
        };
      } catch (error) {
        console.error('获取Goroutine统计信息失败:', error);
        this.goroutineStats = {
          active: 0,
          avgTime: '0ms',
          maxDepth: 0
        };
      }
    },

    sortHotFunctions(sortBy) {
      this.hotFunctionSortBy = sortBy;
      this.fetchHotFunctions();
    },

    searchByFunctionName() {
      if (this.isSearching) return;
      
      if (this.functionName) {
        this.updateFunctionSuggestions();
        this.$nextTick(() => {
          this.fetchGIDsByFunctionName();
        });
      } else {
        this.filteredFunctionNames = [];
        this.fetchGIDs();
      }
    },

    selectFunction(name) {
      this.functionName = name;
      this.filteredFunctionNames = [];
      this.showFunctionSuggestions = false;
      
      // 选择函数后自动执行搜索
      this.searchByFunctionName();
      
      // 保持输入框焦点
      this.$nextTick(() => {
        if (this.$refs.searchInput) {
          this.$refs.searchInput.focus();
        }
      });
    },

    async fetchGIDsByFunctionName() {
      if (!this.isComponentMounted || this.isSearching) return;
      
      this.isSearching = true;
      
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.filteredGIDs = [];
          this.totalPages = 0;
          this.currentPage = 1;
          return;
        }
        
        const response = await axios.post('/api/runtime/gids/function', {
          functionName: this.functionName,
          path: dbpath,
          includeMetrics: true
        });
        
        if (!this.isComponentMounted) return;
        
        if (response.data && response.data.body) {
          this.filteredGIDs = response.data.body.map(item => {
            const mockData = this.generateMockMetrics();
            
            return {
              GID: item.gid || item.GID || '',
              InitialFunc: item.initialFunc || item.InitialFunc || this.functionName,
              depth: item.depth || mockData.depth,
              executionTime: item.executionTime || mockData.executionTime
            };
          }).filter(item => item.GID);
          
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

    generateMockMetrics() {
      const depth = Math.floor(Math.random() * 20) + 1;
      
      const execTimeMs = Math.floor(Math.random() * 100) + 1;
      
      return {
        depth: depth,
        executionTime: `${execTimeMs}ms`
      };
    },

    async showFunctionCallGraph(gid) {
      console.log(`显示函数调用图，GID: ${gid}`);
      this.currentGid = gid;
      
      // 确保数据库路径已设置
      if (!this.dbpath) {
        // 尝试从项目路径获取
        if (this.projectPathInput) {
          this.dbpath = this.projectPathInput;
          console.log('从项目路径设置数据库路径:', this.dbpath);
        } else {
          console.error('数据库路径未设置，无法获取调用图');
          alert('请先在上方输入Go项目路径');
          return;
        }
      }
      
      console.log(`使用数据库路径: ${this.dbpath}`);
      
      // 增加渲染计数，确保每次都是新实例
      this.chartRenderCount = (this.chartRenderCount || 0) + 1;
      
      // 切换显示状态，强制组件重新渲染
      this.showChart = false;
      await this.$nextTick();
      
      // 直接测试API连接
      try {
        await this.testGraphApi(gid);
      } catch (error) {
        console.error('测试图表API失败:', error);
      }
      
      // 延迟显示图表，确保DOM已准备好
      setTimeout(() => {
        this.showChart = true;
        console.log('函数调用图组件已重新渲染，渲染计数:', this.chartRenderCount);
      }, 500);
    },
    
    // 直接测试图表API
    async testGraphApi(gid) {
      try {
        console.log(`直接测试图表API，GID: ${gid}, 数据库路径: ${this.dbpath}`);
        
        // 构建请求URL和参数
        const url = `/api/runtime/traces/graph`;
        const params = { 
          gid: gid,
          dbpath: this.dbpath 
        };
        
        console.log('直接发送API请求:', { url, params });
        
        // 使用axios发送请求
        const response = await axios.post(url, params);
        console.log('API响应状态:', response.status);
        
        if (response.status === 200) {
          const data = response.data;
          console.log('API响应数据:', data);
          
          if (!data.nodes || !data.edges) {
            console.error('API返回的数据格式不正确 - 缺少nodes或edges字段:', data);
          } else if (data.nodes.length === 0) {
            console.warn('API返回的节点数据为空数组');
          } else {
            console.log(`API返回了${data.nodes.length}个节点和${data.edges.length}条边`);
          }
        } else {
          console.error(`API请求失败，状态码: ${response.status}`);
        }
      } catch (error) {
        console.error('测试图表API失败:', error);
      }
    },
    
    // 获取函数调用关系图
    async getFunctionCallGraph(dbpath, functionName, depth, direction) {
      try {
        const response = await axios({
          url: '/api/runtime/function/graph',
          method: 'post',
          data: {
            dbpath,
            functionName,
            depth,
            direction
          }
        });
        return response.data;
      } catch (error) {
        console.error('获取函数调用关系图失败:', error);
        throw error;
      }
    },
    
    handleChartError(errorMessage) {
      console.error('图表错误:', errorMessage);
      this.$message.error(errorMessage);
    },

    // 获取当前数据库路径
    getCurrentDbPath() {
      console.log('获取数据库路径，当前状态:', {
        dbpath: this.dbpath,
        projectPath: this.projectPath
      });
      
      // 如果已经设置了数据库路径，直接返回
      if (this.dbpath) {
        console.log('使用已设置的数据库路径:', this.dbpath);
        return this.dbpath;
      }
      
      // 否则使用项目路径作为数据库路径
      if (this.projectPath) {
        this.dbpath = this.projectPath;
        console.log('使用项目路径作为数据库路径:', this.dbpath);
        return this.dbpath;
      }
      
      // 如果都没有，返回空字符串
      console.warn('数据库路径为空');
      return '';
    },


    // 处理语言变化
    handleLanguageChange(event) {
      console.log('RuntimeAnalysis - Language changed:', event.detail.locale);
      // 强制刷新组件中的国际化文本
      this.$forceUpdate();
    },

    // 判断执行是否已完成
    isExecutionFinished(executionTime) {
      if (!executionTime) return false;
      
      // 如果执行时间字符串包含具体时间值，则认为已完成
      return executionTime.match(/\d+(\.\d+)?(ms|s|µs)/i) !== null;
    },
    
    // 获取未完成函数列表
    async fetchUnfinishedFunctions() {
      const dbpath = this.getCurrentDbPath();
      if (!dbpath) {
        this.unfinishedFunctions = [];
        this.unfinishedFunctionsTotal = 0;
        this.unfinishedFunctionsTotalPages = 1;
        return;
      }
      
      try {
        this.loadingUnfinishedFunctions = true;
        
        // 确保阻塞阈值是有效的数字
        const threshold = parseInt(this.blockingThreshold);
        if (isNaN(threshold) || threshold < 100) {
          this.blockingThreshold = 1000; // 如果无效，设置为默认值
        }
        
        console.log(`获取未完成函数，阻塞阈值: ${this.blockingThreshold}ms`);
        
        const response = await axios.post('/api/runtime/unfinished-functions', {
          dbpath: dbpath,
          threshold: parseInt(this.blockingThreshold), // 确保发送数字类型
          page: this.unfinishedFunctionsPage,
          limit: this.unfinishedFunctionsLimit
        });
        
        this.unfinishedFunctions = response.data.functions || [];
        this.unfinishedFunctionsTotal = response.data.total || 0;
        this.unfinishedFunctionsTotalPages = Math.ceil(this.unfinishedFunctionsTotal / this.unfinishedFunctionsLimit) || 1;
        
        // 如果当前页码超出总页数，则回到第一页
        if (this.unfinishedFunctionsPage > this.unfinishedFunctionsTotalPages && this.unfinishedFunctionsTotalPages > 0) {
          this.unfinishedFunctionsPage = 1;
          await this.fetchUnfinishedFunctions();
          return;
        }
        
        console.log('未完成函数列表:', this.unfinishedFunctions);
        console.log('未完成函数总数:', this.unfinishedFunctionsTotal);
        console.log('未完成函数总页数:', this.unfinishedFunctionsTotalPages);
      } catch (error) {
        console.error('获取未完成函数列表失败:', error);
        this.$message?.error?.('获取未完成函数列表失败') || alert('获取未完成函数列表失败');
        this.unfinishedFunctions = [];
        this.unfinishedFunctionsTotal = 0;
        this.unfinishedFunctionsTotalPages = 1;
      } finally {
        this.loadingUnfinishedFunctions = false;
      }
    },
    
    // 更新阻塞阈值并刷新未完成函数列表
    updateBlockingThreshold() {
      // 验证输入是否为有效数字
      const threshold = parseInt(this.blockingThreshold);
      if (isNaN(threshold) || threshold < 0) {
        this.$message?.warning?.('无效的阻塞阈值') || alert('无效的阻塞阈值');
        this.blockingThreshold = 1000; // 重置为默认值
        return;
      }
      
      // 设置最小值
      if (threshold < 100) {
        this.blockingThreshold = 100;
      }
      
      // 保存到本地存储
      localStorage.setItem('blockingThreshold', this.blockingThreshold.toString());
      
      // 显示提示信息
      this.$message?.success?.(`阻塞阈值已更新为 ${this.blockingThreshold}ms`) || 
        alert(`阻塞阈值已更新为 ${this.blockingThreshold}ms`);
      
      // 重置页码并刷新未完成函数列表
      this.unfinishedFunctionsPage = 1;
      this.fetchUnfinishedFunctions();
    },
    
    // 前往上一页
    prevUnfinishedFunctionsPage() {
      if (this.unfinishedFunctionsPage > 1) {
        this.unfinishedFunctionsPage--;
        this.fetchUnfinishedFunctions();
      }
    },
    
    // 前往下一页
    nextUnfinishedFunctionsPage() {
      if (this.unfinishedFunctionsPage < this.unfinishedFunctionsTotalPages) {
        this.unfinishedFunctionsPage++;
        this.fetchUnfinishedFunctions();
      }
    },
    
    // 前往指定页
    goToUnfinishedFunctionsPage(page) {
      if (page !== this.unfinishedFunctionsPage) {
        this.unfinishedFunctionsPage = page;
        this.fetchUnfinishedFunctions();
      }
    },
    
    // 查看函数调用链
    async viewFunctionCallChain(functionId, gid) {
      // 保存高亮函数ID到本地存储
      localStorage.setItem('highlightedFunctionId', functionId);
      
      // 跳转到调用链页面
      this.$router.push({
        name: 'TraceDetails',
        params: { gid: gid }
      });
    },
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

/* 阻塞阈值控件样式 */
.threshold-control {
  display: flex;
  align-items: center;
  background-color: #f8f9fa;
  padding: 8px 12px;
  border-radius: 6px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.threshold-control .form-label {
  color: #495057;
  white-space: nowrap;
}

.threshold-control .input-group {
  width: auto;
  flex-wrap: nowrap;
}

.threshold-control .form-control {
  text-align: center;
  font-weight: 500;
}

.threshold-control .btn-primary {
  min-width: 80px;
}

@media (max-width: 768px) {
  .pagination-info {
    margin-top: 1rem;
  }
  
  .function-suggestions {
    width: 90%;
    top: 180px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start !important;
  }
  
  .card-header .d-flex {
    margin-top: 10px;
    width: 100%;
  }
  
  .threshold-control {
    width: 100%;
    margin-bottom: 10px;
    flex-direction: column;
    align-items: flex-start;
  }
  
  .threshold-control .form-label {
    margin-bottom: 5px !important;
  }
  
  .threshold-control .input-group {
    width: 100%;
  }
}
</style> 