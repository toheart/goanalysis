<template>
  <div class="function-analysis">
    <!-- 函数搜索区域 -->
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-search me-2"></i>函数查询</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-8 search-container">
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-code-square"></i></span>
              <input
                v-model="functionName"
                type="text"
                placeholder="输入函数名称搜索"
                class="form-control"
                @input="onFunctionNameInput"
                @focus="showFunctionSuggestions = true"
                @blur="hideSuggestionsDelayed"
              />
              <button class="btn btn-primary" @click="searchFunction">
                <i class="bi bi-search me-1"></i>查询
              </button>
            </div>
            <small class="text-muted mt-2 d-block">
              输入函数名称进行查询，支持模糊匹配
            </small>
          </div>
          <div class="col-md-4">
            <div class="form-group">
              <div class="btn-group w-100">
                <button 
                  class="btn btn-primary"
                  disabled
                >
                  <i class="bi bi-arrow-down-square me-1"></i>调用者
                </button>
              </div>
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

    <!-- 函数调用链分析结果 -->
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">加载中...</span>
      </div>
      <p class="mt-3">正在分析函数调用关系...</p>
    </div>

    <div v-else-if="functionName && !loading && !functionCallData.length" class="card mb-4">
      <div class="card-body text-center py-5">
        <i class="bi bi-search text-muted display-1"></i>
        <h4 class="mt-3">未找到相关函数调用关系</h4>
        <p>请尝试其他函数名称</p>
      </div>
    </div>

    <div v-else-if="functionCallData.length > 0" class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="bi bi-arrow-down-square"></i>
          调用关系分析: <code>{{ functionName }}</code>
        </h5>
        <div class="btn-group">
          <button class="btn btn-sm btn-outline-primary" @click="exportToCSV">
            <i class="bi bi-download me-1"></i>导出
          </button>
        </div>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover table-striped mb-0">
            <thead class="table-light">
              <tr>
                <th>函数名称</th>
                <th>包路径</th>
                <th class="text-center">调用层级</th>
                <th class="text-center">调用次数</th>
                <th class="text-center">平均耗时</th>
                <th class="text-center">操作</th>
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
                    <i class="bi bi-search me-1"></i>查询
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
import axios from '../axios';

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
      functionName: '',
      searchType: 'caller', // 固定为'caller'
      functionNames: [],
      filteredFunctionNames: [],
      showFunctionSuggestions: false,
      suggestionsTimer: null,
      loading: false,
      functionCallData: [],
      collapsedNodes: new Set(),
      isComponentMounted: false
    };
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
    
    // 检查API连接状态
    this.checkApiConnection().then(connected => {
      if (!connected) {
        console.warn('API连接检查失败，可能会影响功能使用');
      }
    });
    
    // 添加点击事件监听器，点击页面其他地方时隐藏建议列表
    document.addEventListener('click', this.handleDocumentClick);
  },
  beforeUnmount() {
    this.isComponentMounted = false;
    
    // 移除事件监听器
    document.removeEventListener('click', this.handleDocumentClick);
    
    // 清除计时器
    if (this.suggestionsTimer) {
      clearTimeout(this.suggestionsTimer);
    }
  },
  methods: {
    // 检查API连接状态
    async checkApiConnection() {
      try {
        console.log('检查API连接状态...');
        const response = await axios.get('/api/check/db');
        console.log('API连接状态:', response.data);
        return response.data.exists;
      } catch (error) {
        console.error('API连接检查失败:', error);
        return false;
      }
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
    
    // 选择函数
    selectFunction(name) {
      this.functionName = name;
      this.filteredFunctionNames = [];
      this.showFunctionSuggestions = false;
    },
    
    // 查询函数
    async searchFunction() {
      if (!this.functionName) return;
      
      this.loading = true;
      this.functionCallData = [];
      this.collapsedNodes.clear();
      
      try {
        const response = await axios.post('/api/function/analysis', {
          functionName: this.functionName,
          type: 'caller', // 固定为caller
          path: this.projectPath
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
      this.functionName = name;
      this.searchFunction();
    },
    
    // 获取函数名列表
    async fetchFunctionNames() {
      try {
        const response = await axios.get('/api/functions');
        this.functionNames = response.data.functionNames || [];
      } catch (error) {
        console.error('获取函数名列表失败:', error);
      }
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

@media (max-width: 768px) {
  .function-suggestions {
    width: 90%;
    top: 180px;
  }
}
</style> 