<template>
  <div class="container mt-5">
    <h1 class="text-center">groutine({{gid}})调用图展示:</h1>
    <div id="mermaidContainer" class="mermaid mt-5" style="border: 1px solid #ddd; padding: 20px; border-radius: 5px; background-color: #f9f9f9;">
      <!-- Mermaid 图将动态插入到这里 -->
    </div>
    <router-link :to="{ name: 'TraceViewer' }" class="btn btn-primary">返回</router-link>
  </div>
</template>

<script>
import axios from '../axios';
import mermaid from 'mermaid';
import svgPanZoom from 'svg-pan-zoom';

export default {
  data() {
    return {
      gid: this.$route.params.gid,
      mermaidString: '', // 用于存储 Mermaid 字符串
    };
  },
  mounted() {
    this.fetchMermaidData();
    mermaid.initialize({ 
        startOnLoad: false ,
        maxTextSize: 90000,
        height: 1000,
    });
  },
  methods: {
    async fetchMermaidData() {
      try {
        const response = await axios.get(`/api/traces/${this.gid}/mermaid`);
        this.mermaidString = response.data.image; // 获取 Mermaid 字符串
        this.renderMermaid(); // 调用渲染函数显示图片
      } catch (error) {
        console.error("获取 mermaid 语法字符串失败:", error);
      }
    },
    renderMermaid() {
      const element = document.getElementById('mermaidContainer');
      if (element) {
        element.innerHTML = this.mermaidString; // 设置容器的内容为 Mermaid 字符串
        mermaid.run({
            querySelector: '#mermaidContainer',
            postRenderCallback: (id) => {
                let ele = document.getElementById(id);
                    let svg = ele.getBBox();
                    let height = svg.height;
                    let aHeight = height > 800 ? 800 : height;
                    ele.setAttribute('style','height: '+aHeight+'px;overflow:scroll;');
                    let panZoomTiger = svgPanZoom('#'+id,{
                        zoomEnabled: true,
                        controlIconsEnabled: true
                    });
                    panZoomTiger.resize();
                    panZoomTiger.updateBBox();
            }
        }); // 初始化 Mermaid 渲染
      }
    },
    refreshMermaid() {
      this.renderMermaid(); // 刷新图形
    }
  },
};
</script>

<style>
#mermaidContainer {
  min-height: 500px; /* 设置最小高度以确保有足够的空间显示图 */
  border-radius: 5px; /* 圆角 */
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1); /* 添加阴影 */
  background-color: #f9f9f9; /* 背景颜色 */
}
</style> 