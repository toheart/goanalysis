<template>
  <div class="function-upstream-analysis">
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-diagram-2 me-2"></i>函数上下游分析</h5>
      </div>
      <div class="card-body">
        <!-- 函数搜索 -->
        <div class="row mb-4">
          <div class="col-md-8">
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input 
                type="text" 
                class="form-control" 
                placeholder="输入函数名称进行搜索" 
                v-model="searchQuery"
                @input="debouncedSearch"
                :disabled="searching"
                ref="searchInput"
              >
              <button 
                class="btn btn-primary" 
                @click="searchFunctions" 
                :disabled="!searchQuery || searching"
              >
                <span v-if="searching" class="spinner-border spinner-border-sm me-2" role="status"></span>
                搜索
              </button>
            </div>
            <div v-if="searchResults.length > 0" class="search-results mt-2">
              <div class="list-group">
                <button 
                  v-for="func in searchResults" 
                  :key="func.key" 
                  class="list-group-item list-group-item-action"
                  @click="selectFunction(func)"
                >
                  <div class="d-flex justify-content-between align-items-center">
                    <div>
                      <strong>{{ func.name }}</strong>
                      <small class="d-block text-muted">{{ func.package }}</small>
                    </div>
                    <span class="badge bg-primary rounded-pill">{{ func.callCount || 0 }}</span>
                  </div>
                </button>
              </div>
            </div>
            <div v-else-if="searched && !searching" class="alert alert-warning mt-2">
              <i class="bi bi-exclamation-triangle-fill me-2"></i>未找到匹配的函数
            </div>
          </div>
          <div class="col-md-4">
            <div class="alert alert-info mb-0">
              <h6><i class="bi bi-info-circle me-2"></i>提示</h6>
              <p class="mb-0 small">输入函数名称进行模糊搜索，选择函数后将展示其上游调用关系。</p>
            </div>
          </div>
        </div>
        
        <!-- 选中的函数信息 -->
        <div v-if="selectedFunction" class="selected-function mb-4">
          <div class="card">
            <div class="card-header bg-primary text-white">
              <h6 class="mb-0">已选择的函数</h6>
            </div>
            <div class="card-body">
              <div class="row">
                <div class="col-md-6">
                  <p><strong>函数名称:</strong> {{ selectedFunction.name }}</p>
                  <p><strong>包路径:</strong> {{ selectedFunction.package }}</p>
                </div>
                <div class="col-md-6">
                  <p><strong>被调用次数:</strong> {{ selectedFunction.callCount || 0 }}</p>
                  <p><strong>函数ID:</strong> {{ selectedFunction.key }}</p>
                </div>
              </div>
              <div class="d-flex justify-content-end">
                <div class="btn-group">
                  <button class="btn btn-primary" @click="analyzeUpstream" :disabled="analyzing">
                    <span v-if="analyzing && analysisType === 'upstream'" class="spinner-border spinner-border-sm me-2" role="status"></span>
                    <i v-else class="bi bi-arrow-up me-2"></i>分析上游调用
                  </button>
                  <button class="btn btn-success" @click="analyzeDownstream" :disabled="analyzing">
                    <span v-if="analyzing && analysisType === 'downstream'" class="spinner-border spinner-border-sm me-2" role="status"></span>
                    <i v-else class="bi bi-arrow-down me-2"></i>分析下游调用
                  </button>
                  <button class="btn btn-info" @click="analyzeFullChain" :disabled="analyzing">
                    <span v-if="analyzing && analysisType === 'fullchain'" class="spinner-border spinner-border-sm me-2" role="status"></span>
                    <i v-else class="bi bi-diagram-3 me-2"></i>分析全链路
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 图形展示区域 -->
        <div v-if="graphData.nodes.length > 0" class="graph-container">
          <div class="card">
            <div class="card-header">
              <h6 class="mb-0">
                <span v-if="analysisType === 'upstream'">函数上游调用关系图 (箭头方向: 调用者 → 被调用者)</span>
                <span v-else-if="analysisType === 'downstream'">函数下游调用关系图 (箭头方向: 调用者 → 被调用者)</span>
                <span v-else-if="analysisType === 'fullchain'">函数全链路调用关系图 (箭头方向: 调用者 → 被调用者)</span>
              </h6>
              <small class="text-muted">提示：可以使用鼠标滚轮缩放，按住鼠标左键拖动图表</small>
            </div>
            <div class="card-body p-0">
              <div id="function-graph" class="function-graph" ref="functionGraph"></div>
            </div>
          </div>
        </div>
        
        <!-- 无数据提示 -->
        <div v-else-if="analyzed && !analyzing" class="alert alert-warning">
          <i class="bi bi-exclamation-triangle-fill me-2"></i>未找到上游调用关系
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../../axios';
import * as echarts from 'echarts';

