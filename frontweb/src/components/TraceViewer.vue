<template>
  <div class="mt-5">
    <h1 class="text-center">Trace Data Viewer</h1>

    <!-- 文件路径输入界面 -->
    <div v-if="!isPathVerified" class="container">
      <div class="row justify-content-center">
        <div class="col-md-8">
          <div class="card shadow">
            <div class="card-body">
              <h3 class="card-title text-center mb-4">请输入项目路径</h3>
              <div class="mb-3">
                <input
                  v-model="projectPath"
                  type="text"
                  class="form-control"
                  placeholder="例如: /path/to/your/project"
                  :class="{'is-invalid': pathError}"
                />
                <div class="invalid-feedback" v-if="pathError">
                  {{ pathError }}
                </div>
              </div>
              <div class="text-center">
                <button 
                  class="btn btn-primary"
                  @click="verifyPath"
                  :disabled="isVerifying"
                >
                  {{ isVerifying ? '验证中...' : '开始分析' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 分析界面 -->
    <div v-else>
      <div class="text-center mb-4">
        <div class="input-group w-50 mx-auto">
          <input
            v-model="functionName"
            type="text"
            placeholder="输入函数名称"
            class="form-control"
            @input="searchByFunctionName"
          />
          <button class="btn btn-primary" @click="searchByFunctionName">查询</button>
        </div>
        
        <!-- 显示当前项目路径 -->
        <div class="mt-2 text-muted">
          当前项目: {{ projectPath }}
          <button class="btn btn-sm btn-outline-secondary ms-2" @click="changePath">
            更换项目
          </button>
        </div>

        <!-- 函数名建议列表 -->
        <ul v-if="filteredFunctionNames.length" class="list-group mt-2 w-50 mx-auto">
          <li
            v-for="(name, index) in filteredFunctionNames"
            :key="index"
            class="list-group-item list-group-item-action"
            @click="selectFunction(name)"
          >
            {{ name }}
          </li>
        </ul>
      </div>

      <!-- 显示当前页和总页数 -->
      <div class="text-center mt-4">
        <p>当前页: {{ currentPage }} / {{ totalPages }}</p>
      </div>

      <!-- GID 表格展示 -->
      <h2 class="mt-5">相关 GIDs</h2>
      <table class="table table-bordered">
        <thead>
          <tr>
            <th>GID</th>
            <th>初始函数</th>
            <th>操作类型</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="result in filteredGIDs" :key="result.GID">
            <td>{{ result.GID }}</td>
            <td>{{ result.InitialFunc }}</td>
            <td>
              <router-link :to="{ name: 'TraceDetails', params: { gid: result.GID } }" class="btn btn-primary me-2">
                查看
              </router-link>
              <router-link :to="{ name: 'MermaidViewer', params: { gid: result.GID } }" class="btn btn-success">
                生成图片
              </router-link>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页按钮 -->
      <div class="text-center">
        <button @click="prevPage" :disabled="currentPage === 1">上一页</button>
        <button @click="nextPage" :disabled="currentPage >= totalPages">下一页</button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../axios';

export default {
  data() {
    return {
      projectPath: '',
      isPathVerified: false,
      isVerifying: false,
      pathError: '',
      gids: [],
      functionName: '',
      filteredGIDs: [],
      functionNames: [],
      filteredFunctionNames: [],
      currentPage: 1,
      itemsPerPage: 10, // 每页显示的 GID 数量
      totalPages: 0, // 总页数
    };
  },
  mounted() {
    // 检查本地存储中是否有已验证的路径
    const savedPath = localStorage.getItem('verifiedProjectPath');
    if (savedPath) {
      this.projectPath = savedPath;
      this.isPathVerified = true;
      this.initializeData();
    }
  },
  methods: {
    async verifyPath() {
      if (!this.projectPath.trim()) {
        this.pathError = '请输入项目路径';
        return;
      }

      this.isVerifying = true;
      this.pathError = '';

      try {
        // 发送验证请求
        const response = await axios.post('/api/verify/path', {
          path: this.projectPath
        });

        if (response.data.verified) {
          this.isPathVerified = true;
          // 保存验证通过的路径
          localStorage.setItem('verifiedProjectPath', this.projectPath);
          this.initializeData();
        } else {
          this.pathError = response.data.message || '项目路径验证失败';
        }
      } catch (error) {
        this.pathError = '验证过程出错: ' + (error.response?.data?.message || error.message);
      } finally {
        this.isVerifying = false;
      }
    },

    changePath() {
      this.isPathVerified = false;
      this.filteredGIDs = [];
      this.functionNames = [];
      this.filteredFunctionNames = [];
      localStorage.removeItem('verifiedProjectPath');
    },

    async initializeData() {
      await Promise.all([
        this.fetchGIDs(),
        this.fetchFunctionNames()
      ]);
    },

    async fetchGIDs() {
      try {
        const response = await axios.get('/api/gids', {
          params: {
            page: this.currentPage,
            limit: this.itemsPerPage,
          },
        });
        
        this.filteredGIDs = (response.data.body || []).map(item => ({
          GID: item.gid,
          InitialFunc: item.initialFunc
        }));
        
        this.totalPages = Math.ceil(response.data.total / this.itemsPerPage);
      } catch (error) {
        alert('获取GIDs失败: ' + error.message);
      }
    },

    async fetchFunctionNames() {
      try {
        const response = await axios.get('/api/functions');
        this.functionNames = response.data.functionNames || [];
      } catch (error) {
        alert('获取函数名列表失败: ' + error.message);
      }
    },

    searchByFunctionName() {
      if (this.functionName) {
        const searchTerm = this.functionName.toLowerCase();
        this.filteredFunctionNames = this.functionNames.filter(name =>
          name.toLowerCase().includes(searchTerm)
        );
        this.fetchGIDsByFunctionName();
      } else {
        this.filteredFunctionNames = [];
        this.filteredGIDs = this.gids;
      }
    },

    selectFunction(name) {
      this.functionName = name;
      this.filteredFunctionNames = [];
      this.searchByFunctionName();
    },

    async fetchGIDsByFunctionName() {
      try {
        const response = await axios.post('/api/gids/function', {
          functionName: this.functionName,
          path: this.projectPath
        });
        this.filteredGIDs = response.data.gids || [];
      } catch (error) {
        alert('搜索函数相关GIDs失败: ' + error.message);
      }
    },

    nextPage() {
      this.currentPage++;
      this.fetchGIDs(); // 重新获取 GIDs
    },

    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--;
        this.fetchGIDs(); // 重新获取 GIDs
      }
    },

    async getFirstFunctionName(gid) {
      try {
        const response = await axios.get(`/api/traces/${gid}`); // 假设这个 API 返回该 GID 的 trace 数据

        // 检查 trace_data 是否存在并且是数组
        if (response.data.traceData && Array.isArray(response.data.traceData) && response.data.traceData.length > 0) {
          return response.data.traceData[0].name; // 返回第一个函数名
        } else {
          console.warn('trace_data is empty or not an array:', response.data.traceData);
        }
      } catch (error) {
        console.error('获取函数名失败:', error);
      }
      return '无函数名'; // 返回默认值
    },
  },
};
</script>

<style scoped>
.card {
  transition: transform 0.2s;
}
.card:hover {
  transform: scale(1.05);
}
.list-group-item {
  cursor: pointer;
}
.input-group {
  transition: all 0.3s ease;
}
.input-group:focus-within {
  transform: scale(1.02);
}
</style>