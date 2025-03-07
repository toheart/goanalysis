<template>
  <div class="trace-graph container mt-5">
    <h1 class="page-title text-center mb-4">Goroutine ({{ gid }}) 调用图</h1>
    
    <!-- 使用向导提示 -->
    <div v-if="showGuide" class="guide-overlay">
      <div class="guide-content">
        <h4><i class="bi bi-info-circle me-2"></i>使用指南</h4>
        <ul class="guide-list">
          <li><i class="bi bi-mouse"></i> <strong>拖拽:</strong> 点击空白处并拖动可移动整个图</li>
          <li><i class="bi bi-zoom-in"></i> <strong>缩放:</strong> 使用鼠标滚轮放大或缩小</li>
          <li><i class="bi bi-hand-index-thumb"></i> <strong>选择:</strong> 点击节点查看详细信息</li>
          <li><i class="bi bi-arrows-move"></i> <strong>重新布局:</strong> 使用上方的布局选项改变视图</li>
        </ul>
        <button class="btn btn-primary" @click="closeGuide">了解了</button>
      </div>
    </div>
    
    <!-- 控制面板 -->
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-sliders me-2"></i>图形控制</h5>
      </div>
      <div class="card-body">
        <div class="control-panel">
          <div class="row">
            <div class="col-md-4">
              <div class="control-group">
                <label class="form-label">视图控制</label>
                <div class="btn-group w-100">
                  <button class="btn btn-outline-primary" @click="resetView" title="重置视图">
                    <i class="bi bi-arrow-counterclockwise me-1"></i>重置
                  </button>
                  <button class="btn btn-outline-primary" @click="fitView" title="适应屏幕">
                    <i class="bi bi-fullscreen me-1"></i>适应
                  </button>
                </div>
              </div>
            </div>
            
            <div class="col-md-4">
              <div class="control-group">
                <label class="form-label">布局选择</label>
                <select class="form-select" v-model="currentLayout" @change="changeLayout(currentLayout)">
                  <option value="dagre">层次布局</option>
                  <option value="cose">力导向布局</option>
                  <option value="grid">网格布局</option>
                  <option value="concentric">同心圆布局</option>
                </select>
              </div>
            </div>
            
            <div class="col-md-4">
              <div class="control-group">
                <label class="form-label">显示选项</label>
                <div class="form-check form-switch">
                  <input class="form-check-input" type="checkbox" id="edgeLabelsSwitch" v-model="showEdgeLabels">
                  <label class="form-check-label" for="edgeLabelsSwitch">显示边标签</label>
                </div>
                <div class="form-check form-switch">
                  <input class="form-check-input" type="checkbox" id="nodeLabelsSwitch" v-model="showNodeLabels">
                  <label class="form-check-label" for="nodeLabelsSwitch">显示节点标签</label>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 图形容器 -->
    <div class="card mb-4">
      <div class="card-body p-0">
        <div id="cy" class="cy-container" ref="cyContainer">
          <!-- 加载指示器 -->
          <div v-if="loading" class="loading-overlay">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">加载中...</span>
            </div>
            <p class="mt-2">正在生成调用图...</p>
          </div>
          
          <!-- 拖拽提示 -->
          <div v-if="!loading && showDragHint" class="drag-hint">
            <i class="bi bi-arrows-move"></i> 点击并拖动可移动图形
          </div>
        </div>
      </div>
    </div>
    
    <!-- 节点详情面板 -->
    <div v-if="selectedNode" class="card mb-4 node-details">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-info-circle me-2"></i>函数详情</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-6">
            <p><strong>函数名称:</strong></p>
            <pre class="function-name">{{ selectedNode.name }}</pre>
          </div>
          <div class="col-md-6">
            <div class="row">
              <div class="col-6">
                <div class="stat-card">
                  <div class="stat-label">调用次数</div>
                  <div class="stat-value">{{ selectedNode.callCount || 0 }}</div>
                </div>
              </div>
              <div class="col-6">
                <div class="stat-card">
                  <div class="stat-label">耗时</div>
                  <div class="stat-value">{{ selectedNode.timeCost || 'N/A' }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 图形统计信息 -->
    <div class="row mb-4">
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-diagram-3 me-2"></i>节点数量</h5>
            <p class="display-4">{{ nodeCount }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-arrow-left-right me-2"></i>边数量</h5>
            <p class="display-4">{{ edgeCount }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <h5 class="card-title"><i class="bi bi-layers me-2"></i>最大深度</h5>
            <p class="display-4">{{ maxDepth }}</p>
          </div>
        </div>
      </div>
    </div>
    
    <div class="mt-3 text-center">
      <router-link :to="{ name: 'TraceViewer' }" class="btn btn-primary">
        <i class="bi bi-arrow-left me-1"></i>返回列表
      </router-link>
    </div>
  </div>
</template>

<script>
import axios from '../axios';
import cytoscape from 'cytoscape';
import dagre from 'cytoscape-dagre';
import cose from 'cytoscape-cose-bilkent';
import popper from 'cytoscape-popper';

// 注册布局插件
cytoscape.use(dagre);
cytoscape.use(cose);
cytoscape.use(popper);

export default {
  data() {
    return {
      gid: this.$route.params.gid,
      cy: null,
      graphData: null,
      selectedNode: null,
      currentLayout: 'dagre',
      showEdgeLabels: true,
      showNodeLabels: true,
      loading: true,
      showGuide: true,
      showDragHint: true,
      dragHintTimer: null,
      nodeCount: 0,
      edgeCount: 0,
      maxDepth: 0
    };
  },
  mounted() {
    this.fetchGraphData();
    
    // 检查是否已经显示过指南
    const hasSeenGuide = localStorage.getItem('hasSeenGraphGuide');
    if (hasSeenGuide) {
      this.showGuide = false;
    }
  },
  beforeUnmount() {
    if (this.dragHintTimer) {
      clearTimeout(this.dragHintTimer);
    }
  },
  watch: {
    showEdgeLabels(newVal) {
      if (this.cy) {
        this.cy.edges().style('text-opacity', newVal ? 1 : 0);
      }
    },
    showNodeLabels() {
      this.updateNodeStyles();
    }
  },
  methods: {
    async fetchGraphData() {
      try {
        this.loading = true;
        const response = await axios.get(`/api/traces/${this.gid}/graph`);
        this.graphData = {
          nodes: response.data.nodes || [],
          edges: response.data.edges || []
        };
        this.initCytoscape();
      } catch (error) {
        console.error("获取调用图数据失败:", error);
        alert("获取图形数据失败，请检查后端服务是否正常运行");
      } finally {
        this.loading = false;
        
        // 设置拖拽提示定时器
        this.dragHintTimer = setTimeout(() => {
          this.showDragHint = false;
        }, 5000); // 5秒后自动隐藏提示
      }
    },
    
    closeGuide() {
      this.showGuide = false;
      localStorage.setItem('hasSeenGraphGuide', 'true');  
    },
    
    initCytoscape() {
      // 准备节点和边的数据
      const elements = this.prepareGraphElements();
      
      // 初始化 Cytoscape 实例
      this.cy = cytoscape({
        container: document.getElementById('cy'),
        elements: elements,
        style: [
          {
            selector: 'node',
            style: {
              'label': 'data(label)',
              'text-valign': 'center',
              'text-halign': 'center',
              'background-color': '#4285f4',
              'color': '#fff',
              'text-outline-width': 1,
              'text-outline-color': '#4285f4',
              'width': 'mapData(callCount, 0, 10, 30, 60)',
              'height': 'mapData(callCount, 0, 10, 30, 60)',
              'border-width': 2,
              'border-color': '#2a56c6',
              'text-wrap': 'ellipsis',
              'text-max-width': '80px'
            }
          },
          {
            selector: 'edge',
            style: {
              'width': 2,
              'line-color': '#999',
              'target-arrow-color': '#999',
              'target-arrow-shape': 'triangle',
              'curve-style': 'bezier',
              'label': 'data(label)',
              'text-rotation': 'autorotate',
              'text-margin-y': -10,
              'text-background-color': 'white',
              'text-background-opacity': 0.7,
              'text-background-padding': 2,
              'text-opacity': this.showEdgeLabels ? 1 : 0
            }
          },
          {
            selector: ':selected',
            style: {
              'background-color': '#ff5722',
              'border-color': '#ff8a65',
              'border-width': 3,
              'line-color': '#ff5722',
              'target-arrow-color': '#ff5722'
            }
          }
        ],
        layout: {
          name: this.currentLayout,
          rankDir: 'TB',
          padding: 50,
          animate: true
        },
        wheelSensitivity: 0.3
      });
      
      // 添加事件监听
      this.cy.on('tap', 'node', (evt) => {
        const node = evt.target;
        this.selectedNode = {
          id: node.id(),
          name: node.data('name'),
          callCount: node.data('callCount')
        };
      });
      
      this.cy.on('tap', (evt) => {
        if (evt.target === this.cy) {
          this.selectedNode = null;
        }
      });
      
      // 监听拖拽事件，隐藏拖拽提示
      this.cy.on('dragfree', () => {
        this.showDragHint = false;
        if (this.dragHintTimer) {
          clearTimeout(this.dragHintTimer);
        }
      });
      
      // 初始化完成后适应屏幕
      this.fitView();
    },
    
    prepareGraphElements() {
      // 这里需要根据后端返回的数据格式进行转换
      // 假设 this.graphData 包含 nodes 和 edges 数组
      const elements = [];
      
      if (this.graphData && this.graphData.nodes) {
        // 添加节点
        this.graphData.nodes.forEach(node => {
          // 处理节点名称，确保安全显示
          const processedName = this.sanitizeFunctionName(node.name);
          const shortName = this.shortenFunctionName(processedName);
          
          elements.push({
            data: {
              id: node.id,
              label: shortName,
              name: processedName,
              callCount: node.callCount || 1,
              originalName: node.name // 保存原始名称用于详情显示
            }
          });
        });
        
        // 添加边
        if (this.graphData.edges) {
          this.graphData.edges.forEach(edge => {
            elements.push({
              data: {
                id: `${edge.source}-${edge.target}`,
                source: edge.source,
                target: edge.target,
                label: edge.label || ''
              }
            });
          });
        }
      }
      
      return elements;
    },
    
    sanitizeFunctionName(name) {
      if (!name) return '';
      
      // 替换可能导致显示问题的字符
      return name
        .replace(/\[/g, '(')
        .replace(/\]/g, ')')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
    },
    
    shortenFunctionName(name) {
      if (!name) return '';
      
      // 移除包路径，只保留最后一部分
      let parts = name.split('.');
      let shortName = parts[parts.length - 1];
      
      // 如果名称中包含括号，尝试提取函数名
      if (shortName.includes('(')) {
        shortName = shortName.split('(')[0];
      }
      
      // 如果名称太长，截断它
      if (shortName.length > 20) {
        shortName = shortName.substring(0, 17) + '...';
      }
      
      return shortName;
    },
    
    changeLayout(layoutName) {
      console.log(`Changing layout to: ${layoutName}`);
      this.currentLayout = layoutName;
      
      // 定义布局选项
      const layoutOptions = {
        name: layoutName,
        fit: true,
        animate: true,
        padding: 50
      };
      
      // 根据不同布局添加特定选项
      if (layoutName === 'dagre') {
        layoutOptions.rankDir = 'TB';
        layoutOptions.rankSep = 100;
        layoutOptions.nodeSep = 50;
      } else if (layoutName === 'cose') {
        layoutOptions.idealEdgeLength = 100;
        layoutOptions.nodeOverlap = 20;
        layoutOptions.refresh = 20;
        layoutOptions.fit = true;
        layoutOptions.padding = 30;
        layoutOptions.randomize = false;
      } else if (layoutName === 'concentric') {
        layoutOptions.minNodeSpacing = 50;
        layoutOptions.concentric = function(node) {
          return node.degree();
        };
        layoutOptions.levelWidth = function(nodes) {
          return nodes.maxDegree() / 4;
        };
      }
      
      // 应用新布局
      if (this.cy) {
        const layout = this.cy.layout(layoutOptions);
        layout.run();
      }
    },
    
    resetView() {
      if (this.cy) {
        this.cy.reset();
      }
    },
    
    fitView() {
      if (this.cy) {
        this.cy.fit();
      }
    },
    
    // 更新统计信息
    updateStats() {
      if (this.cy) {
        this.nodeCount = this.cy.nodes().length;
        this.edgeCount = this.cy.edges().length;
        
        // 计算最大深度 - 从根节点到最远叶子节点的最长路径
        const roots = this.cy.nodes().roots();
        if (roots.length > 0) {
          let maxDepth = 0;
          roots.forEach(root => {
            const bfs = this.cy.elements().bfs({
              roots: root,
              directed: true
            });
            const depths = {};
            bfs.path.forEach(ele => {
              if (ele.isNode()) {
                const parent = ele.incomers().nodes();
                depths[ele.id()] = parent.length > 0 ? depths[parent.id()] + 1 : 0;
                maxDepth = Math.max(maxDepth, depths[ele.id()]);
              }
            });
          });
          this.maxDepth = maxDepth;
        }
      }
    },
    
    // 修改节点样式
    updateNodeStyles() {
      if (this.cy) {
        this.cy.style()
          .selector('node')
          .style({
            'label': this.showNodeLabels ? 'data(label)' : '',
            'text-opacity': this.showNodeLabels ? 1 : 0
          })
          .update();
      }
    }
  }
};
</script>

<style scoped>
.cy-container {
  height: 70vh;
  position: relative;
  background-color: white;
}

.guide-list {
  list-style-type: none;
  padding-left: 0;
}

.guide-list li {
  padding: 0.75rem 0;
  border-bottom: 1px solid #eee;
}

.guide-list li:last-child {
  border-bottom: none;
}

.guide-list i {
  margin-right: 0.5rem;
  color: var(--primary-color);
}

.control-group {
  margin-bottom: 1rem;
}

.drag-hint {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background-color: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 2rem;
  font-size: 0.9rem;
  pointer-events: none;
  opacity: 0.8;
}

.function-name {
  background-color: #f8f9fa;
  padding: 0.75rem;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  overflow-x: auto;
  margin-bottom: 0;
}

.stat-card {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 1rem;
  text-align: center;
  height: 100%;
}

.stat-label {
  font-size: 0.9rem;
  color: #6c757d;
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--primary-color);
}

@media (max-width: 768px) {
  .control-panel .row {
    flex-direction: column;
  }
  
  .control-panel .col-md-4 {
    margin-bottom: 1rem;
  }
}
</style> 