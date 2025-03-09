<template>
  <div class="function-analysis">
    <!-- 函数搜索区域 -->
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-search me-2"></i>{{ $t('runtimeAnalysis.functionAnalysis.title') }}</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-8 search-container">
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-code-square"></i></span>
              <input
                v-model="functionName"
                type="text"
                :placeholder="$t('runtimeAnalysis.functionAnalysis.inputFunctionName')"
                class="form-control"
                @input="onFunctionNameInput"
                @focus="showFunctionSuggestions = true"
                @blur="hideSuggestionsDelayed"
                ref="searchInput"
              />
              <button class="btn btn-primary" @click="searchFunction">
                <i class="bi bi-search me-1"></i> {{ $t('runtimeAnalysis.functionAnalysis.search') }}
              </button>
            </div>
            <small class="text-muted mt-2 d-block">
              {{ $t('runtimeAnalysis.functionAnalysis.inputFunctionName') }}
            </small>
            
            <!-- 使用teleport将下拉框移到body层级 -->
            <teleport to="body">
              <!-- 函数名建议列表 -->
              <div class="suggestions-container" v-if="showFunctionSuggestions && filteredFunctionNames.length" :style="suggestionsStyle">
                <div class="suggestions-wrapper">
                  <ul class="list-group function-suggestions" ref="suggestionsList">
                    <li
                      v-for="(item, index) in filteredFunctionNames"
                      :key="index"
                      class="list-group-item list-group-item-action"
                      :class="{ 'active': index === selectedSuggestionIndex }"
                      @mousedown.prevent="selectFunction(item.name)"
                      @mouseover="selectedSuggestionIndex = index"
                    >
                      <span v-html="highlightMatch(item.name, functionName)"></span>
                      <small v-if="item.package" class="text-muted d-block">{{ item.package }}</small>
                    </li>
                  </ul>
                </div>
              </div>
            </teleport>
          </div>
          <div class="col-md-4">
            <div class="form-group">
              <div class="btn-group w-100">
                <button 
                  class="btn btn-primary"
                  disabled
                >
                  <i class="bi bi-arrow-down-square me-1"></i> {{ $t('runtimeAnalysis.functionAnalysis.caller') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 函数调用链分析结果 -->
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">{{ $t('runtimeAnalysis.functionAnalysis.loading') }}</span>
      </div>
      <p class="mt-3">{{ $t('runtimeAnalysis.functionAnalysis.analyzing') }}</p>
    </div>

    <div v-else-if="functionName && !loading && !functionCallData.length" class="card mb-4">
      <div class="card-body text-center py-5">
        <i class="bi bi-search text-muted display-1"></i>
        <h4 class="mt-3">{{ $t('runtimeAnalysis.functionAnalysis.noRelatedFunction') }}</h4>
        <p>{{ $t('runtimeAnalysis.functionAnalysis.tryOtherFunctionName') }}</p>
      </div>
    </div>

    <div v-else-if="functionCallData.length > 0" class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="bi bi-arrow-down-square"></i>
          {{ $t('runtimeAnalysis.functionAnalysis.callRelationAnalysis') }}: <code>{{ functionName }}</code>
        </h5>
        <div class="btn-group">
          <button class="btn btn-sm btn-outline-primary" @click="exportToCSV">
            <i class="bi bi-download me-1"></i> {{ $t('runtimeAnalysis.functionAnalysis.export') }}
          </button>
        </div>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover table-striped mb-0">
            <thead class="table-light">
              <tr>
                <th>{{ $t('runtimeAnalysis.functionAnalysis.functionName') }}</th>
                <th>{{ $t('runtimeAnalysis.functionAnalysis.packagePath') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.functionAnalysis.callLevel') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.functionAnalysis.callCount') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.functionAnalysis.avgTime') }}</th>
                <th class="text-center">{{ $t('runtimeAnalysis.functionAnalysis.operation') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in flattenedCallData" :key="item.id" :class="{'table-primary': item.name === functionName}">
                <td>
                  <div class="d-flex align-items-center">
                    <span v-if="item.level > 0" :style="{ marginLeft: (item.level * 20) + 'px' }">
                      <i class="bi bi-arrow-return-right me-2 text-muted"></i>
                    </span>
                    <code class="function-name">{{ item.name }}</code>
                  </div>
                </td>
                <td><small class="text-muted">{{ item.package }}</small></td>
                <td class="text-center">{{ item.level }}</td>
                <td class="text-center">
                  <span class="badge bg-primary">{{ item.callCount || 0 }}</span>
                </td>
                <td class="text-center">{{ item.avgTime || 'N/A' }}</td>
                <td class="text-center">
                  <button v-if="item.name !== functionName" 
                          class="btn btn-sm btn-outline-info"
                          @click="searchRelatedFunction(item.name)">
                    <i class="bi bi-search me-1"></i> {{ $t('runtimeAnalysis.functionAnalysis.search') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../../axios';
import { nextTick } from 'vue';
import { useI18n } from 'vue-i18n';

export default {
  name: 'FunctionAnalysis',
  props: {
    projectPath: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      loading: false,
      functionName: '',
      functionNames: [],
      filteredFunctionNames: [],
      showFunctionSuggestions: false,
      suggestionsTimer: null,
      inputPosition: { top: 0, left: 0, width: 0 },
      functionCallData: [],
      collapsedNodes: new Set(),
      dbpath: '', // 当前使用的数据库路径
      searchTimeout: null, // 用于防抖
      selectedSuggestionIndex: 0, // 当前选中的建议索引
      maxSuggestions: 20, // 增加最大显示建议数量
      suggestionsStyle: {
        position: 'fixed',
        top: '0px',
        left: '0px',
        width: '300px',
        zIndex: 9999
      }
    };
  },
  setup() {
    const { t, locale } = useI18n({ useScope: 'global' });
    return { t, locale };
  },
  computed: {
    processedCallData() {
      if (!this.functionCallData.length) return [];
      
      const result = [];
      
      // 递归处理调用树
      const processNode = (node, level, parentId) => {
        // 添加当前节点
        const processedNode = {
          ...node,
          level,
          parentId
        };
        
        result.push(processedNode);
        
        // 如果节点被折叠，不处理其子节点
        if (this.isCollapsed(node.id)) return;
        
        // 处理子节点
        if (node.children && node.children.length) {
          node.children.forEach(child => {
            processNode(child, level + 1, node.id);
          });
        }
      };
      
      // 处理根节点
      this.functionCallData.forEach(node => {
        processNode(node, 0, null);
      });
      
      return result;
    },
    flattenedCallData() {
      // 将处理后的调用数据转换为扁平结构，适合表格展示
      return this.processedCallData.map(item => ({
        id: item.id,
        name: item.name,
        package: item.package || '',
        level: item.level,
        callCount: item.callCount || 0,
        avgTime: item.avgTime || 'N/A',
        parentId: item.parentId
      }));
    }
  },
  mounted() {
    this.isComponentMounted = true;
    this.fetchFunctionNames();
    
    // 添加点击事件监听器，点击页面其他地方时隐藏建议列表
    document.addEventListener('click', this.handleDocumentClick);
    
    // 添加键盘事件监听器，用于导航建议列表
    document.addEventListener('keydown', this.handleKeyDown);
    
    // 监听窗口大小变化，更新提示框位置
    window.addEventListener('resize', this.handleResize);

    window.addEventListener('languageChanged', this.handleLanguageChange);
    
    // 监听滚动事件，确保下拉框位置正确
    window.addEventListener('scroll', this.handleScroll, true);
  },
  beforeUnmount() {
    this.isComponentMounted = false;
    
    // 移除事件监听器
    document.removeEventListener('click', this.handleDocumentClick);
    document.removeEventListener('keydown', this.handleKeyDown);
    window.removeEventListener('resize', this.handleResize);
    window.removeEventListener('languageChanged', this.handleLanguageChange);
    window.removeEventListener('scroll', this.handleScroll, true);
    
    // 清除计时器
    if (this.suggestionsTimer) {
      clearTimeout(this.suggestionsTimer);
    }
  },
  methods: {
    
    // 处理文档点击事件
    handleDocumentClick(event) {
      // 检查点击是否在建议列表或输入框之外
      const suggestionsList = document.querySelector('.function-suggestions');
      const inputField = this.$refs.searchInput;
      
      // 如果点击的是下拉框或输入框，不隐藏下拉框
      if (!this.showFunctionSuggestions || 
          !suggestionsList || 
          suggestionsList.contains(event.target) || 
          event.target === inputField || 
          inputField.contains(event.target)) {
        return;
      }
      
      // 否则隐藏下拉框
      this.showFunctionSuggestions = false;
    },
    
    // 输入函数名时触发
    onFunctionNameInput() {
      // 使用防抖处理，避免频繁触发搜索
      this.debouncedUpdateSuggestions();
      
      // 更新提示框位置
      this.updateSuggestionsPosition();
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
    
    // 处理键盘事件
    handleKeyDown(event) {
      // 只有当建议列表显示时才处理键盘事件
      if (!this.showFunctionSuggestions || !this.filteredFunctionNames.length) {
        return;
      }
      
      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault();
          this.selectedSuggestionIndex = (this.selectedSuggestionIndex + 1) % this.filteredFunctionNames.length;
          this.scrollToSelectedSuggestion();
          break;
        case 'ArrowUp':
          event.preventDefault();
          this.selectedSuggestionIndex = (this.selectedSuggestionIndex - 1 + this.filteredFunctionNames.length) % this.filteredFunctionNames.length;
          this.scrollToSelectedSuggestion();
          break;
        case 'Enter':
          if (this.filteredFunctionNames.length > 0) {
            event.preventDefault();
            this.selectFunction(this.filteredFunctionNames[this.selectedSuggestionIndex].name);
          }
          break;
        case 'Escape':
          event.preventDefault();
          this.showFunctionSuggestions = false;
          break;
        case 'Tab':
          if (this.filteredFunctionNames.length > 0) {
            event.preventDefault();
            this.selectFunction(this.filteredFunctionNames[this.selectedSuggestionIndex].name);
          }
          break;
      }
    },
    
    // 滚动到选中的建议
    scrollToSelectedSuggestion() {
      this.$nextTick(() => {
        if (this.$refs.suggestionsList) {
          const selectedItem = this.$refs.suggestionsList.children[this.selectedSuggestionIndex];
          if (selectedItem) {
            selectedItem.scrollIntoView({ block: 'nearest' });
          }
        }
      });
    },
    
    // 高亮匹配文本
    highlightMatch(text, query) {
      if (!query) return text;
      
      try {
        const lowerText = text.toLowerCase();
        const lowerQuery = query.toLowerCase();
        
        if (lowerText.includes(lowerQuery)) {
          const startIndex = lowerText.indexOf(lowerQuery);
          const endIndex = startIndex + lowerQuery.length;
          
          return (
            text.substring(0, startIndex) +
            '<span class="highlight">' +
            text.substring(startIndex, endIndex) +
            '</span>' +
            text.substring(endIndex)
          );
        }
      } catch (e) {
        console.error('高亮匹配文本失败:', e);
      }
      
      return text;
    },
    
    // 更新函数建议列表
    updateFunctionSuggestions() {
      this.selectedSuggestionIndex = 0;
      
      if (!this.functionName) {
        this.filteredFunctionNames = [];
        return;
      }
      
      const searchTerm = this.functionName.toLowerCase();
      
      // 对函数名进行评分和排序
      const scoredResults = this.functionNames
        .map(func => {
          const lowerName = func.name.toLowerCase();
          let score = 0;
          
          // 完全匹配得分最高
          if (lowerName === searchTerm) {
            score = 100;
          } 
          // 前缀匹配得分次之
          else if (lowerName.startsWith(searchTerm)) {
            score = 80;
          }
          // 包含匹配得分再次
          else if (lowerName.includes(searchTerm)) {
            score = 60;
          }
          // 单词边界匹配
          else if (lowerName.includes('_' + searchTerm) || lowerName.includes('.' + searchTerm)) {
            score = 40;
          }
          // 不匹配
          else {
            score = 0;
          }
          
          return {
            name: func.name,
            package: func.package || '',
            score: score
          };
        })
        .filter(item => item.score > 0)
        .sort((a, b) => {
          // 首先按分数排序
          if (b.score !== a.score) {
            return b.score - a.score;
          }
          // 然后按名称长度排序（较短的优先）
          return a.name.length - b.name.length;
        })
        .slice(0, this.maxSuggestions);
      
      this.filteredFunctionNames = scoredResults;
      
      // 确保下拉框在下一个渲染周期中正确显示
      this.$nextTick(() => {
        this.updateSuggestionsPosition();
      });
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
    
    // 选择函数
    selectFunction(name) {
      this.functionName = name;
      this.filteredFunctionNames = [];
      this.showFunctionSuggestions = false;
      
      // 选择函数后自动执行搜索
      this.searchFunction();
      
      // 保持输入框焦点
      this.$nextTick(() => {
        if (this.$refs.searchInput) {
          this.$refs.searchInput.focus();
        }
      });
    },
    
    // 查询函数
    async searchFunction() {
      if (!this.functionName) {
        alert('请输入函数名');
        return;
      }
      
      this.loading = true;
      this.functionCallData = [];
      this.collapsedNodes.clear();
      
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          return;
        }
        
        const response = await axios.post('/api/runtime/function/analysis', {
          functionName: this.functionName,
          type: 'caller', // 固定为caller
          path: dbpath
        });
        
        if (response.data && response.data.callData) {
          this.functionCallData = this.processCallData(response.data.callData);
        }
      } catch (error) {
        console.error('获取函数调用关系失败:', error);
        this.$nextTick(() => {
          alert('获取函数调用关系失败: ' + error.message);
        });
      } finally {
        this.loading = false;
      }
    },
    
    // 处理调用数据，添加唯一ID
    processCallData(data) {
      let nextId = 1;
      
      const addIds = (node) => {
        node.id = nextId++;
        
        if (node.children && node.children.length) {
          node.children.forEach(child => {
            addIds(child);
          });
        }
        
        return node;
      };
      
      return data.map(node => addIds({...node}));
    },
    
    // 查询相关函数
    searchRelatedFunction(name) {
      // 重置选中索引
      this.selectedSuggestionIndex = 0;
      this.functionName = name;
      this.searchFunction();
    },
    
    // 获取函数名列表
    async fetchFunctionNames() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          this.functionNames = [];
          return;
        }
        
        const response = await axios.post('/api/runtime/functions', {
          dbpath: dbpath,
          includePackage: true // 请求包含包路径信息
        });
        
        if (response.data && response.data.functions) {
          // 如果API返回了包含包路径的函数信息
          this.functionNames = response.data.functions.map(func => ({
            name: func.name,
            package: func.package || ''
          }));
        } else {
          // 兼容旧版API，只返回函数名列表
          this.functionNames = (response.data.functionNames || []).map(name => ({
            name: name,
            package: ''
          }));
        }
      } catch (error) {
        console.error('获取函数名列表失败:', error);
        this.functionNames = [];
      }
    },
    
    // 获取当前数据库路径
    getCurrentDbPath() {
      // 如果已经设置了数据库路径，直接返回
      if (this.dbpath) {
        return this.dbpath;
      }
      
      // 否则使用项目路径
      if (this.projectPath) {
        this.dbpath = this.projectPath;
        return this.dbpath;
      }
      
      // 如果都没有，返回空字符串
      return '';
    },
    
    // 检查节点是否有子节点
    hasChildren(node) {
      const originalNode = this.findOriginalNode(node.id);
      return originalNode && originalNode.children && originalNode.children.length > 0;
    },
    
    // 查找原始节点
    findOriginalNode(id) {
      const findNode = (nodes) => {
        for (const node of nodes) {
          if (node.id === id) return node;
          
          if (node.children && node.children.length) {
            const found = findNode(node.children);
            if (found) return found;
          }
        }
        return null;
      };
      
      return findNode(this.functionCallData);
    },
    
    // 切换节点的折叠状态
    toggleNode(id) {
      if (this.collapsedNodes.has(id)) {
        this.collapsedNodes.delete(id);
      } else {
        this.collapsedNodes.add(id);
      }
    },
    
    // 检查节点是否已折叠
    isCollapsed(id) {
      return this.collapsedNodes.has(id);
    },
    
    // 展开所有节点
    expandAll() {
      this.collapsedNodes.clear();
    },
    
    // 折叠所有节点
    collapseAll() {
      // 找出所有有子节点的节点ID
      const findNodesWithChildren = (nodes) => {
        let result = [];
        
        for (const node of nodes) {
          if (node.children && node.children.length) {
            result.push(node.id);
            result = result.concat(findNodesWithChildren(node.children));
          }
        }
        
        return result;
      };
      
      const nodeIds = findNodesWithChildren(this.functionCallData);
      this.collapsedNodes = new Set(nodeIds);
    },
    
    // 导出为CSV
    exportToCSV() {
      if (!this.flattenedCallData.length) {
        alert('没有数据可导出');
        return;
      }
      
      // 定义CSV头部
      const headers = ['函数名称', '包路径', '调用层级', '调用次数', '平均耗时'];
      
      // 将数据转换为CSV格式
      const csvData = this.flattenedCallData.map(item => [
        item.name,
        item.package,
        item.level,
        item.callCount,
        item.avgTime
      ]);
      
      // 添加头部
      csvData.unshift(headers);
      
      // 将数组转换为CSV字符串
      const csvString = csvData.map(row => row.map(cell => {
        // 如果单元格包含逗号、双引号或换行符，则用双引号包裹并转义双引号
        if (cell && (String(cell).includes(',') || String(cell).includes('"') || String(cell).includes('\n'))) {
          return `"${String(cell).replace(/"/g, '""')}"`;
        }
        return String(cell);
      }).join(',')).join('\n');
      
      // 创建Blob对象
      const blob = new Blob([csvString], { type: 'text/csv;charset=utf-8;' });
      
      // 创建下载链接
      const link = document.createElement('a');
      const url = URL.createObjectURL(blob);
      
      // 设置下载属性
      link.setAttribute('href', url);
      link.setAttribute('download', `函数调用关系_${this.functionName}_${new Date().toISOString().slice(0, 10)}.csv`);
      link.style.visibility = 'hidden';
      
      // 添加到文档并触发点击
      document.body.appendChild(link);
      link.click();
      
      // 清理
      document.body.removeChild(link);
    },
    
    // 更新提示框位置
    updateSuggestionsPosition() {
      nextTick(() => {
        if (this.$refs.searchInput) {
          // 获取输入框的位置和尺寸
          const inputEl = this.$refs.searchInput;
          const rect = inputEl.getBoundingClientRect();
          
          // 更新下拉框的位置和宽度
          this.suggestionsStyle = {
            position: 'fixed',
            top: `${rect.bottom}px`,
            left: `${rect.left}px`,
            width: `${rect.width}px`,
            zIndex: 9999,
            maxHeight: `${window.innerHeight - rect.bottom - 20}px` // 确保不超出视口底部
          };
        }
      });
    },
    
    // 处理窗口大小变化
    handleResize() {
      if (this.showFunctionSuggestions && this.filteredFunctionNames.length) {
        this.updateSuggestionsPosition();
      }
    },
    
    // 处理滚动事件
    handleScroll() {
      if (this.showFunctionSuggestions && this.filteredFunctionNames.length) {
        this.updateSuggestionsPosition();
      }
    },
  }
};
</script>

<style scoped>
.search-container {
  position: relative;
  z-index: 1050; /* 确保搜索容器有较高的z-index */
}

/* 全局样式，确保teleport后的元素样式正确 */
.suggestions-container {
  position: fixed;
  z-index: 9999 !important; /* 确保在最顶层 */
}

.suggestions-wrapper {
  position: relative;
  width: 100%;
  max-width: 100%;
  overflow: visible;
}

.function-suggestions {
  width: 100%;
  max-height: 300px;
  overflow-y: auto;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
  background-color: white;
  border: 1px solid #dee2e6;
  border-radius: 0 0 0.375rem 0.375rem;
  margin-top: 0;
  padding: 0;
  position: relative; /* 确保定位上下文正确 */
  display: block !important; /* 确保始终显示 */
}

.function-suggestions .list-group-item {
  border-left: none;
  border-right: none;
  border-radius: 0;
  padding: 0.75rem 1rem;
  font-family: 'Consolas', 'Monaco', monospace;
  transition: all 0.2s ease;
  cursor: pointer;
  white-space: nowrap; /* 防止文本换行 */
  overflow: hidden; /* 隐藏溢出内容 */
  text-overflow: ellipsis; /* 显示省略号 */
  position: relative; /* 确保定位上下文正确 */
  z-index: 1; /* 确保在列表中正确显示 */
}

.function-suggestions .list-group-item small {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
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

.function-suggestions .list-group-item.active {
  background-color: #e9f5ff;
  color: #0d6efd;
  border-color: #dee2e6;
  z-index: 2; /* 确保选中项在其他项之上 */
}

.highlight {
  background-color: rgba(255, 230, 0, 0.4);
  font-weight: bold;
  border-radius: 2px;
  padding: 0 2px;
}

.function-name {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 500;
}

/* 表格样式 */
.table {
  margin-bottom: 0;
}

.table th {
  font-weight: 600;
  background-color: #f8f9fa;
}

.table td {
  vertical-align: middle;
}

.table-primary {
  --bs-table-bg: rgba(13, 110, 253, 0.1);
}

.table-hover tbody tr:hover {
  background-color: rgba(0, 0, 0, 0.03);
}
</style> 