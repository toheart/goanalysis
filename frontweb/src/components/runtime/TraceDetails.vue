<!-- Start of Selection -->
<template>
  <div class="trace-details container mt-5">
    <!-- 检查是否有已验证的项目路径 -->
    <div v-if="!hasVerifiedPath" class="text-center">
      <div class="alert alert-warning" role="alert">
        <h4 class="alert-heading mb-3"><i class="bi bi-exclamation-triangle me-2"></i>未设置项目路径</h4>
        <p>请先在主页设置项目路径后再查看追踪详情。</p>
        <hr>
        <p class="mb-0">
          <router-link to="/allgids" class="btn btn-primary">
            <i class="bi bi-arrow-left me-1"></i>返回主页设置项目
          </router-link>
        </p>
      </div>
    </div>

    <!-- 原有内容，当有项目路径时显示 -->
    <div v-else>
      <div class="d-flex justify-content-between align-items-center mb-4">
        <h1 class="page-title">GID: {{ gid }} 的调用详情</h1>
        <button @click="$router.go(-1)" class="btn btn-secondary">
          <i class="bi bi-arrow-left me-1"></i>返回
        </button>
      </div>
      
      <!-- 调用链信息卡片 -->
      <div class="card mb-4">
        <div class="card-header d-flex justify-content-between align-items-center">
          <h5 class="mb-0"><i class="bi bi-list-check me-2"></i>调用链</h5>
          <div class="controls">
            <button class="btn btn-sm btn-outline-primary me-2" @click="expandAll">
              <i class="bi bi-arrows-expand"></i> 展开全部
            </button>
            <button class="btn btn-sm btn-outline-primary" @click="collapseAll">
              <i class="bi bi-arrows-collapse"></i> 折叠全部
            </button>
          </div>
        </div>
        <div class="card-body p-0">
          <div v-if="traceData" class="trace-container p-3">
            <template v-for="(value, key) in processedTraceData" :key="key">
              <div class="stack-item" :style="{ marginLeft: value.indent*20 + 'px' }">
                <div class="trace-row" :class="{'has-children': hasChildren(value)}">
                  <div class="row align-items-center">
                    <div class="col-md-9 trace-info">
                      <div class="d-flex align-items-center">
                        <button v-if="hasChildren(value)" 
                                class="btn btn-sm btn-link toggle-btn me-2" 
                                @click="toggleNode(value.originalIndex)">
                          <i class="bi" :class="isCollapsed(value.originalIndex) ? 'bi-plus-square' : 'bi-dash-square'"></i>
                        </button>
                        <span v-else class="me-4 ms-1"><i class="bi bi-dot"></i></span>
                        <div>
                          <p class="mb-1 function-name">{{ value.name }}</p>
                          <small class="text-muted">耗时: {{ value.timeCost }}</small>
                          <small v-if="value.parentFuncname" class="text-muted d-block">父函数: {{ value.parentFuncname }}</small>
                        </div>
                      </div>
                    </div>
                    <div class="col-md-3 text-end">
                      <button v-if="value.paramCount > 0" 
                              class="btn btn-sm btn-outline-info" 
                              @click="() => viewParameters(value.id)">
                        <i class="bi bi-list-ul me-1"></i>查看参数 ({{ value.paramCount }})
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </div>
          <div v-else class="p-5 text-center">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">加载中...</span>
            </div>
            <p class="mt-3">正在加载调用链数据...</p>
          </div>
        </div>
      </div>
      
      <!-- 调用链统计信息 -->
      <div class="row mb-4" v-if="traceData">
        <div class="col-md-4">
          <div class="card h-100">
            <div class="card-body text-center">
              <h5 class="card-title"><i class="bi bi-diagram-3 me-2"></i>调用深度</h5>
              <p class="display-4">{{ maxDepth }}</p>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card h-100">
            <div class="card-body text-center">
              <h5 class="card-title"><i class="bi bi-clock-history me-2"></i>总耗时</h5>
              <p class="display-4">{{ totalTime }}</p>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card h-100">
            <div class="card-body text-center">
              <h5 class="card-title"><i class="bi bi-code-square me-2"></i>函数调用数</h5>
              <p class="display-4">{{ functionCount }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 参数详情模态框 -->
    <div class="modal fade" id="paramsModal" tabindex="-1" aria-labelledby="paramsModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="paramsModalLabel">
              <i class="bi bi-braces me-2"></i>参数详情
            </h5>
            <button type="button" class="btn-close" @click="closeModal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="table-responsive">
              <table class="table table-striped">
                <thead>
                  <tr>
                    <th scope="col" class="text-center" width="15%">位置</th>
                    <th scope="col">参数值</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(param, index) in parameters" :key="index">
                    <td class="text-center"><span class="badge bg-secondary">{{ param.pos }}</span></td>
                    <td class="param-value">{{ param.param }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModal">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<script>
import axios from '../../axios';
import { Modal } from 'bootstrap';

export default {
  data() {
    return {
      gid: this.$route.params.gid,
      traceData: null,
      traceDataProcessed: [],
      parameters: [],
      modal: null,
      hasVerifiedPath: false,
      collapsedNodes: new Set(),
      maxDepth: 0,
      totalTime: '0ms',
      functionCount: 0,
      childrenMap: new Map(), // 存储函数的子函数
      parentMap: new Map(), // 存储函数的父函数
      expandedNodes: new Set(), // 存储展开的节点
      dbpath: '', // 当前使用的数据库路径
    };
  },
  computed: {
    // 处理后的追踪数据，考虑折叠状态
    processedTraceData() {
      if (!this.traceData || typeof this.traceData !== 'object') return [];
      
      try {
        const result = [];
        const traceArray = Object.values(this.traceData);
        
        console.log('折叠的节点索引:', this.collapsedNodes);
        
        // 创建一个新数组，用于存储处理后的数据
        for (let i = 0; i < traceArray.length; i++) {
          const node = traceArray[i];
          if (!node) continue;
          
          // 检查该节点是否应该显示
          let shouldShow = true;
          
          // 使用 parentFuncname 检查节点是否应该被隐藏
          // 如果节点的任何祖先节点被折叠，则不显示该节点
          let currentNode = node;
          let ancestorIndex = this.findNodeIndexByName(currentNode.parentFuncname);
          
          while (ancestorIndex !== -1) {
            if (this.collapsedNodes.has(ancestorIndex)) {
              shouldShow = false;
              break;
            }
            
            // 继续向上查找祖先节点
            const ancestorNode = traceArray[ancestorIndex];
            if (!ancestorNode || !ancestorNode.parentFuncname) break;
            
            ancestorIndex = this.findNodeIndexByName(ancestorNode.parentFuncname);
          }
          
          if (shouldShow) {
            // 确保节点有所有必要的属性
            const processedNode = {...node};
            processedNode.originalIndex = i;
            result.push(processedNode);
          }
        }
        
        return result;
      } catch (error) {
        console.error("Error processing trace data:", error);
        return [];
      }
    }
  },
  mounted() {
    this.checkProjectPath();
    this.modal = new Modal(document.getElementById('paramsModal'));
  },
  methods: {
    // 检查是否有已验证的项目路径
    checkProjectPath() {
      const savedPath = localStorage.getItem('verifiedProjectPath');
      this.hasVerifiedPath = !!savedPath;
      
      // 如果有已验证的路径，则获取追踪详情
      if (this.hasVerifiedPath) {
        this.fetchTraceDetails();
      }
    },
    
    // 获取当前数据库路径
    getCurrentDbPath() {
      // 如果已经设置了数据库路径，直接返回
      if (this.dbpath) {
        return this.dbpath;
      }
      
      // 否则使用本地存储的已验证路径
      const savedPath = localStorage.getItem('verifiedProjectPath');
      if (savedPath) {
        this.dbpath = savedPath;
        return this.dbpath;
      }
      
      // 如果都没有，返回空字符串
      return '';
    },
    
    async fetchTraceDetails() {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          return;
        }
        
        const response = await axios.post(`/api/runtime/traces/${this.gid}`, {
          dbpath: dbpath
        });
        this.traceData = response.data.traceData || 'No trace data available.';
        
        // 构建函数调用关系映射
        this.buildFunctionRelationships();
      } catch (error) {
        console.error("Error fetching trace details:", error);
      }
    },
    
    // 构建函数调用关系映射
    buildFunctionRelationships() {
      if (!this.traceData) return;
      
      const traceArray = Object.values(this.traceData);
      
      // 清空映射
      this.childrenMap.clear();
      this.parentMap.clear();
      
      // 构建映射关系
      for (let i = 0; i < traceArray.length; i++) {
        const node = traceArray[i];
        if (!node) continue;
        
        // 记录父子关系
        if (node.parentFuncname) {
          // 添加到父函数的子函数列表
          if (!this.childrenMap.has(node.parentFuncname)) {
            this.childrenMap.set(node.parentFuncname, []);
          }
          this.childrenMap.get(node.parentFuncname).push(i);
          
          // 记录节点的父函数
          this.parentMap.set(i, this.findNodeIndexByName(node.parentFuncname));
        }
      }
    },
    
    // 根据函数名查找节点索引
    findNodeIndexByName(funcName) {
      if (!funcName || !this.traceData) return -1;
      
      const traceArray = Object.values(this.traceData);
      for (let i = 0; i < traceArray.length; i++) {
        if (traceArray[i] && traceArray[i].name === funcName) {
          return i;
        }
      }
      return -1;
    },
    
    async viewParameters(id) {
      try {
        const dbpath = this.getCurrentDbPath();
        if (!dbpath) {
          console.error('数据库路径为空');
          return;
        }
        
        const response = await axios.post(`/api/runtime/params/${id}`, {
          dbpath: dbpath
        });
        this.parameters = response.data.params || []; // 修改为返回的参数格式
        this.showModal(); // 显示模态框
      } catch (error) {
        console.error("Error fetching parameters:", error);
      }
    },
    showModal() {
      const modalElement = document.getElementById('paramsModal');
      if (modalElement) {
        const modal = Modal.getOrCreateInstance(modalElement);
        modal.show();
      } else {
        console.error("Modal element not found.");
      }
    },
    closeModal() {
      const modalElement = document.getElementById('paramsModal');
      if (modalElement) {
        const modal = Modal.getInstance(modalElement);
        modal.hide();
      } else {
        console.error("Modal element not found.");
      }
    },
    // 检查节点是否有子节点（基于 parentFuncname）
    hasChildren(node) {
      if (!this.traceData || !node) return false;
      
      // 使用 childrenMap 检查是否有子节点
      return this.childrenMap.has(node.name) && this.childrenMap.get(node.name).length > 0;
    },
    
    // 切换节点的折叠状态
    toggleNode(index) {
      if (this.collapsedNodes.has(index)) {
        this.collapsedNodes.delete(index);
      } else {
        this.collapsedNodes.add(index);
      }
    },
    
    // 检查节点是否已折叠
    isCollapsed(index) {
      return this.collapsedNodes.has(index);
    },
    
    // 新增方法
    expandAll() {
      this.collapsedNodes.clear();
    },
    
    collapseAll() {
      // 找出所有有子节点的节点索引
      if (this.traceData) {
        const traceArray = Object.values(this.traceData);
        for (let i = 0; i < traceArray.length; i++) {
          const node = traceArray[i];
          if (node && this.childrenMap.has(node.name) && this.childrenMap.get(node.name).length > 0) {
            this.collapsedNodes.add(i);
          }
        }
      }
    },
    
    // 计算统计信息
    calculateStats() {
      if (!this.traceData) return;
      
      // 计算最大深度
      this.maxDepth = Math.max(...this.processedTraceData.map(item => item.indent));
      
      // 计算总耗时（取第一个函数的耗时）
      if (this.traceData.length > 0) {
        this.totalTime = this.traceData[0].timeCost || '未知';
      }
      
      // 函数调用数量
      this.functionCount = this.traceData.length;
    }
  },
  watch: {
    processedTraceData() {
      this.calculateStats();
    }
  }
};
</script>

<style scoped>
.trace-container {
  max-height: 70vh;
  overflow-y: auto;
}

.trace-row {
  padding: 0.75rem;
  background-color: white;
  border-radius: var(--border-radius);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: var(--transition);
  margin-bottom: 0.5rem;
}

.trace-row:hover {
  background-color: var(--light-bg);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.trace-row.has-children {
  border-left: 3px solid var(--primary-color);
}

.function-name {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 500;
}

.toggle-btn {
  padding: 0;
  color: var(--primary-color);
  background: transparent;
  border: none;
  font-size: 1.2rem;
}

.toggle-btn:hover {
  color: var(--secondary-color);
}

.param-value {
  font-family: 'Consolas', 'Monaco', monospace;
  word-break: break-all;
  white-space: pre-wrap;
  background-color: #f8f9fa;
  padding: 0.5rem;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .trace-row .row {
    flex-direction: column;
  }
  
  .trace-row .col-md-3 {
    margin-top: 0.5rem;
    text-align: left !important;
  }
}
</style> 