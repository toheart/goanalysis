<!-- Start of Selection -->
<template>
  <div class="container mt-5">
    <!-- 检查是否有已验证的项目路径 -->
    <div v-if="!hasVerifiedPath" class="text-center">
      <div class="alert alert-warning" role="alert">
        <h4 class="alert-heading mb-3">未设置项目路径</h4>
        <p>请先在主页设置项目路径后再查看追踪详情。</p>
        <hr>
        <p class="mb-0">
          <router-link to="/" class="btn btn-primary">
            返回主页设置项目
          </router-link>
        </p>
      </div>
    </div>

    <!-- 原有内容，当有项目路径时显示 -->
    <div v-else>
      <h1 class="text-center">Trace Details for GID: {{ gid }}</h1>
      <button @click="$router.go(-1)" class="btn btn-secondary mb-3" style="float: right;">返回</button><br>
      <div v-if="traceData">
        <div class="mt-4">
          <template v-for="(value, key) in processedTraceData" :key="key">
            <div class="stack-item" :style="{ marginLeft: value.indent*5 + 'em' }">
              <div class="row trace-row">
                <div class="col-md-9 trace-info">
                  <div class="d-flex align-items-center">
                    <button v-if="hasChildren(value)" 
                            class="btn btn-sm btn-link toggle-btn me-2" 
                            @click="toggleNode(value.originalIndex)">
                      <svg v-if="isCollapsed(value.originalIndex)" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                        <path d="M14 1a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h12zM2 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H2z"/>
                        <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
                      </svg>
                      <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                        <path d="M14 1a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h12zM2 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H2z"/>
                        <path d="M4 8a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7A.5.5 0 0 1 4 8z"/>
                      </svg>
                    </button>
                    <span v-else class="me-4"></span>
                    <p class="mb-1 text-left">Name: {{ value.name }} ({{ value.timeCost }})</p>
                  </div>
                </div>
                <div class="col-md-3 text-end">
                  <button v-if="value.paramCount > 0" class="btn btn-sm btn-primary" @click="() => viewParameters(value.id)">查看参数</button>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
      <div v-else>
        <p>Loading...</p>
      </div>
    </div>

    <!-- 修改模态框部分 -->
    <div class="modal fade" id="paramsModal" tabindex="-1" aria-labelledby="paramsModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="paramsModalLabel">参数详情</h5>
            <button type="button" class="btn-close" @click="closeModal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="table-responsive">
              <table class="table">
                <thead>
                  <tr>
                    <th scope="col" class="text-center">位置</th>
                    <th scope="col" class="text-center">参数</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(param, index) in parameters" :key="index">
                    <td class="text-center">{{ param.pos }}</td>
                    <td class="text-break">{{ param.param }}</td>
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
import axios from '../axios';
import { Modal } from 'bootstrap';

