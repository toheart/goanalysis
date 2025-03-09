<template>
  <div class="function-call-graph">
    <div v-if="visible" class="modal-backdrop" @click="closeModal"></div>
    <div v-if="visible" class="modal-container">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            <i class="bi bi-diagram-3 me-2"></i>函数调用关系图 (GID: {{ gid }})
          </h5>
          <button type="button" class="btn-close" @click="closeModal"></button>
        </div>
        <div class="modal-body">
          <div v-if="loading" class="text-center py-5">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">加载中...</span>
            </div>
            <p class="mt-3">正在加载函数调用图数据...</p>
          </div>
          <div v-else-if="error" class="alert alert-danger">
            <i class="bi bi-exclamation-triangle-fill me-2"></i>{{ error }}
            <div v-if="showDebugInfo" class="mt-2">
              <button class="btn btn-sm btn-outline-danger" @click="retryWithMockData">
                使用模拟数据
              </button>
              <button class="btn btn-sm btn-outline-secondary ms-2" @click="retryFetch">
                重试
              </button>
            </div>
          </div>
          <div v-else-if="!hasData" class="alert alert-warning">
            <i class="bi bi-info-circle-fill me-2"></i>没有找到调用图数据
            <div v-if="showDebugInfo" class="mt-2">
              <button class="btn btn-sm btn-outline-primary" @click="useMockDataAndRefresh">
                使用模拟数据
              </button>
            </div>
          </div>
          <!-- 始终保持图表容器存在，但根据条件控制显示 -->
          <div class="chart-container" ref="chartContainer" :style="{ display: (!loading && !error && hasData) ? 'block' : 'none' }"></div>
        </div>
        <div class="modal-footer">
          <div class="chart-controls">
            <div class="btn-group me-2">
              <button class="btn btn-sm btn-outline-secondary" @click="zoomIn">
                <i class="bi bi-zoom-in"></i>
              </button>
              <button class="btn btn-sm btn-outline-secondary" @click="zoomOut">
                <i class="bi bi-zoom-out"></i>
              </button>
              <button class="btn btn-sm btn-outline-secondary" @click="resetZoom">
                <i class="bi bi-arrows-fullscreen"></i>
              </button>
            </div>
            <div class="btn-group">
              <button class="btn btn-sm btn-outline-primary" @click="changeLayout('force')" :class="{ active: currentLayout === 'force' }">
                力导向图
              </button>
              <button class="btn btn-sm btn-outline-primary" @click="changeLayout('circular')" :class="{ active: currentLayout === 'circular' }">
                环形布局
              </button>
              <button class="btn btn-sm btn-outline-primary" @click="changeLayout('tree')" :class="{ active: currentLayout === 'tree' }">
                树形布局
              </button>
            </div>
          </div>
          <div>
            <button type="button" class="btn btn-sm btn-outline-info me-2" @click="toggleDebugInfo" v-if="!showDebugInfo">
              <i class="bi bi-bug"></i>
            </button>
            <button type="button" class="btn btn-secondary" @click="closeModal">关闭</button>
          </div>
        </div>
        
        <!-- 调试信息面板 -->
        <div v-if="showDebugInfo" class="debug-panel">
          <div class="debug-header">
            <h6 class="mb-0">调试信息</h6>
            <button type="button" class="btn-close btn-sm" @click="showDebugInfo = false"></button>
          </div>
          <div class="debug-body">
            <div class="mb-2">
              <strong>GID:</strong> {{ gid }}
            </div>
            <div class="mb-2">
              <strong>数据库路径:</strong> {{ dbpath }}
            </div>
            <div class="mb-2">
              <strong>当前布局:</strong> {{ currentLayout }}
            </div>
            <div class="mb-2">
              <strong>节点数量:</strong> {{ graphData && graphData.nodes ? graphData.nodes.length : 0 }}
            </div>
            <div class="mb-2">
              <strong>边数量:</strong> {{ graphData && graphData.edges ? graphData.edges.length : 0 }}
            </div>
            <div class="mb-3">
              <button class="btn btn-sm btn-outline-primary" @click="refreshChart">
                刷新图表
              </button>
              <button class="btn btn-sm btn-outline-warning ms-2" @click="useMockDataAndRefresh">
                使用模拟数据
              </button>
              <button class="btn btn-sm btn-outline-info ms-2" @click="viewRawApiData">
                查看原始数据
              </button>
              <button class="btn btn-sm btn-outline-danger ms-2" @click="tryAlternativeFormat">
                尝试替代格式
              </button>
            </div>
            <div class="small text-muted">
              <strong>API URL:</strong> /api/runtime/traces/graph
            </div>
            <div class="mt-2">
              <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" id="autoFixData" v-model="autoFixData">
                <label class="form-check-label" for="autoFixData">自动修复数据</label>
              </div>
            </div>
            <div v-if="dataFormatInfo" class="mt-2 alert alert-info small">
              {{ dataFormatInfo }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../../axios';
import * as echarts from 'echarts';

export default {
  name: 'FunctionCallGraph',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    gid: {
      type: String,
      required: true
    },
    dbpath: {
      type: String,
      required: true
    },
    useMockData: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      loading: true,
      error: null,
      chartInstance: null,
      graphData: null,
      currentLayout: 'force',
      zoomLevel: 1,
      hasData: false,
      showDebugInfo: false,
      autoFixData: true,
      originalApiData: null,
      dataFormatInfo: '',
      alternativeFormatIndex: 0,
      resizeTimer: null
    };
  },
  watch: {
    visible(newVal) {
      if (newVal) {
        // 延迟初始化图表，确保模态框已完全显示
        setTimeout(() => {
          this.initChart();
        }, 300);
      } else {
        this.disposeChart();
      }
    },
    gid(newVal) {
      if (this.visible && newVal) {
        // 延迟初始化图表，确保模态框已完全显示
        setTimeout(() => {
          this.initChart();
        }, 300);
      }
    }
  },
  mounted() {
    if (this.visible) {
      // 延迟初始化图表，确保模态框已完全显示
      setTimeout(() => {
        this.initChart();
      }, 300);
    }
    
    window.addEventListener('resize', this.handleResize);
  },
  beforeUnmount() {
    this.disposeChart();
    window.removeEventListener('resize', this.handleResize);
  },
  methods: {
    async initChart() {
      console.log('开始初始化图表');
      this.loading = true;
      this.error = null;
      this.hasData = false;
      
      try {
        // 加载数据
        await this.fetchGraphData();
        
        // 检查数据是否有效
        if (!this.graphData || !this.graphData.nodes || this.graphData.nodes.length === 0) {
          console.warn('图表数据无效或为空');
          this.hasData = false;
          this.loading = false;
          return;
        }
        
        // 尝试转换数据格式
        this.tryConvertGraphData();
        
        // 标记数据已加载
        this.hasData = true;
        
        // 确保 DOM 已更新
        this.$nextTick(() => {
          // 使用更长的延迟，确保模态框完全显示并且 DOM 已渲染
          setTimeout(() => {
            console.log('DOM 已更新，开始创建图表');
            
            // 确保图表容器可见
            const chartContainer = this.$refs.chartContainer;
            if (chartContainer) {
              chartContainer.style.display = 'block';
              
              // 延迟一小段时间，确保样式已应用
              setTimeout(() => {
                this.createChart();
                this.loading = false;
              }, 100);
            } else {
              console.error('图表容器不存在，无法创建图表');
              this.error = '无法创建图表：容器元素不存在';
              this.loading = false;
            }
          }, 500);
        });
      } catch (err) {
        console.error('加载调用图数据失败:', err);
        this.error = `加载调用图数据失败: ${err.message}`;
        this.loading = false;
        this.$emit('error', this.error);
      }
    },
    
    async fetchGraphData() {
      if (this.useMockData) {
        console.log('使用模拟数据');
        this.graphData = this.getMockGraphData();
        return;
      }
      
      try {
        console.log(`获取GID ${this.gid}的调用图数据，数据库路径: ${this.dbpath}`);
        
        // 后端API应该返回以下格式的数据:
        // {
        //   nodes: [
        //     {
        //       id: "函数唯一标识符",
        //       name: "完整函数名称",
        //       package: "包路径",
        //       callCount: 调用次数,
        //       timeCost: "执行时间，如10ms",
        //       category: 分类编号
        //     },
        //     ...
        //   ],
        //   edges: [
        //     {
        //       source: "源节点ID",
        //       target: "目标节点ID",
        //       value: 权重值,
        //       callCount: 调用次数,
        //       timeCost: "执行时间"
        //     },
        //     ...
        //   ]
        // }
        const response = await axios.post('/api/runtime/traces/graph', {
          gid: this.gid,
          dbpath: this.dbpath
        });
        
        // 检查响应数据格式
        if (!response.data) {
          throw new Error('API返回数据为空');
        }
        
        // 保存原始响应数据，用于调试
        this.originalApiData = JSON.parse(JSON.stringify(response.data));
        console.log('API原始响应数据:', JSON.stringify(this.originalApiData));
        
        // 尝试解析和转换数据
        let data = this.parseResponseData(response.data);
        
        // 确保数据格式正确
        if (!data.nodes || !Array.isArray(data.nodes)) {
          console.error('API返回的节点数据格式不正确:', data);
          data.nodes = [];
        }
        
        if (!data.edges || !Array.isArray(data.edges)) {
          console.error('API返回的边数据格式不正确:', data);
          data.edges = [];
        }
        
        // 确保节点和边的ID匹配
        if (this.autoFixData) {
          this.validateAndFixGraphData(data);
        }
        
        this.graphData = data;
        
        // 记录数据格式信息
        this.updateDataFormatInfo();
        
        console.log('处理后的图表数据:', this.graphData);
      } catch (error) {
        console.error('获取调用图数据失败:', error);
        // 不自动使用模拟数据，而是显示错误
        throw error;
      }
    },
    
    parseResponseData(responseData) {
      // 创建一个深拷贝，避免修改原始数据
      const data = JSON.parse(JSON.stringify(responseData));
      
      // 如果已经是正确的格式，直接返回
      if (data.nodes && Array.isArray(data.nodes) && data.edges && Array.isArray(data.edges)) {
        return data;
      }
      
      console.log('尝试解析响应数据格式');
      
      // 检查是否是特定的API格式
      if (data.nodes === undefined && data.edges === undefined) {
        // 尝试解析特定的API格式
        return this.parseSpecificApiFormat(data);
      }
      
      // 检查常见的数据结构
      const possibleDataPaths = [
        'data', 'graph', 'result', 'body', 'response', 'content', 'value'
      ];
      
      // 遍历可能的路径
      for (const path of possibleDataPaths) {
        if (data[path]) {
          if (data[path].nodes && Array.isArray(data[path].nodes) && 
              data[path].edges && Array.isArray(data[path].edges)) {
            console.log(`从 ${path} 字段中提取节点和边`);
            return data[path];
          }
        }
      }
      
      // 检查是否有特殊格式的数据
      if (data.nodes && typeof data.nodes === 'object' && !Array.isArray(data.nodes)) {
        // 可能是对象形式的节点集合，转换为数组
        const nodesArray = Object.keys(data.nodes).map(key => {
          const node = data.nodes[key];
          return {
            id: key,
            ...node
          };
        });
        data.nodes = nodesArray;
      }
      
      if (data.edges && typeof data.edges === 'object' && !Array.isArray(data.edges)) {
        // 可能是对象形式的边集合，转换为数组
        const edgesArray = Object.keys(data.edges).map(key => {
          const edge = data.edges[key];
          return {
            id: key,
            ...edge
          };
        });
        data.edges = edgesArray;
      }
      
      // 检查是否有其他格式的数据结构
      if (Array.isArray(data)) {
        // 可能是节点数组，尝试构建边
        console.log('检测到数组格式的数据，尝试构建图结构');
        const nodes = data.map((item, index) => ({
          id: item.id || `node_${index}`,
          name: item.name || item.label || item.id || `Node ${index}`,
          ...item
        }));
        
        // 尝试从节点中提取边信息
        const edges = [];
        nodes.forEach(node => {
          if (node.children && Array.isArray(node.children)) {
            node.children.forEach(childId => {
              edges.push({
                source: node.id,
                target: childId,
                value: 1
              });
            });
          }
          
          if (node.parents && Array.isArray(node.parents)) {
            node.parents.forEach(parentId => {
              edges.push({
                source: parentId,
                target: node.id,
                value: 1
              });
            });
          }
          
          if (node.calls && Array.isArray(node.calls)) {
            node.calls.forEach(call => {
              const targetId = typeof call === 'string' ? call : call.id || call.target;
              if (targetId) {
                edges.push({
                  source: node.id,
                  target: targetId,
                  value: call.value || call.count || 1
                });
              }
            });
          }
        });
        
        return { nodes, edges };
      }
      
      // 如果没有找到合适的数据结构，创建空的数据结构
      return { 
        nodes: data.nodes || [], 
        edges: data.edges || [] 
      };
    },
    
    parseSpecificApiFormat(data) {
      console.log('尝试解析特定API格式的数据');
      
      // 检查是否有函数调用关系数据
      if (Array.isArray(data)) {
        console.log('检测到数组格式的API数据');
        
        // 创建节点和边的集合
        const nodesMap = new Map();
        const edges = [];
        
        // 遍历数据，提取节点和边
        data.forEach((item) => {
          // 检查是否有函数名
          if (item.func || item.function || item.name) {
            const funcName = item.func || item.function || item.name;
            
            // 如果节点不存在，创建节点
            if (!nodesMap.has(funcName)) {
              nodesMap.set(funcName, {
                id: funcName,
                name: funcName,
                callCount: 0,
                executionTime: '0ms',
                category: 0
              });
            }
            
            // 更新节点信息
            const node = nodesMap.get(funcName);
            node.callCount = (node.callCount || 0) + 1;
            
            if (item.time || item.executionTime) {
              node.executionTime = item.time || item.executionTime;
            }
            
            // 检查是否有调用关系
            if (item.calls && Array.isArray(item.calls)) {
              item.calls.forEach(call => {
                const targetFunc = call.func || call.function || call.name || call;
                
                // 如果目标节点不存在，创建节点
                if (!nodesMap.has(targetFunc)) {
                  nodesMap.set(targetFunc, {
                    id: targetFunc,
                    name: targetFunc,
                    callCount: 0,
                    executionTime: '0ms',
                    category: 0
                  });
                }
                
                // 创建边
                edges.push({
                  source: funcName,
                  target: targetFunc,
                  value: call.count || 1,
                  callCount: call.count || 1
                });
              });
            }
            
            // 检查是否有调用者
            if (item.calledBy && Array.isArray(item.calledBy)) {
              item.calledBy.forEach(caller => {
                const callerFunc = caller.func || caller.function || caller.name || caller;
                
                // 如果调用者节点不存在，创建节点
                if (!nodesMap.has(callerFunc)) {
                  nodesMap.set(callerFunc, {
                    id: callerFunc,
                    name: callerFunc,
                    callCount: 0,
                    executionTime: '0ms',
                    category: 0
                  });
                }
                
                // 创建边
                edges.push({
                  source: callerFunc,
                  target: funcName,
                  value: caller.count || 1,
                  callCount: caller.count || 1
                });
              });
            }
          }
        });
        
        // 将节点集合转换为数组
        const nodes = Array.from(nodesMap.values());
        
        console.log(`从API数据中提取了 ${nodes.length} 个节点和 ${edges.length} 条边`);
        return { nodes, edges };
      }
      
      // 检查是否有特定格式的对象
      if (typeof data === 'object' && data !== null) {
        // 检查是否有函数调用图数据
        if (data.functions || data.calls || data.trace) {
          const traceData = data.functions || data.calls || data.trace;
          
          if (Array.isArray(traceData)) {
            return this.parseSpecificApiFormat(traceData);
          }
        }
      }
      
      // 如果无法解析，返回空数据
      return { nodes: [], edges: [] };
    },
    
    validateAndFixGraphData(data) {
      // 确保所有节点都有唯一ID
      const nodeIds = new Set();
      data.nodes.forEach((node, index) => {
        if (!node.id) {
          node.id = `node_${index}`;
        }
        
        // 确保ID是字符串类型
        node.id = String(node.id);
        nodeIds.add(node.id);
        
        // 确保节点有名称
        if (!node.name) {
          node.name = node.label || node.id || `Node ${index}`;
        }
        
        // 尝试解析包路径
        if (!node.package && node.name) {
          node.package = this.extractPackagePath(node.name);
        }
        
        // 确保有执行时间
        if (!node.executionTime && !node.timeCost) {
          node.timeCost = '0ms';
        }
      });
      
      // 确保所有边的source和target都存在于节点中，并且是字符串类型
      const validEdges = data.edges.filter(edge => {
        // 确保source和target是字符串类型
        if (edge.source) edge.source = String(edge.source);
        if (edge.target) edge.target = String(edge.target);
        
        // 检查source和target是否存在
        const sourceExists = edge.source && nodeIds.has(edge.source);
        const targetExists = edge.target && nodeIds.has(edge.target);
        
        if (!sourceExists || !targetExists) {
          console.warn(`边 ${edge.source} -> ${edge.target} 的节点不存在，将被过滤`);
        }
        
        return sourceExists && targetExists;
      });
      
      // 更新边数据
      data.edges = validEdges;
      
      // 如果没有边，但有节点，尝试创建一些边以便显示
      if (data.edges.length === 0 && data.nodes.length > 0) {
        console.warn('没有有效的边数据，尝试创建一些边以便显示');
        
        // 创建一个简单的链式结构
        for (let i = 1; i < data.nodes.length; i++) {
          data.edges.push({
            source: data.nodes[i-1].id,
            target: data.nodes[i].id,
            value: 1
          });
        }
      }
      
      return data;
    },
    
    extractPackagePath(fullName) {
      if (!fullName || typeof fullName !== 'string') {
        return '';
      }
      
      // 尝试从完整函数名中提取包路径
      const lastDotIndex = fullName.lastIndexOf('.');
      if (lastDotIndex > 0) {
        return fullName.substring(0, lastDotIndex);
      }
      
      return '';
    },
    
    createChart() {
      // 清理旧的图表实例
      if (this.chartInstance) {
        try {
          console.log('清理旧图表实例');
          this.chartInstance.dispose();
        } catch (e) {
          console.warn('清理旧图表实例时出错:', e);
        }
        this.chartInstance = null;
      }
      
      // 确保数据有效
      if (!this.graphData || !this.graphData.nodes || this.graphData.nodes.length === 0) {
        console.warn('图表数据无效或为空');
        this.hasData = false;
        this.error = '没有有效的图表数据';
        this.loading = false;
        return;
      }
      
      console.log(`准备创建${this.currentLayout}布局图表，节点数: ${this.graphData.nodes.length}, 边数: ${this.graphData.edges.length}`);
      
      // 获取图表容器
      const chartContainer = this.$refs.chartContainer;
      if (!chartContainer) {
        console.error('图表容器不存在');
        this.error = '无法创建图表：容器元素不存在';
        this.loading = false;
        
        // 延迟重试
        setTimeout(() => {
          console.log('延迟重试创建图表...');
          if (this.$refs.chartContainer) {
            this.createChart();
          }
        }, 1000);
        return;
      }
      
      // 确保容器尺寸合适
      const width = chartContainer.clientWidth;
      const height = chartContainer.clientHeight;
      
      if (width <= 0 || height <= 0) {
        console.warn(`图表容器尺寸异常: 宽度=${width}, 高度=${height}`);
        
        // 设置最小尺寸
        chartContainer.style.width = '100%';
        chartContainer.style.height = '600px';
        chartContainer.style.display = 'block';
        
        // 延迟重试，等待样式应用
        setTimeout(() => {
          console.log('容器尺寸已调整，重试创建图表...');
          this.createChart();
        }, 500);
        return;
      }
      
      try {
        console.log(`创建图表，容器尺寸: 宽度=${width}, 高度=${height}, 布局: ${this.currentLayout}`);
        
        // 创建图表实例
        this.chartInstance = echarts.init(chartContainer);
        
        // 设置图表选项
        const option = this.getChartOption();
        console.log('图表配置:', JSON.stringify(option.series[0].force));
        this.chartInstance.setOption(option);
        
        // 添加点击事件
        this.chartInstance.on('click', 'series.graph.nodes', (params) => {
          console.log('点击节点:', params.data);
        });
        
        // 添加拖拽结束事件
        this.chartInstance.on('dragend', 'series.graph.nodes', this.handleDragEnd);
        
        // 设置图表加载完成
        this.hasData = true;
        this.loading = false;
        
        console.log('图表创建成功');
      } catch (error) {
        console.error('创建图表失败:', error);
        this.error = `创建图表失败: ${error.message}`;
        this.loading = false;
        
        // 如果是因为容器问题，尝试延迟重试
        if (error.message.includes('container') || error.message.includes('DOM')) {
          setTimeout(() => {
            console.log('因容器问题创建失败，延迟重试...');
            this.createChart();
          }, 1000);
        }
      }
    },
    
    getChartOption() {
      // 处理节点和边的数据
      const nodes = this.graphData.nodes.map(node => ({
        id: node.id,
        name: this.formatFunctionName(node.name || node.id),
        value: node.value || 1,
        symbolSize: this.calculateNodeSize(node),
        category: node.category || 0,
        // 如果节点有保存的位置，则使用该位置
        x: node.x,
        y: node.y,
        label: {
          show: true
        },
        itemStyle: {
          color: this.getNodeColor(node)
        },
        tooltip: {
          formatter: () => {
            return `<div>
              <div><strong>函数:</strong> ${node.name || node.id}</div>
              <div><strong>包路径:</strong> ${node.package || '未知'}</div>
              <div><strong>调用次数:</strong> ${node.callCount || 1}</div>
              <div><strong>执行时间:</strong> ${node.executionTime || node.timeCost || '未知'}</div>
            </div>`;
          }
        },
        // 原始数据
        rawData: node
      }));
      
      const edges = this.graphData.edges.map(edge => ({
        source: edge.source,
        target: edge.target,
        value: edge.value || 1,
        lineStyle: {
          width: Math.max(1, Math.min(5, edge.value || 1)),
          curveness: 0.2
        },
        // 添加箭头
        symbol: ['none', 'arrow'],
        symbolSize: [5, 10],
        tooltip: {
          formatter: () => {
            return `<div>
              <div><strong>调用次数:</strong> ${edge.callCount || 1}</div>
              <div><strong>执行时间:</strong> ${edge.timeCost || '未知'}</div>
            </div>`;
          }
        }
      }));
      
      // 根据不同布局返回不同的配置
      switch (this.currentLayout) {
        case 'circular':
          return this.getCircularLayoutOption(nodes, edges);
        case 'tree':
          return this.getTreeLayoutOption(nodes, edges);
        case 'force':
        default:
          return this.getForceLayoutOption(nodes, edges);
      }
    },
    
    getForceLayoutOption(nodes, edges) {
      // 确保有节点数据
      if (!nodes || nodes.length === 0) {
        return {
          title: {
            text: '函数调用关系图 - 无数据',
            subtext: `GID: ${this.gid}`,
            left: 'center'
          }
        };
      }
      
      console.log('创建力导向图配置，节点数:', nodes.length, '边数:', edges.length);
      
      // 使用更简单的力导向图配置
      return {
        title: {
          text: '函数调用关系图',
          subtext: `GID: ${this.gid} - 节点数: ${nodes.length}, 边数: ${edges.length}`,
          left: 'center'
        },
        tooltip: {
          trigger: 'item',
          confine: true
        },
        // 添加图表右上角的提示信息
        graphic: [
          {
            type: 'text',
            right: 20,
            top: 20,
            style: {
              text: '提示: 可拖拽节点调整位置',
              fontSize: 12,
              fill: '#999'
            }
          }
        ],
        series: [{
          name: '函数调用',
          type: 'graph',
          layout: 'force',
          data: nodes,
          links: edges,
          roam: true,
          draggable: true,  // 允许节点拖拽
          label: {
            show: true,
            position: 'right',
            formatter: '{b}'
          },
          // 确保边有箭头
          edgeSymbol: ['none', 'arrow'],
          edgeSymbolSize: [0, 8],
          lineStyle: {
            color: '#bbb',
            curveness: 0.3,
            width: 1.5
          },
          force: {
            repulsion: 100,
            edgeLength: 80
          }
        }]
      };
    },
    
    getCircularLayoutOption(nodes, edges) {
      // 确保有节点数据
      if (!nodes || nodes.length === 0) {
        return {
          title: {
            text: '函数调用关系图 (环形布局) - 无数据',
            subtext: `GID: ${this.gid}`,
            left: 'center'
          }
        };
      }
      
      return {
        title: {
          text: '函数调用关系图 (环形布局)',
          subtext: `GID: ${this.gid} - 节点数: ${nodes.length}, 边数: ${edges.length}`,
          left: 'center'
        },
        tooltip: {
          trigger: 'item',
          confine: true
        },
        // 添加图表右上角的提示信息
        graphic: [
          {
            type: 'text',
            right: 20,
            top: 20,
            style: {
              text: '提示: 可拖拽节点调整位置',
              fontSize: 12,
              fill: '#999'
            }
          }
        ],
        animationDuration: 1500,
        animationEasingUpdate: 'quinticInOut',
        series: [{
          name: '函数调用',
          type: 'graph',
          layout: 'circular',
          circular: {
            rotateLabel: true
          },
          data: nodes,
          links: edges,
          roam: true,
          draggable: true,  // 允许节点拖拽
          label: {
            show: true,
            position: 'right',
            formatter: '{b}'
          },
          // 确保边有箭头
          edgeSymbol: ['none', 'arrow'],
          edgeSymbolSize: [0, 8],
          lineStyle: {
            color: '#bbb',
            curveness: 0.1,
            width: 1.5
          },
          emphasis: {
            focus: 'adjacency',
            lineStyle: {
              width: 4
            }
          }
        }]
      };
    },
    
    getTreeLayoutOption(nodes, edges) {
      // 确保有节点数据
      if (!nodes || nodes.length === 0) {
        return {
          title: {
            text: '函数调用关系图 (树形布局) - 无数据',
            subtext: `GID: ${this.gid}`,
            left: 'center'
          }
        };
      }
      
      // 为树形布局找到根节点
      const rootNode = this.findRootNode(nodes, edges);
      
      return {
        title: {
          text: '函数调用关系图 (树形布局)',
          subtext: `GID: ${this.gid} - 根节点: ${rootNode ? rootNode.name : '未知'} - 节点数: ${nodes.length}`,
          left: 'center'
        },
        tooltip: {
          trigger: 'item',
          confine: true
        },
        // 添加图表右上角的提示信息
        graphic: [
          {
            type: 'text',
            right: 20,
            top: 20,
            style: {
              text: '提示: 可拖拽节点调整位置',
              fontSize: 12,
              fill: '#999'
            }
          }
        ],
        animationDuration: 1500,
        animationEasingUpdate: 'quinticInOut',
        series: [{
          name: '函数调用',
          type: 'graph',
          layout: 'force',
          data: nodes,
          links: edges,
          roam: true,
          draggable: true,  // 允许节点拖拽
          label: {
            show: true,
            position: 'right',
            formatter: '{b}'
          },
          // 确保边有箭头
          edgeSymbol: ['none', 'arrow'],
          edgeSymbolSize: [0, 8],
          lineStyle: {
            color: '#bbb',
            curveness: 0.1,
            width: 1.5
          },
          emphasis: {
            focus: 'adjacency',
            lineStyle: {
              width: 4
            }
          },
          force: {
            initLayout: 'circular',
            gravity: 0.1,
            repulsion: 150,
            edgeLength: 80
          }
        }]
      };
    },
    
    findRootNode(nodes, edges) {
      // 找出入度为0的节点作为根节点
      const targetIds = new Set(edges.map(edge => edge.target));
      const rootNodes = nodes.filter(node => !targetIds.has(node.id));
      
      return rootNodes.length > 0 ? rootNodes[0] : nodes[0];
    },
    
    formatFunctionName(name) {
      if (!name) return 'unknown';
      
      // 从完整路径中提取包名和函数名
      // 例如：github.com/nsqio/nsq/nsqd.(*Topic).PutMessage -> nsqd.(*Topic).PutMessage
      // 例如：github.com/nsqio/nsq/internal/http_api.RespondV1 -> http_api.RespondV1
      
      // 查找最后一个斜杠位置
      const lastSlashIndex = name.lastIndexOf('/');
      if (lastSlashIndex >= 0) {
        // 从斜杠后开始截取
        const packageAndFunc = name.substring(lastSlashIndex + 1);
        return packageAndFunc;
      }
      
      // 如果没有斜杠，直接返回原名称
      return name;
    },
    
    calculateNodeSize(node) {
      // 根据调用次数计算节点大小
      const baseSize = 10;
      const callCount = node.callCount || 1;
      
      return Math.max(baseSize, Math.min(baseSize * 3, baseSize + Math.log2(callCount) * 5));
    },
    
    getNodeColor(node) {
      // 根据节点类型或其他属性设置颜色
      const colors = [
        '#5470c6', '#91cc75', '#fac858', '#ee6666',
        '#73c0de', '#3ba272', '#fc8452', '#9a60b4'
      ];
      
      const category = node.category || 0;
      return colors[category % colors.length];
    },
    
    resizeChart() {
      if (this.chartInstance) {
        try {
          console.log('调整图表大小');
          this.chartInstance.resize();
        } catch (e) {
          console.warn('调整图表大小时出错:', e);
        }
      }
    },
    
    handleResize() {
      // 使用防抖，避免频繁调整大小
      if (this.resizeTimer) {
        clearTimeout(this.resizeTimer);
      }
      
      this.resizeTimer = setTimeout(() => {
        if (this.visible && this.chartInstance) {
          this.resizeChart();
        }
      }, 200);
    },
    
    disposeChart() {
      // 清理图表实例
      if (this.chartInstance) {
        try {
          console.log('销毁图表实例');
          this.chartInstance.dispose();
        } catch (e) {
          console.warn('销毁图表实例时出错:', e);
        }
        this.chartInstance = null;
      }
      
      // 清理定时器
      if (this.resizeTimer) {
        clearTimeout(this.resizeTimer);
        this.resizeTimer = null;
      }
    },
    
    closeModal() {
      this.$emit('update:visible', false);
    },
    
    zoomIn() {
      this.zoomLevel *= 1.2;
      this.applyZoom();
    },
    
    zoomOut() {
      this.zoomLevel /= 1.2;
      this.applyZoom();
    },
    
    resetZoom() {
      this.zoomLevel = 1;
      this.applyZoom();
    },
    
    applyZoom() {
      if (this.chartInstance) {
        const option = this.chartInstance.getOption();
        
        // 应用缩放级别
        if (option.series && option.series[0]) {
          // 更新缩放相关配置
          this.chartInstance.setOption({
            series: [{
              zoom: this.zoomLevel,
              roam: true
            }]
          });
        }
      }
    },
    
    changeLayout(layout) {
      console.log(`切换布局: ${this.currentLayout} -> ${layout}`);
      this.currentLayout = layout;
      
      if (this.chartInstance) {
        // 先销毁当前图表实例
        try {
          this.chartInstance.dispose();
          this.chartInstance = null;
        } catch (e) {
          console.warn('切换布局时销毁图表实例出错:', e);
        }
        
        // 延迟创建新图表，确保DOM已更新
        this.$nextTick(() => {
          setTimeout(() => {
            this.createChart();
          }, 300);
        });
      }
    },
    
    getMockGraphData() {
      console.log('生成模拟数据');
      // 生成简单的模拟数据
      const nodes = [];
      const edges = [];
      
      // 创建一些简单的节点
      for (let i = 0; i < 10; i++) {
        nodes.push({
          id: `node${i}`,
          name: `github.com/toheart/goanalysis/pkg/module${i}.Function_${i}`,
          package: `github.com/toheart/goanalysis/pkg/module${i}`,
          value: 1,
          symbolSize: 20,
          callCount: Math.floor(Math.random() * 100) + 1,
          timeCost: `${Math.floor(Math.random() * 100)}ms`,
          category: Math.floor(i / 3)
        });
      }
      
      // 创建一些简单的边
      for (let i = 1; i < 10; i++) {
        edges.push({
          source: `node0`,
          target: `node${i}`,
          value: 1,
          callCount: Math.floor(Math.random() * 50) + 1,
          timeCost: `${Math.floor(Math.random() * 50)}ms`
        });
      }
      
      // 添加一些额外的连接
      for (let i = 1; i < 5; i++) {
        edges.push({
          source: `node${i}`,
          target: `node${i+5}`,
          value: 1,
          callCount: Math.floor(Math.random() * 30) + 1,
          timeCost: `${Math.floor(Math.random() * 30)}ms`
        });
      }
      
      console.log('生成的模拟数据:', { nodes, edges });
      return { nodes, edges };
    },
    
    tryConvertGraphData() {
      // 如果数据格式不是预期的格式，尝试转换
      if (this.graphData) {
        console.log('尝试转换数据格式，当前数据:', this.graphData);
        
        // 检查是否有其他格式的数据
        if (!this.graphData.nodes && !this.graphData.edges) {
          // 尝试解析数据
          const parsedData = this.parseResponseData(this.graphData);
          if (parsedData.nodes || parsedData.edges) {
            console.log('成功解析数据格式:', parsedData);
            this.graphData = parsedData;
          }
        }
        
        // 确保有节点和边数组
        if (!this.graphData.nodes) {
          this.graphData.nodes = [];
        }
        
        if (!this.graphData.edges) {
          this.graphData.edges = [];
        }
        
        // 再次验证和修复数据
        this.validateAndFixGraphData(this.graphData);
      }
    },
    
    toggleDebugInfo() {
      this.showDebugInfo = !this.showDebugInfo;
    },
    
    retryFetch() {
      this.initChart();
    },
    
    retryWithMockData() {
      this.useMockDataAndRefresh();
    },
    
    useMockDataAndRefresh() {
      console.log('使用模拟数据并刷新图表');
      this.graphData = this.getMockGraphData();
      this.error = null;
      this.hasData = true;
      this.loading = false;
      
      // 确保图表容器可见
      const chartContainer = this.$refs.chartContainer;
      if (chartContainer) {
        chartContainer.style.display = 'block';
      }
      
      // 使用 nextTick 确保 DOM 已更新
      this.$nextTick(() => {
        // 添加延迟，确保模态框完全显示并且 DOM 已渲染
        setTimeout(() => {
          console.log('使用模拟数据创建图表');
          this.createChart();
        }, 500);
      });
    },
    
    refreshChart() {
      if (this.chartInstance) {
        this.chartInstance.dispose();
      }
      this.createChart();
    },
    
    viewRawApiData() {
      if (this.showDebugInfo) {
        // 简单地在控制台输出数据，避免HTML生成的问题
        console.log('原始API数据:', this.originalApiData);
        console.log('处理后的图表数据:', this.graphData);
        
        // 显示一个简单的提示
        alert('数据已输出到控制台，请按F12查看');
      }
    },
    
    updateDataFormatInfo() {
      if (!this.graphData) {
        this.dataFormatInfo = '无数据';
        return;
      }
      
      const nodeCount = this.graphData.nodes ? this.graphData.nodes.length : 0;
      const edgeCount = this.graphData.edges ? this.graphData.edges.length : 0;
      
      let info = `数据格式: 节点数=${nodeCount}, 边数=${edgeCount}`;
      
      // 检查节点格式
      if (nodeCount > 0) {
        const sampleNode = this.graphData.nodes[0];
        const nodeKeys = Object.keys(sampleNode).join(', ');
        info += `\n节点字段: ${nodeKeys}`;
      }
      
      // 检查边格式
      if (edgeCount > 0) {
        const sampleEdge = this.graphData.edges[0];
        const edgeKeys = Object.keys(sampleEdge).join(', ');
        info += `\n边字段: ${edgeKeys}`;
      }
      
      this.dataFormatInfo = info;
    },
    
    tryAlternativeFormat() {
      if (!this.originalApiData) {
        alert('没有原始API数据可供尝试');
        return;
      }
      
      this.alternativeFormatIndex = (this.alternativeFormatIndex + 1) % 4;
      
      let data;
      switch (this.alternativeFormatIndex) {
        case 0:
          // 尝试直接使用原始数据
          data = this.parseResponseData(this.originalApiData);
          break;
        case 1:
          // 尝试将原始数据视为节点数组
          if (Array.isArray(this.originalApiData)) {
            data = this.convertArrayToGraph(this.originalApiData);
          } else {
            data = { nodes: [], edges: [] };
          }
          break;
        case 2:
          // 尝试从原始数据中提取特定字段
          data = this.extractGraphFromObject(this.originalApiData);
          break;
        case 3:
          // 尝试使用特定的API格式解析
          data = this.parseSpecificApiFormat(this.originalApiData);
          break;
      }
      
      // 确保数据格式正确
      if (!data.nodes || !Array.isArray(data.nodes)) {
        data.nodes = [];
      }
      
      if (!data.edges || !Array.isArray(data.edges)) {
        data.edges = [];
      }
      
      // 确保节点和边的ID匹配
      if (this.autoFixData) {
        this.validateAndFixGraphData(data);
      }
      
      this.graphData = data;
      
      // 更新数据格式信息
      this.updateDataFormatInfo();
      
      // 刷新图表
      this.refreshChart();
    },
    
    convertArrayToGraph(array) {
      if (!Array.isArray(array)) {
        return { nodes: [], edges: [] };
      }
      
      // 创建节点
      const nodes = array.map((item, index) => {
        return {
          id: item.id || item.name || `node_${index}`,
          name: item.name || item.id || `Node ${index}`,
          category: 0,
          value: 1,
          ...item
        };
      });
      
      // 创建边
      const edges = [];
      for (let i = 1; i < nodes.length; i++) {
        edges.push({
          source: nodes[i-1].id,
          target: nodes[i].id,
          value: 1
        });
      }
      
      return { nodes, edges };
    },
    
    extractGraphFromObject(obj) {
      if (!obj || typeof obj !== 'object') {
        return { nodes: [], edges: [] };
      }
      
      // 尝试从对象中提取节点和边
      let nodes = [];
      let edges = [];
      
      // 检查常见的字段名
      const nodeFields = ['nodes', 'vertices', 'points', 'functions', 'calls'];
      const edgeFields = ['edges', 'links', 'connections', 'relations', 'calls'];
      
      // 尝试提取节点
      for (const field of nodeFields) {
        if (obj[field] && (Array.isArray(obj[field]) || typeof obj[field] === 'object')) {
          if (Array.isArray(obj[field])) {
            nodes = obj[field].map((node, index) => {
              return {
                id: node.id || `node_${index}`,
                name: node.name || node.label || node.id || `Node ${index}`,
                category: 0,
                value: 1,
                ...node
              };
            });
          } else {
            // 对象形式的节点集合
            nodes = Object.keys(obj[field]).map((key) => {
              const node = obj[field][key];
              return {
                id: key,
                name: node.name || node.label || key,
                category: 0,
                value: 1,
                ...node
              };
            });
          }
          break;
        }
      }
      
      // 尝试提取边
      for (const field of edgeFields) {
        if (obj[field] && (Array.isArray(obj[field]) || typeof obj[field] === 'object')) {
          if (Array.isArray(obj[field])) {
            edges = obj[field].map(edge => {
              return {
                source: edge.source || edge.from || edge.start,
                target: edge.target || edge.to || edge.end,
                value: edge.value || edge.weight || 1,
                ...edge
              };
            });
          } else {
            // 对象形式的边集合
            edges = Object.keys(obj[field]).map(key => {
              const edge = obj[field][key];
              return {
                source: edge.source || edge.from || edge.start,
                target: edge.target || edge.to || edge.end,
                value: edge.value || edge.weight || 1,
                ...edge
              };
            });
          }
          break;
        }
      }
      
      return { nodes, edges };
    },
    
    checkChartContainer() {
      // 检查图表容器是否存在
      if (!this.$refs.chartContainer) {
        console.warn('图表容器不存在，可能是DOM还未渲染完成');
        return false;
      }
      
      // 检查容器尺寸
      const container = this.$refs.chartContainer;
      const width = container.clientWidth;
      const height = container.clientHeight;
      
      if (width === 0 || height === 0) {
        console.warn(`图表容器尺寸异常: 宽度=${width}, 高度=${height}`);
        return false;
      }
      
      return true;
    },
    
    // 添加拖拽结束事件处理
    handleDragEnd(params) {
      console.log('节点拖拽结束:', params.dataIndex, params.data);
      
      // 获取被拖拽的节点
      const nodeIndex = params.dataIndex;
      const nodeData = params.data;
      
      // 更新原始数据中的节点位置，但不固定节点
      if (this.graphData && this.graphData.nodes && this.graphData.nodes[nodeIndex]) {
        // 保存节点的新位置
        this.graphData.nodes[nodeIndex].x = nodeData.x;
        this.graphData.nodes[nodeIndex].y = nodeData.y;
        // 不再固定节点位置，让节点可以继续参与力导向布局
        // this.graphData.nodes[nodeIndex].fixed = true;
        
        console.log(`节点 ${nodeData.name} 位置已保存: x=${nodeData.x}, y=${nodeData.y}`);
      }
    }
  }
};
</script>

<style scoped>
.function-call-graph {
  position: relative;
}

.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1050;
}

.modal-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1051;
  padding: 20px;
}

.modal-content {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 1200px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid #dee2e6;
}

.modal-body {
  flex: 1;
  overflow: auto;
  padding: 1rem;
  min-height: 400px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-top: 1px solid #dee2e6;
}

.chart-container {
  width: 100%;
  height: 600px;
}

.chart-controls {
  display: flex;
  align-items: center;
}

.debug-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  background-color: #f8f9fa;
  border-top: 1px solid #dee2e6;
  z-index: 1060;
  max-height: 200px;
  overflow-y: auto;
}

.debug-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid #dee2e6;
  background-color: #e9ecef;
}

.debug-body {
  padding: 1rem;
}

.btn-group .btn.active {
  background-color: #0d6efd;
  color: white;
}

@media (max-width: 768px) {
  .modal-content {
    width: 95%;
    max-height: 95vh;
  }
  
  .chart-container {
    height: 400px;
  }
  
  .modal-footer {
    flex-direction: column;
    gap: 10px;
  }
  
  .chart-controls {
    width: 100%;
    justify-content: center;
    margin-bottom: 10px;
  }
}
</style> 