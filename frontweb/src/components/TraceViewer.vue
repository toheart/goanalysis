<template>
  <div class="mt-5">
    <h1 class="text-center">Trace Data Viewer</h1>
    <div class="text-center mb-4">
      <input
        v-model="functionName"
        type="text"
        placeholder="输入函数名称"
        class="form-control w-50 d-inline"
        @input="searchByFunctionName"
      />
      <button class="btn btn-primary" @click="searchByFunctionName">查询</button>
      <ul v-if="filteredFunctionNames.length" class="list-group mt-2">
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
    <h2 class="mt-5">相关 GIDs</h2>
    <div class="row">
      <div class="col-md-4 mb-4" v-for="gid in filteredGIDs" :key="gid">
        <div class="card shadow-sm">
          <div class="card-body text-center">
            <h5 class="card-title">GID: {{ gid }}</h5>
            <router-link :to="{ name: 'TraceDetails', params: { gid } }" class="btn btn-primary">查看</router-link>
            <router-link :to="{ name: 'MermaidViewer', params: { gid } }" class="btn btn-success">生成图片</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../axios';

export default {
  data() {
    return {
      gids: [],
      functionName: '',
      filteredGIDs: [],
      functionNames: [], // 用于存储所有函数名
      filteredFunctionNames: [], // 用于存储过滤后的函数名
    };
  },
  mounted() {
    this.fetchGIDs();
    this.fetchFunctionNames(); // 获取所有函数名
  },
  methods: {
    async fetchGIDs() {
      const response = await axios.get('/api/gids');
      this.gids = response.data.gids || [];
      this.filteredGIDs = this.gids; // 初始化为所有 GIDs
    },
    async fetchFunctionNames() {
      const response = await axios.get('/api/functions'); // 假设有一个 API 获取所有函数名
      this.functionNames = response.data.functionNames || [];
    },
    searchByFunctionName() {
      // 根据输入的函数名称过滤函数名
      if (this.functionName) {
        const searchTerm = this.functionName.toLowerCase();
        this.filteredFunctionNames = this.functionNames.filter(name =>
          name.toLowerCase().includes(searchTerm)
        );
        this.fetchGIDsByFunctionName();
      } else {
        this.filteredFunctionNames = []; // 如果没有输入，清空下拉框
      }
    },
    selectFunction(name) {
      this.functionName = name; // 设置输入框的值为选中的函数名
      this.filteredFunctionNames = []; // 清空下拉框
      this.searchByFunctionName(); // 触发搜索 GIDs
    },
    async fetchGIDsByFunctionName() {
      const response = await axios.post(`/api/gids/function`, {
        functionName: this.functionName
      });
      this.filteredGIDs = response.data.gids || [];
      // 使用gids重新渲染
      this.$forceUpdate();
    },
  },
};
</script>

<style>
/* 添加样式 */
.card {
  transition: transform 0.2s;
}
.card:hover {
  transform: scale(1.05);
}
.list-group-item {
  cursor: pointer;
}
</style>