export default {
  data() {
    return {
      gid: this.$route.params.gid,
      traceData: null,
      parameters: [], // 用于存储参数数据
      hasVerifiedPath: false, // 新增：用于检查是否有已验证的项目路径
      collapsedNodes: [], // 存储已折叠节点的索引
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
          
          // 检查该节点的所有潜在父节点是否被折叠
          for (let j = 0; j < i; j++) {
            const potentialParent = traceArray[j];
            if (!potentialParent) continue;
            
            // 如果找到一个折叠的节点，且它是当前节点的父节点（基于缩进）
            if (this.collapsedNodes.includes(j) && 
                this.isParentOf(potentialParent, node, traceArray, j, i)) {
              console.log(`节点 ${node.name} (索引: ${i}) 被隐藏，因为父节点 ${potentialParent.name} (索引: ${j}) 已折叠`);
              shouldShow = false;
              break;
            }
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
    // 检查本地存储中是否有已验证的路径
    const savedPath = localStorage.getItem('verifiedProjectPath');
    if (savedPath) {
      this.hasVerifiedPath = true;
      this.fetchTraceDetails();
    }
  },
  methods: {
    async fetchTraceDetails() {
      try {
        const response = await axios.get(`/api/traces/${this.gid}`);
        this.traceData = response.data.traceData || 'No trace data available.';
      } catch (error) {
        console.error("Error fetching trace details:", error);
      }
    },
    async viewParameters(id) {
      try {
        const response = await axios.get(`/api/params/${id}`);
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
    // 检查节点是否有子节点（基于缩进）
    hasChildren(node) {
      if (!this.traceData) return false;
      if (!node || typeof node.originalIndex === 'undefined') return false;
      
      const traceArray = Object.values(this.traceData);
      const originalIndex = node.originalIndex;
      
      // 确保下一个节点存在
      if (originalIndex >= traceArray.length - 1) return false;
      
      const nextNode = traceArray[originalIndex + 1];
      // 确保下一个节点有indent属性
      if (!nextNode || typeof nextNode.indent === 'undefined') return false;
      
      // 如果下一个节点的缩进比当前节点大，则当前节点有子节点
      return nextNode.indent > node.indent;
    },
    
    // 切换节点的折叠状态
    toggleNode(index) {
      const idx = this.collapsedNodes.indexOf(index);
      if (idx !== -1) {
        // 节点已折叠，展开它
        console.log(`展开节点: 索引 ${index}`);
        // 使用 splice 确保数组变化被 Vue 检测到
        this.collapsedNodes.splice(idx, 1);
      } else {
        // 节点未折叠，折叠它
        console.log(`折叠节点: 索引 ${index}`);
        // 使用 push 确保数组变化被 Vue 检测到
        this.collapsedNodes.push(index);
      }
    },
    
    // 检查节点是否已折叠
    isCollapsed(index) {
      return this.collapsedNodes.includes(index);
    },
    
    // 判断节点A是否是节点B的父节点（基于缩进）
    isParentOf(nodeA, nodeB, traceArray, indexA, indexB) {
      // 如果B的缩进不大于A，则A不可能是B的父节点
      if (nodeB.indent <= nodeA.indent) return false;
      
      // 检查A和B之间是否有缩进级别小于等于A的节点
      // 如果有，则A不是B的直接父节点
      for (let i = indexA + 1; i < indexB; i++) {
        const middleNode = traceArray[i];
        if (!middleNode) continue;
        
        if (middleNode.indent <= nodeA.indent) {
          return false;
        }
      }
      
      return true;
    }
  },
};
</script>

<style scoped>
.modal-dialog {
  margin: 1.75rem auto;
  max-width: 800px;
}

.modal-content {
  border-radius: 0.5rem;
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
}

.modal-body {
  max-height: 70vh;
  overflow-y: auto;
}

.table-responsive {
  margin: 0;
}

.text-break {
  word-break: break-word;
  max-width: 500px;
}

/* 保留原有的其他样式 */
.list-group-item {
  border: 1px solid #ddd;
  border-radius: 5px;
  margin-bottom: 10px;
}

.table {
  margin: 0;
  width: 100%;
}

/* 新增样式 */
.alert {
  max-width: 600px;
  margin: 0 auto;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.alert-heading {
  color: #856404;
}

/* 新增和修改的样式 */
.stack-item {
  margin-bottom: 8px;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.stack-item:hover {
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
}

.trace-row {
  background-color: #f8f9fa;
  padding: 8px;
  margin: 0;
  border-left: 4px solid #007bff;
}

.trace-info {
  background-color: #f8f9fa;
  padding: 8px 12px;
}

.btn-primary {
  background-color: #007bff;
  border-color: #007bff;
}

.btn-primary:hover {
  background-color: #0069d9;
  border-color: #0062cc;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .stack-item {
    margin-left: 0 !important;
    padding-left: calc(var(--indent, 0) * 20px);
  }
  
  .trace-row {
    flex-direction: column;
  }
  
  .col-md-3.text-end {
    text-align: left !important;
    margin-top: 8px;
  }
}

/* 修改折叠图标样式 */
.toggle-btn {
  padding: 0;
  color: #007bff;
  background: transparent;
  border: none;
  transition: all 0.2s;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toggle-icon {
  font-size: 18px;
  font-weight: bold;
  line-height: 1;
}

/* 删除不再需要的图标样式 */
.bi,
.bi-plus-square::before,
.bi-dash-square::before {
  display: none;
}
</style> 