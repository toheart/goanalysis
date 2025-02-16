<!-- Start of Selection -->
<template>
  <div class="container mt-5">
      <h1 class="text-center">Trace Details for GID: {{ gid }}</h1>
      <button @click="$router.go(-1)" class="btn btn-secondary mb-3" style="float: right;">返回</button><br>
    <div v-if="traceData">
      <div class="mt-4">
        <template v-for="(value, key) in traceData" :key="key">
          <div class="stack-item" :style="{ marginLeft: value.indent*5 + 'em', backgroundColor: '#f8f9fa' }">
            <div class="row" style="background-color: #f1f1f1;">
                <div class="col-md-8" style="background-color: #e9ecef;">
                    <p class="mb-1 text-left" style="text-align: left;">Name: {{ value.name }} ({{ value.timeCost }})</p>
                </div>
                <div class="col-md-4" style="background-color: #e9ecef;">
                <button v-if="value.paramCount > 0" class="btn btn-sm btn-outline-primary" @click="() => viewParameters(value.id)">查看参数</button>
                </div>
            </div>
          </div>
        </template>
      </div>
    </div>
    <div v-else>
      <p>Loading...</p>
    </div>

    <!-- 模态框 -->
    <div class="modal fade" id="paramsModal" tabindex="-1" role="dialog" aria-labelledby="paramsModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title text-center" id="paramsModalLabel">参数详情</h5>
          </div>
          <div class="modal-body">
            <table class="table text-center">
              <thead>
                <tr>
                  <th>位置</th>
                  <th>参数</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(param, index) in parameters" :key="index">
                  <td>{{ param.pos }}</td>
                  <td>{{ param.param }}</td>
                </tr>
              </tbody>
            </table>
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
    };
  },
  mounted() {
    this.fetchTraceDetails();
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
    }
  },
};
</script>

<style>
/* 添加样式 */
.list-group-item {
  border: 1px solid #ddd;
  border-radius: 5px;
  margin-bottom: 10px;
}

#paramsModal {
  max-width: 600px; /* 设置模态框的最大宽度 */
}
.table {
  margin: auto; /* 使表格居中 */
}
</style> 