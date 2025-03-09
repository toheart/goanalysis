<template>
  <div class="static-analysis-with-monitor">
    <!-- 路径输入和验证部分 -->
    <div class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="bi bi-folder2-open me-2"></i>{{ $t('staticAnalysis.form.projectPath') }}</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-8">
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-folder"></i></span>
              <input 
                type="text" 
                class="form-control" 
                :placeholder="$t('staticAnalysis.form.projectPathPlaceholder')" 
                v-model="projectPath"
                :disabled="isAnalyzing"
              >
              <button 
                class="btn btn-primary" 
                @click="startAnalysis" 
                :disabled="!projectPath || isAnalyzing"
              >
                <span v-if="isAnalyzing" class="spinner-border spinner-border-sm me-2" role="status"></span>
                {{ isAnalyzing ? $t('staticAnalysis.form.analyzing') : $t('staticAnalysis.form.startAnalysis') }}
              </button>
            </div>
            
            <div v-if="pathError" class="alert alert-danger mt-2">
              <i class="bi bi-exclamation-triangle-fill me-2"></i>{{ pathError }}
            </div>
          </div>
          <div class="col-md-4">
            <div class="alert alert-info mb-0">
              <h6><i class="bi bi-info-circle me-2"></i>{{ $t('staticAnalysis.form.tip') }}</h6>
              <p class="mb-0 small">{{ $t('staticAnalysis.form.pathTip') }}</p>
            </div>
          </div>
        </div>
        
        <!-- 分析选项 -->
        <div class="row mt-3">
          <div class="col-12">
            <div class="card analysis-options-card" id="analysis-options">
              <div class="card-header d-flex justify-content-between align-items-center bg-primary text-white">
                <h6 class="mb-0"><i class="bi bi-gear me-2"></i>{{ $t('staticAnalysis.options.title') }}</h6>
                <button class="btn btn-sm btn-light" @click="showOptions = !showOptions">
                  {{ showOptions ? $t('staticAnalysis.options.hideOptions') : $t('staticAnalysis.options.showOptions') }}
                </button>
              </div>
              <div class="card-body" v-if="showOptions">
                <div class="alert alert-primary mb-3">
                  <i class="bi bi-info-circle-fill me-2"></i>
                  <strong>{{ $t('staticAnalysis.form.tip') }}：</strong> {{ $t('staticAnalysis.options.optionsTip') }}
                </div>
                <div class="row">
                  <!-- 算法选择 -->
                  <div class="col-md-6 mb-3">
                    <label class="form-label fw-bold">{{ $t('staticAnalysis.options.algorithm') }}</label>
                    <select 
                      class="form-select" 
                      v-model="analysisOptions.algo"
                    >
                      <option value="vta">{{ $t('staticAnalysis.options.algorithms.vta') }}</option>
                      <option value="rta">{{ $t('staticAnalysis.options.algorithms.rta') }}</option>
                      <option value="cha">{{ $t('staticAnalysis.options.algorithms.cha') }}</option>
                      <option value="static">{{ $t('staticAnalysis.options.algorithms.static') }}</option>
                    </select>
                    <div class="form-text">{{ $t('staticAnalysis.options.algorithmTip') }}</div>
                  </div>
                                  
                  <!-- 仅分析特定方法 -->
                  <div class="col-md-12 mb-3">
                    <label class="form-label fw-bold">{{ $t('staticAnalysis.options.ignoreMethod') }}</label>
                    <input 
                      type="text" 
                      class="form-control" 
                      v-model="analysisOptions.ignoreMethod"
                      :placeholder="$t('staticAnalysis.options.ignoreMethodPlaceholder')"
                    >
                    <div class="form-text">{{ $t('staticAnalysis.options.ignoreMethodTip') }}</div>
                  </div>
                </div>
              </div>
              <div class="card-body text-center" v-else>
                <button class="btn btn-outline-primary" @click="showOptions = true">
                  <i class="bi bi-sliders me-2"></i>{{ $t('staticAnalysis.options.showOptions') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 任务监控部分 -->
    <div class="card" v-if="taskId">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="bi bi-activity me-2"></i>{{ $t('staticAnalysis.monitor.title') }}
        </h5>
        <div>
          <span v-if="isConnected" class="badge bg-success me-2">
            <i class="bi bi-wifi me-1"></i>{{ $t('staticAnalysis.monitor.connected') }}
          </span>
          <span v-else class="badge bg-danger me-2">
            <i class="bi bi-wifi-off me-1"></i>{{ $t('staticAnalysis.monitor.disconnected') }}
          </span>
          <button 
            v-if="!isConnected" 
            class="btn btn-sm btn-outline-primary" 
            @click="connect"
            :disabled="connecting"
          >
            <span v-if="connecting" class="spinner-border spinner-border-sm me-1" role="status"></span>
            {{ connecting ? $t('staticAnalysis.monitor.connecting') : $t('staticAnalysis.monitor.connectServer') }}
          </button>
          <button 
            v-else 
            class="btn btn-sm btn-outline-danger" 
            @click="disconnect"
          >
            {{ $t('staticAnalysis.monitor.disconnect') }}
          </button>
        </div>
      </div>
      <div class="card-body">
        <div v-if="taskId && taskStatus">
          <div class="d-flex justify-content-between align-items-center mb-3">
            <h6 class="mb-0">{{ $t('staticAnalysis.monitor.taskId') }}: {{ taskId }}</h6>
            <span 
              class="badge" 
              :class="{
                'bg-primary': taskStatus.status === 'processing',
                'bg-success': taskStatus.status === 'completed',
                'bg-danger': taskStatus.status === 'failed',
                'bg-secondary': taskStatus.status === 'not_found'
              }"
            >
              {{ getStatusText(taskStatus.status) }}
            </span>
          </div>
          
          <div class="progress mb-3">
            <div 
              class="progress-bar progress-bar-striped" 
              :class="{'progress-bar-animated': taskStatus.status === 'processing'}"
              role="progressbar" 
              :style="{width: taskStatus.progress + '%'}" 
              :aria-valuenow="taskStatus.progress" 
              aria-valuemin="0" 
              aria-valuemax="100"
            >
              {{ Math.round(taskStatus.progress) }}%
            </div>
          </div>
          
          <div class="card mb-3">
            <div class="card-header d-flex justify-content-between align-items-center">
              <h6 class="mb-0"><i class="bi bi-terminal me-2"></i>{{ $t('staticAnalysis.monitor.logs') }}</h6>
              <span class="badge bg-info">{{ messages.length }} {{ $t('staticAnalysis.monitor.messages') }}</span>
            </div>
            <div class="message-container p-3">
              <div v-for="(message, index) in messages" :key="index" class="message">
                <div class="message-content p-2 rounded" :class="getMessageClass(message)">
                  {{ formatMessage(message) }}
                </div>
              </div>
              <div v-if="messages.length === 0" class="text-center text-muted py-4">
                <i class="bi bi-chat-square-text me-2"></i>{{ $t('staticAnalysis.monitor.noMessages') }}
              </div>
            </div>
          </div>
          
          <div v-if="taskStatus.status === 'completed'" class="alert alert-success">
            <i class="bi bi-check-circle-fill me-2"></i>{{ $t('staticAnalysis.monitor.completed') }}
            <button class="btn btn-sm btn-success ms-2" @click="$emit('refresh-db-files')">
              {{ $t('staticAnalysis.monitor.refreshDb') }}
            </button>
          </div>
          
          <div v-if="taskStatus.status === 'failed'" class="alert alert-danger">
            <i class="bi bi-exclamation-triangle-fill me-2"></i>{{ $t('staticAnalysis.monitor.failed') }}: {{ taskStatus.message }}
          </div>
        </div>
        
        <div v-else-if="isConnected && !taskStatus" class="text-center py-4">
          <i class="bi bi-hourglass-split text-primary display-4"></i>
          <h5 class="mt-3">{{ $t('staticAnalysis.monitor.waitingTask') }}</h5>
          <p class="text-muted">{{ $t('staticAnalysis.monitor.waitingTaskTip') }}</p>
        </div>
        
        <div v-else-if="!isConnected" class="text-center py-4">
          <i class="bi bi-wifi-off text-secondary display-4"></i>
          <h5 class="mt-3">{{ $t('staticAnalysis.monitor.notConnected') }}</h5>
          <p class="text-muted">{{ $t('staticAnalysis.monitor.notConnectedTip') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useI18n } from 'vue-i18n';
import axios from '../../axios';

export default {
  name: 'StaticAnalysisWithMonitor',
  props: {
    initialProjectPath: {
      type: String,
      default: ''
    },
    initialAnalysisOptions: {
      type: Object,
      default: () => ({
        algo: 'vta',
        ignoreMethod: ''
      })
    }
  },
  setup() {
    const { t, locale } = useI18n({ useScope: 'global' });
    return { t, locale };
  },
  data() {
    return {
      projectPath: this.initialProjectPath,
      analysisOptions: { ...this.initialAnalysisOptions },
      showOptions: false,
      isAnalyzing: false,
      pathError: '',
      
      // 任务监控相关数据
      taskId: '',
      eventSource: null,
      isConnected: false,
      connecting: false,
      taskStatus: null,
      messages: [],
      statusCheckInterval: null,
      parseErrorCount: 0
    }
  },
  mounted() {
    // 添加语言变化监听
    window.addEventListener('languageChanged', this.handleLanguageChange);
  },
  beforeUnmount() {
    // 移除语言变化监听
    window.removeEventListener('languageChanged', this.handleLanguageChange);
    this.disconnect();
  },
  methods: {
    // 处理语言变化
    handleLanguageChange(event) {
      console.log('StaticAnalysisWithMonitor - Language changed:', event.detail.locale);
      // 强制刷新组件中的国际化文本
      this.$forceUpdate();
    },
    
    async startAnalysis() {
      if (!this.projectPath) {
        this.pathError = this.$t('staticAnalysis.form.pathError');
        return;
      }
      
      this.isAnalyzing = true;
      this.pathError = '';
      
      try {
        const apiBaseUrl = process.env.VUE_APP_API_URL || '';
        const response = await axios.post(`${apiBaseUrl}/api/static/analyze/path`, {
          path: this.projectPath,
          algo: this.analysisOptions.algo,
          ignoreMethod: this.analysisOptions.ignoreMethod
        });

        if (response.status !== 200) {
          throw new Error(`HTTP ${response.status}: ${response.data || this.$t('common.error')}`);
        }

        const data = response.data;

        if (data.taskId) {
          this.taskId = data.taskId;
          this.$emit('analysis-started', this.taskId);
          
          // 自动连接到事件流
          this.connect();
          this.startStatusCheck();
        } else {
          throw new Error(this.$t('staticAnalysis.form.noTaskId'));
        }
      } catch (error) {
        console.error('启动分析任务失败:', error);
        this.pathError = `${this.$t('staticAnalysis.form.analysisError')}: ${error.message}`;
        this.isAnalyzing = false;
      }
    },
    
    // 任务监控相关方法
    connect() {
      if (!this.taskId) return;
      
      if (this.eventSource) {
        this.disconnect();
      }
      
      this.connecting = true;
      
      try {
        // 构建 SSE URL
        const apiBaseUrl = process.env.VUE_APP_API_URL || '';
        const sseUrl = `${apiBaseUrl}/api/static/analysis/${this.taskId}`;
        console.log('Connecting to SSE:', sseUrl);
        
        this.eventSource = new EventSource(sseUrl);
        
        this.eventSource.onopen = () => {
          console.log('SSE connection established');
          this.isConnected = true;
          this.connecting = false;
          this.$emit('connected');
        };
        
        this.eventSource.onmessage = (event) => {
          try {
            // 将消息添加到消息列表中
            this.messages.push(event.data);
            this.scrollToBottom();
            
            // 尝试解析消息内容
            let messageType = 0;
            let isCompletionMessage = false;
            
            try {
              // 尝试从消息中提取type和message
              if (event.data.includes('"type"') && event.data.includes('"message"')) {
                const typeMatch = event.data.match(/"type":(\d+)/);
                if (typeMatch && typeMatch[1]) {
                  messageType = parseInt(typeMatch[1]);
                  // 检查是否是完成消息
                  isCompletionMessage = (messageType === 2);
                }
              }
              
              // 检查消息内容是否表示完成
              isCompletionMessage = isCompletionMessage || 
                                   event.data.includes('"type":2') || 
                                   event.data.includes('Analysis task completed') ||
                                   event.data.includes('分析完成');
            } catch (parseError) {
              console.warn('解析消息类型失败:', parseError);
            }
            
            // 根据消息类型更新任务状态
            if (isCompletionMessage) {
              console.log('收到分析完成消息，更新状态并断开连接');
              // 更新任务状态为完成
              this.taskStatus = {
                status: 'completed',
                progress: 100,
                message: this.formatMessage(event.data)
              };
              this.$emit('task-completed', this.taskStatus);
              this.$emit('status-updated', this.taskStatus);
              this.isAnalyzing = false;
              
              // 延迟断开连接，确保消息已经处理完毕
              setTimeout(() => {
                if (this.eventSource) {
                  console.log('分析完成，主动断开SSE连接');
                  this.disconnect();
                }
              }, 500);
            }
          } catch (e) {
            console.error('处理SSE消息失败:', e);
            // 如果处理失败，将原始消息添加到列表
            this.messages.push(`${this.$t('staticAnalysis.monitor.parseError')}: ${e.message}`);
            this.scrollToBottom();
            
            // 如果连续解析失败次数过多，则断开连接
            this.parseErrorCount = (this.parseErrorCount || 0) + 1;
            if (this.parseErrorCount > 3) {
              console.error('连续解析失败次数过多，断开连接');
              this.taskStatus = {
                status: 'failed',
                progress: 0,
                message: this.$t('staticAnalysis.monitor.tooManyParseErrors')
              };
              this.$emit('task-completed', this.taskStatus);
              this.$emit('status-updated', this.taskStatus);
              this.isAnalyzing = false;
              this.disconnect();
              this.stopStatusCheck();
            }
          }
        };
        
        this.eventSource.onerror = (error) => {
          console.error('SSE error:', error);
          
          // 检查是否已经完成分析，如果已完成则不显示错误
          if (this.taskStatus && this.taskStatus.status === 'completed') {
            console.log('SSE connection closed after task completion, ignoring error');
            this.isConnected = false;
            this.connecting = false;
            return;
          }
          
          // 检查消息中是否包含分析完成的标志
          const hasCompletionMessage = this.messages.some(msg => 
            msg.includes('"type":2') || 
            msg.includes('Analysis task completed') || 
            msg.includes('分析完成')
          );
          
          if (hasCompletionMessage) {
            console.log('Task already completed, ignoring SSE error');
            // 更新连接状态但不显示错误
            this.isConnected = false;
            this.connecting = false;
            return;
          }
          
          // 如果不是因为任务完成导致的连接关闭，则显示错误
          this.isConnected = false;
          this.connecting = false;
          this.$emit('error', error);
          
          // 更新任务状态为失败
          this.taskStatus = {
            status: 'failed',
            progress: 0,
            message: this.$t('staticAnalysis.monitor.connectionError')
          };
          this.$emit('task-completed', this.taskStatus);
          this.$emit('status-updated', this.taskStatus);
          this.isAnalyzing = false;
          
          // 断开连接并停止状态检查
          this.disconnect();
          this.stopStatusCheck();
        };
      } catch (error) {
        console.error('Failed to create SSE connection:', error);
        this.isConnected = false;
        this.connecting = false;
        this.$emit('error', error);
        this.isAnalyzing = false;
      }
    },
    
    disconnect() {
      if (this.eventSource) {
        this.eventSource.close();
        this.eventSource = null;
        this.isConnected = false;
        this.stopStatusCheck();
      }
    },
    
    startStatusCheck() {
      this.fetchTaskStatus();
      this.stopStatusCheck();
      
      this.statusCheckInterval = setInterval(() => {
        this.fetchTaskStatus();
      }, 5000);
    },
    
    stopStatusCheck() {
      if (this.statusCheckInterval) {
        clearInterval(this.statusCheckInterval);
        this.statusCheckInterval = null;
      }
    },
    
    async fetchTaskStatus() {
      if (!this.taskId) return;
      
      try {
        const apiBaseUrl = process.env.VUE_APP_API_URL || '';
        const response = await fetch(`${apiBaseUrl}/api/static/task/${this.taskId}/status`);
        
        if (!response.ok) {
          const errorText = await response.text();
          console.error(`获取任务状态失败: HTTP ${response.status} - ${errorText}`);
          this.taskStatus = {
            status: 'failed',
            progress: 0,
            message: `${this.$t('staticAnalysis.monitor.requestFailed')}: HTTP ${response.status} - ${errorText || this.$t('common.error')}`
          };
          this.stopStatusCheck();
          this.$emit('task-completed', this.taskStatus);
          this.$emit('status-updated', this.taskStatus);
          this.isAnalyzing = false;
          this.disconnect();
          return;
        }
        
        const data = await response.json();
        
        // 更新任务状态，包括进度信息
        this.taskStatus = {
          status: this.getStatusFromCode(data.status),
          message: data.message,
          progress: data.progress || 0 // 使用服务器返回的进度，如果没有则默认为0
        };
        
        // 触发状态更新事件
        this.$emit('status-updated', this.taskStatus);
        
        if (data.status === 2 || data.status === -1) { // 完成或失败
          this.stopStatusCheck();
          this.$emit('task-completed', this.taskStatus);
          this.isAnalyzing = false;
          this.disconnect();
        }
      } catch (error) {
        console.error('获取任务状态失败:', error);
        this.taskStatus = {
          status: 'failed',
          progress: 0,
          message: `${this.$t('staticAnalysis.monitor.requestException')}: ${error.message || this.$t('common.error')}`
        };
        this.stopStatusCheck();
        this.$emit('task-completed', this.taskStatus);
        this.$emit('status-updated', this.taskStatus);
        this.isAnalyzing = false;
        this.disconnect();
      }
    },
    
    // 将状态码转换为状态字符串
    getStatusFromCode(code) {
      switch (code) {
        case 0: return 'processing'; // 启动中
        case 1: return 'processing'; // 处理中
        case 2: return 'completed';  // 已完成
        case -1: return 'failed';    // 失败
        default: return 'not_found'; // 未找到
      }
    },
    
    getStatusText(status) {
      return this.$t(`staticAnalysis.monitor.status.${status}`) || status;
    },
    
    getMessageClass(message) {
      // 尝试从JSON中提取消息内容
      let content = message;
      try {
        if (message.startsWith('{') && message.endsWith('}')) {
          const data = JSON.parse(message);
          if (data.message) {
            content = data.message;
          }
        } else if (message.includes('"message"')) {
          const match = message.match(/"message":"([^"]+)"/);
          if (match && match[1]) {
            content = match[1];
          }
        }
      } catch (e) {
        // 解析失败，使用原始消息
        content = message;
      }
      
      // 根据消息内容设置样式
      if (content.includes('失败') || content.includes('错误') || 
          content.includes('failed') || content.includes('Failed') || 
          content.includes('error') || content.includes('Error')) {
        return 'bg-danger text-white';
      } else if (content.includes('完成') || content.includes('completed') || 
                content.includes('Completed') || content.includes('success') || 
                content.includes('Success') || content.includes('Saved')) {
        return 'bg-success text-white';
      } else if (content.includes('开始') || content.includes('Starting') || 
                content.includes('start') || content.includes('Start')) {
        return 'bg-primary text-white';
      } else {
        return 'bg-white border';
      }
    },
    
    scrollToBottom() {
      this.$nextTick(() => {
        const container = document.querySelector('.message-container');
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
    },
    
    formatMessage(message) {
      try {
        // 尝试解析JSON格式的消息
        if (message.startsWith('{') && message.endsWith('}')) {
          const data = JSON.parse(message);
          if (data.message) {
            return data.message;
          }
        }
        
        // 处理特定格式的消息
        if (message.includes('"type"') && message.includes('"message"')) {
          // 尝试提取message字段
          const match = message.match(/"message":"([^"]+)"/);
          if (match && match[1]) {
            return match[1];
          }
        }
        
        // 如果无法解析，则返回原始消息
        return message;
      } catch (e) {
        console.error('格式化消息失败:', e);
        return message;
      }
    }
  }
}
</script>

<style scoped>
.analysis-options-card {
  border: 2px solid #007bff;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 1.5rem;
}

.analysis-options-card .card-header {
  background-color: #007bff;
  color: white;
  font-weight: bold;
}

.analysis-options-card .form-label {
  color: #333;
}

.analysis-options-card .form-text {
  color: #6c757d;
  font-size: 0.85rem;
}

.message-container {
  border: 1px solid #dee2e6;
  max-height: 400px;
  overflow-y: auto;
  background-color: #f8f9fa;
}

.message-content {
  border-left: 3px solid #6c757d;
  margin-bottom: 8px;
  padding: 8px 12px;
}

.message-content.bg-danger {
  border-left-color: #dc3545;
}

.message-content.bg-success {
  border-left-color: #28a745;
}

.message-content.bg-primary {
  border-left-color: #007bff;
}

.message-content.bg-white {
  border-left-color: #6c757d;
  background-color: #ffffff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
</style> 