export default {
  name: 'FunctionCallAnalysis',
  props: {
    dbFilePath: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      searchQuery: '',
      searching: false,
      searched: false,
      searchResults: [],
      selectedFunction: null,
      analyzing: false,
      analyzed: false,
      analysisType: 'upstream', // 'upstream', 'downstream', 'fullchain'
      graphData: {
        nodes: [],
        links: []
      },
      chart: null,
      searchTimeout: null
    };
  },
  created() {
    // 创建防抖函数
    this.debouncedSearch = this.debounce(this.searchFunctions, 300);
  },
  mounted() {
    // 监听窗口大小变化，调整图表大小
    window.addEventListener('resize', this.resizeChart);
  },
  beforeUnmount() {
    // 组件销毁前移除事件监听
    window.removeEventListener('resize', this.resizeChart);
    // 销毁图表实例
    if (this.chart) {
      this.chart.dispose();
      this.chart = null;
    }
  },
  updated() {
    // 只在有图形数据且图表未初始化时初始化
    if (this.graphData.nodes.length > 0 && !this.chart && document.getElementById('function-graph')) {
      this.$nextTick(() => {
        this.initChart();
        this.renderGraph();
      });
    }
  },
  methods: {
    // 防抖函数
    debounce(fn, delay) {
      return function(...args) {
        if (this.searchTimeout) clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
          fn.apply(this, args);
        }, delay);
      };
    },
    
    async searchFunctions() {
      if (!this.searchQuery || this.searchQuery.length < 2) return;
      
      this.searching = true;
      this.searched = false;
      this.searchResults = [];
      
      try {
        const decodedPath = decodeURIComponent(this.dbFilePath);
        const response = await axios.post('/api/static/search-functions', {
          dbPath: decodedPath,
          query: this.searchQuery
        });
        
        this.searchResults = response.data.functions || [];
        this.searched = true;
      } catch (error) {
        console.error('搜索函数失败:', error);
      } finally {
        this.searching = false;
      }
    },
    
    selectFunction(func) {
      this.selectedFunction = func;
      this.searchResults = [];
      this.searchQuery = func.name;
      
      // 选择函数后自动执行分析
      this.analyzeUpstream();
      
      // 保持输入框焦点
      this.$nextTick(() => {
        if (this.$refs.searchInput) {
          this.$refs.searchInput.focus();
        }
      });
    },
    
    async analyzeUpstream() {
      if (!this.selectedFunction) return;
      
      this.analyzing = true;
      this.analysisType = 'upstream';
      
      try {
        const decodedPath = decodeURIComponent(this.dbFilePath);
        console.log('分析上游调用，请求参数:', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        const response = await axios.post('/api/static/function-upstream', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        console.log('分析上游调用，响应数据:', response.data);
        
        // 确保数据有效
        if (!response.data || !response.data.nodes || !response.data.edges) {
          console.error('服务器返回的数据格式不正确:', response.data);
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 处理数据并更新图形数据
        const newGraphData = this.processGraphData(response.data);
        console.log('处理后的图形数据:', newGraphData);
        
        // 确保有节点数据
        if (newGraphData.nodes.length === 0) {
          console.warn('没有节点数据，无法渲染图形');
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 成功获取数据后，更新图形数据
        this.graphData = newGraphData;
        this.analyzed = true;
        
        // 确保DOM已经更新后再初始化和渲染图形
        this.$nextTick(() => {
          // 如果图表已存在，先销毁
          if (this.chart) {
            this.chart.dispose();
            this.chart = null;
          }
          
          // 初始化并渲染图表
          this.initChart();
          this.renderGraph();
        });
      } catch (error) {
        console.error('分析上游调用失败:', error);
      } finally {
        this.analyzing = false;
      }
    },
    
    processGraphData(data) {
      // 处理后端返回的数据，转换为图表可用的格式
      if (!data || !data.nodes || !data.edges) {
        console.error('无效的图形数据:', data);
        return { nodes: [], links: [] };
      }
      
      const nodes = data.nodes.map(node => ({
        id: node.id,
        name: node.name,
        package: node.package,
        callCount: node.callCount || 0,
        group: node.id === this.selectedFunction.key ? 1 : 2
      }));
      
      const links = data.edges.map(edge => ({
        source: edge.source,
        target: edge.target,
        value: edge.value || 1
      }));
      
      return { nodes, links };
    },
    
    initChart() {
      console.log('初始化图表');
      const container = document.getElementById('function-graph');
      if (!container) {
        console.error('找不到图表容器元素 #function-graph');
        return;
      }
      
      try {
        // 初始化ECharts实例
        this.chart = echarts.init(container);
        
        // 注册窗口大小变化事件
        window.addEventListener('resize', this.resizeChart);
        
        console.log('图表初始化完成');
      } catch (error) {
        console.error('初始化图表时发生错误:', error);
      }
    },
    
    resizeChart() {
      if (this.chart) {
        this.chart.resize();
      }
    },
    
    renderGraph() {
      if (!this.chart) {
        console.error('图表未初始化，无法渲染');
        return;
      }
      
      console.log('开始渲染图表，节点数:', this.graphData.nodes.length, '边数:', this.graphData.links.length);
      
      try {
        // 准备有向图数据
        const graphNodes = this.graphData.nodes.map(node => ({
          id: node.id,
          name: node.name,
          package: node.package,
          callCount: node.callCount || 0,
          symbolSize: 10 + Math.sqrt(node.callCount || 1) * 3,
          itemStyle: {
            color: node.id === this.selectedFunction.key ? '#ff4757' : '#1e90ff'
          },
          label: {
            show: true,
            position: 'right',
            formatter: (params) => {
              const name = params.data.name || '';
              if (name.length > 20) {
                return name.substring(0, 18) + '...';
              }
              return name;
            }
          }
        }));
        
        const graphLinks = this.graphData.links.map(link => ({
          source: link.source,
          target: link.target,
          value: link.value || 1,
          lineStyle: {
            width: 1.5,
            curveness: 0.2,
            color: '#2980b9'
          },
          symbol: ['none', 'arrow'],
          symbolSize: [0, 15]
        }));
        
        // 设置图表选项
        const option = {
          tooltip: {
            trigger: 'item',
            formatter: (params) => {
              if (!params || !params.data) {
                return '无数据';
              }
              const data = params.data;
              return `
                <div style="font-weight:bold;margin-bottom:5px;">${data.name || '未命名'}</div>
                <div>包路径: ${data.package || '无'}</div>
                <div>调用次数: ${data.callCount || 0}</div>
              `;
            }
          },
          legend: {
            data: ['函数节点', '调用关系'],
            bottom: 10,
            left: 'center',
            selectedMode: false
          },
          toolbox: {
            show: true,
            feature: {
              restore: {
                show: true,
                title: '重置视图'
              },
              saveAsImage: {
                show: true,
                title: '保存为图片'
              }
            }
          },
          animationDuration: 1500,
          animationEasingUpdate: 'quinticInOut',
          series: [
            {
              type: 'graph',
              layout: 'force',
              data: graphNodes,
              links: graphLinks,
              roam: true,
              label: {
                show: true,
                position: 'right',
                backgroundColor: 'rgba(255,255,255,0.85)',
                padding: [3, 5],
                borderRadius: 3,
                color: '#333',
                fontSize: 12
              },
              lineStyle: {
                color: '#2980b9',
                width: 1.5,
                curveness: 0.2,
                opacity: 0.8
              },
              emphasis: {
                focus: 'adjacency',
                lineStyle: {
                  width: 3,
                  color: '#ff4757'
                },
                edgeSymbol: ['none', 'arrow'],
                edgeSymbolSize: [0, 20],
                itemStyle: {
                  color: '#fd7e14'
                },
                label: {
                  fontWeight: 'bold',
                  backgroundColor: 'rgba(255,255,255,0.9)'
                }
              },
              force: {
                repulsion: 150,
                gravity: 0.05,
                edgeLength: [80, 200],
                layoutAnimation: true,
                friction: 0.6
              },
              edgeSymbol: ['none', 'arrow'],
              edgeSymbolSize: [0, 15],
              draggable: true,
              // 高亮根节点
              itemStyle: {
                borderColor: '#fff',
                borderWidth: 2
              }
            }
          ]
        };
        
        // 设置图表选项并渲染
        this.chart.setOption(option);
        
        // 将根节点固定在中心位置
        const rootNodeIndex = graphNodes.findIndex(node => node.id === this.selectedFunction.key);
        if (rootNodeIndex !== -1) {
          this.chart.setOption({
            series: [{
              data: graphNodes.map((node, index) => {
                if (index === rootNodeIndex) {
                  return {
                    ...node,
                    fixed: true,
                    x: this.chart.getWidth() / 2,
                    y: this.chart.getHeight() / 2
                  };
                }
                return node;
              })
            }]
          });
        }
        
        // 添加说明文本
        this.$nextTick(() => {
          try {
            // 添加说明文本
            this.chart.setOption({
              graphic: {
                elements: [
                  {
                    type: 'text',
                    left: 'right',
                    top: 'top',
                    style: {
                      text: this.getGraphDescription(),
                      fontSize: 12,
                      fontWeight: 'normal',
                      fill: '#666',
                      backgroundColor: 'rgba(255,255,255,0.7)',
                      padding: [4, 8],
                      borderRadius: 3
                    }
                  }
                ]
              }
            });
            
            // 等待力导向图布局稳定后，自动缩放以显示所有节点
            setTimeout(() => {
              if (this.chart) {
                this.chart.dispatchAction({
                  type: 'graphRoam',
                  zoom: 0.8, // 稍微缩小一点，确保所有节点可见
                  originX: this.chart.getWidth() / 2,
                  originY: this.chart.getHeight() / 2
                });
              }
            }, 1500); // 等待1.5秒，让力导向图布局稳定
          } catch (error) {
            console.error('添加说明文本或自动缩放时发生错误:', error);
          }
        });
        
        console.log('图表渲染完成');
      } catch (error) {
        console.error('渲染图表时发生错误:', error);
      }
    },
    
    // 分析下游调用
    async analyzeDownstream() {
      if (!this.selectedFunction) return;
      
      this.analyzing = true;
      this.analysisType = 'downstream';
      
      try {
        const decodedPath = decodeURIComponent(this.dbFilePath);
        console.log('分析下游调用，请求参数:', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        const response = await axios.post('/api/static/function-downstream', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        console.log('分析下游调用，响应数据:', response.data);
        
        // 确保数据有效
        if (!response.data || !response.data.nodes || !response.data.edges) {
          console.error('服务器返回的数据格式不正确:', response.data);
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 处理数据并更新图形数据
        const newGraphData = this.processGraphData(response.data);
        console.log('处理后的图形数据:', newGraphData);
        
        // 确保有节点数据
        if (newGraphData.nodes.length === 0) {
          console.warn('没有节点数据，无法渲染图形');
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 成功获取数据后，更新图形数据
        this.graphData = newGraphData;
        this.analyzed = true;
        
        // 确保DOM已经更新后再初始化和渲染图形
        this.$nextTick(() => {
          // 如果图表已存在，先销毁
          if (this.chart) {
            this.chart.dispose();
            this.chart = null;
          }
          
          // 初始化并渲染图表
          this.initChart();
          this.renderGraph();
        });
      } catch (error) {
        console.error('分析下游调用失败:', error);
      } finally {
        this.analyzing = false;
      }
    },
    
    // 分析全链路调用
    async analyzeFullChain() {
      if (!this.selectedFunction) return;
      
      this.analyzing = true;
      this.analysisType = 'fullchain';
      
      try {
        const decodedPath = decodeURIComponent(this.dbFilePath);
        console.log('分析全链路调用，请求参数:', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        const response = await axios.post('/api/static/function-fullchain', {
          dbPath: decodedPath,
          functionKey: this.selectedFunction.key
        });
        
        console.log('分析全链路调用，响应数据:', response.data);
        
        // 确保数据有效
        if (!response.data || !response.data.nodes || !response.data.edges) {
          console.error('服务器返回的数据格式不正确:', response.data);
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 处理数据并更新图形数据
        const newGraphData = this.processGraphData(response.data);
        console.log('处理后的图形数据:', newGraphData);
        
        // 确保有节点数据
        if (newGraphData.nodes.length === 0) {
          console.warn('没有节点数据，无法渲染图形');
          this.analyzing = false;
          this.analyzed = true;
          return;
        }
        
        // 成功获取数据后，更新图形数据
        this.graphData = newGraphData;
        this.analyzed = true;
        
        // 确保DOM已经更新后再初始化和渲染图形
        this.$nextTick(() => {
          // 如果图表已存在，先销毁
          if (this.chart) {
            this.chart.dispose();
            this.chart = null;
          }
          
          // 初始化并渲染图表
          this.initChart();
          this.renderGraph();
        });
      } catch (error) {
        console.error('分析全链路调用失败:', error);
      } finally {
        this.analyzing = false;
      }
    },
    
    // 获取图表说明文本
    getGraphDescription() {
      switch (this.analysisType) {
        case 'upstream':
          return '箭头方向表示上游调用关系';
        case 'downstream':
          return '箭头方向表示下游调用关系';
        case 'fullchain':
          return '箭头方向表示函数调用关系';
        default:
          return '箭头方向表示调用关系';
      }
    },
    
    // 创建树形结构数据
    createTreeData() {
      // 找到目标函数节点（根节点）
      const rootNode = this.graphData.nodes.find(node => node.id === this.selectedFunction.key);
      if (!rootNode) {
        console.error('找不到根节点');
        return null;
      }
      
      // 创建节点ID到节点的映射
      const nodeMap = {};
      this.graphData.nodes.forEach(node => {
        nodeMap[node.id] = { ...node };
      });
      
      // 创建目标到源的映射（反向边，表示"被调用"关系）
      const targetToSources = {};
      this.graphData.links.forEach(link => {
        // 确保source和target是字符串
        const source = String(link.source);
        const target = String(link.target);
        
        if (!targetToSources[target]) {
          targetToSources[target] = [];
        }
        
        // 检查是否已经存在相同的源节点，避免重复
        const existingSource = targetToSources[target].find(s => s.id === source);
        if (!existingSource) {
          targetToSources[target].push({
            id: source,
            value: link.value || 1
          });
        }
      });
      
      // 记录已处理的节点，避免循环引用
      const processedNodes = new Set();
      
      // 递归构建树形结构
      const buildTree = (nodeId, depth = 0, maxDepth = 5, path = new Set()) => {
        // 防止无限递归
        if (depth > maxDepth) {
          return null;
        }
        
        // 检查是否形成循环路径
        if (path.has(nodeId)) {
          return null;
        }
        
        // 添加当前节点到路径
        const newPath = new Set(path);
        newPath.add(nodeId);
        
        const node = nodeMap[nodeId];
        if (!node) {
          return null;
        }
        
        // 确保所有必要的属性都存在
        const result = {
          id: node.id || '',
          name: node.name || '未命名',
          package: node.package || '',
          callCount: node.callCount || 0,
          value: node.callCount || 1,
          children: []
        };
        
        // 获取调用当前节点的所有函数（上游调用）
        const sources = targetToSources[nodeId] || [];
        
        // 对源节点进行排序，确保布局稳定
        const sortedSources = [...sources].sort((a, b) => {
          const nodeA = nodeMap[a.id];
          const nodeB = nodeMap[b.id];
          
          // 首先按照调用次数排序
          if (nodeA && nodeB && nodeA.callCount !== nodeB.callCount) {
            return nodeB.callCount - nodeA.callCount;
          }
          
          // 然后按照ID排序
          return (a.id || '').localeCompare(b.id || '');
        });
        
        // 限制每个节点的子节点数量，防止过于拥挤
        const maxChildrenPerNode = 8;
        const limitedSources = sortedSources.slice(0, maxChildrenPerNode);
        
        // 为每个源节点创建子节点
        for (const source of limitedSources) {
          // 避免自引用
          if (source.id === nodeId) {
            continue;
          }
          
          // 避免处理已经在当前路径中的节点，防止循环
          if (newPath.has(source.id)) {
            continue;
          }
          
          // 如果节点已经被处理过，并且深度大于1，则不再展开
          if (processedNodes.has(source.id) && depth > 1) {
            // 创建一个没有子节点的引用节点
            const sourceNode = nodeMap[source.id];
            if (sourceNode) {
              result.children.push({
                id: sourceNode.id || '',
                name: sourceNode.name || 'Unknown',
                package: sourceNode.package || '',
                callCount: sourceNode.callCount || 0,
                value: source.value || 1,
                isReference: true // 标记为引用节点
              });
            }
            continue;
          }
          
          // 递归构建子节点
          const childNode = buildTree(source.id, depth + 1, maxDepth, newPath);
          if (childNode) {
            result.children.push(childNode);
          }
        }
        
        // 如果超过了最大子节点数量，添加一个表示"更多"的节点
        if (sortedSources.length > maxChildrenPerNode) {
          result.children.push({
            id: `more_${nodeId}`,
            name: `...还有${sortedSources.length - maxChildrenPerNode}个调用者`,
            package: "",
            callCount: 0,
            value: 1
          });
        }
        
        // 标记节点已处理
        processedNodes.add(nodeId);
        
        return result;
      };
      
      // 从根节点开始构建树形结构
      return buildTree(rootNode.id);
    }
  }
};
</script>

<style scoped>
.function-graph {
  width: 100%;
  height: 600px;
  border: 1px solid #dee2e6;
  border-radius: 0 0 0.375rem 0.375rem;
  overflow: hidden;
  position: relative;
  background-color: #f8f9fa;
  box-shadow: inset 0 0 10px rgba(0,0,0,0.05);
}

.search-results {
  max-height: 300px;
  border: 1px solid #dee2e6;
  overflow-y: auto;
  border-radius: 0.375rem;
}

.selected-function .card-header {
  background-color: #0d6efd;
}

.card-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.card-header small {
  margin-top: 5px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .function-graph {
    height: 400px;
  }
}
</style> 