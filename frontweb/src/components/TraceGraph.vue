<template>
  <div class="container mt-5">
    <h1 class="text-center">goroutine({{gid}})调用图展示:</h1>
    
    <!-- 使用向导提示 -->
    <div v-if="showGuide" class="guide-overlay">
      <div class="guide-content">
        <h4>使用指南</h4>
        <ul>
          <li><i class="bi bi-mouse"></i> <strong>拖拽:</strong> 点击空白处并拖动可移动整个图</li>
          <li><i class="bi bi-zoom-in"></i> <strong>缩放:</strong> 使用鼠标滚轮放大或缩小</li>
          <li><i class="bi bi-hand-index-thumb"></i> <strong>选择:</strong> 点击节点查看详细信息</li>
          <li><i class="bi bi-arrows-move"></i> <strong>重新布局:</strong> 使用上方的布局选项改变视图</li>
        </ul>
        <button class="btn btn-primary" @click="closeGuide">了解了</button>
      </div>
    </div>
    
    <!-- 控制面板 -->
    <div class="control-panel mb-3">
      <div class="btn-group me-2" role="group">
        <button class="btn btn-outline-primary" @click="resetView">重置视图</button>
        <button class="btn btn-outline-primary" @click="fitView">适应屏幕</button>
      </div>
      
      <!-- 修改布局下拉菜单 - 使用简单的按钮组替代下拉菜单 -->
      <div class="btn-group me-2" role="group">
        <button class="btn btn-outline-primary" @click="changeLayout('dagre')">层次布局</button>
        <button class="btn btn-outline-primary" @click="changeLayout('cose')">力导向布局</button>
        <button class="btn btn-outline-primary" @click="changeLayout('grid')">网格布局</button>
        <button class="btn btn-outline-primary" @click="changeLayout('concentric')">同心圆布局</button>
      </div>
      
      <button class="btn btn-outline-info me-2" @click="showGuide = true">
        <i class="bi bi-question-circle"></i> 帮助
      </button>
      
      <div class="form-check form-switch d-inline-block">
        <input class="form-check-input" type="checkbox" id="edgeLabelsSwitch" v-model="showEdgeLabels">
        <label class="form-check-label" for="edgeLabelsSwitch">显示边标签</label>
      </div>
    </div>
    
    <!-- 图形容器 -->
    <div id="cy" class="cy-container highlighted-container" ref="cyContainer">
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
    
    <!-- 节点详情面板 -->
    <div v-if="selectedNode" class="node-details mt-3 p-3">
      <h5>函数详情</h5>
      <p><strong>名称:</strong> <code>{{ selectedNode.name }}</code></p>
      <p><strong>调用次数:</strong> {{ selectedNode.callCount || 0 }}</p>
    </div>
    
    <div class="mt-3">
      <router-link :to="{ name: 'TraceViewer' }" class="btn btn-primary">返回</router-link>
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
      loading: true,
      showGuide: true,
      showDragHint: true,
      dragHintTimer: null
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
    }
  }
};
</script>

<style scoped>
.cy-container {
  width: 100%;
  height: 600px;
  border: 1px solid #ddd;
  border-radius: 5px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  background-color: #f9f9f9;
  position: relative;
}

.highlighted-container {
  border: 2px solid #4285f4;
  box-shadow: 0 4px 15px rgba(66, 133, 244, 0.3);
  transition: all 0.3s ease;
}

.highlighted-container:hover {
  box-shadow: 0 6px 20px rgba(66, 133, 244, 0.4);
}

.control-panel {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 10px;
  background-color: #f5f5f5;
  border-radius: 5px;
  margin-bottom: 15px;
}

.node-details {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 5px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.8);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 10;
}

.guide-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.guide-content {
  background-color: white;
  padding: 30px;
  border-radius: 10px;
  max-width: 500px;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
}

.guide-content h4 {
  margin-bottom: 20px;
  color: #4285f4;
}

.guide-content ul {
  list-style-type: none;
  padding-left: 0;
  margin-bottom: 25px;
}

.guide-content li {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
}

.guide-content i {
  margin-right: 10px;
  color: #4285f4;
  font-size: 1.2em;
}

.drag-hint {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background-color: rgba(66, 133, 244, 0.9);
  color: white;
  padding: 10px 20px;
  border-radius: 30px;
  font-size: 16px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  z-index: 5;
  animation: pulse 2s infinite;
}

.drag-hint i {
  margin-right: 8px;
  font-size: 1.2em;
}

@keyframes pulse {
  0% {
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    transform: translate(-50%, -50%) scale(1.05);
  }
  100% {
    transform: translate(-50%, -50%) scale(1);
  }
}

/* 添加一个新的样式，使控制面板在小屏幕上可以滚动 */
@media (max-width: 992px) {
  .control-panel {
    flex-wrap: wrap;
    overflow-x: auto;
    white-space: nowrap;
    padding: 10px 5px;
  }
  
  .btn-group {
    margin-bottom: 5px;
  }
}
</style> 