<template>
  <div id="app">
    <div>
      <nav class="navbar navbar-expand-lg navbar-light bg-light">
        <a class="navbar-brand" href="#">go代码分析</a>
        <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav"
          aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav">
            <li class="nav-item">
              <router-link to="/allgids" class="nav-link">程序运行分析</router-link>
            </li>
            <li class="nav-item">
              <router-link to="/static-analysis" class="nav-link">程序调用静态分析</router-link>
            </li>

            <!-- 可以添加更多导航项 -->
          </ul>
        </div>
      </nav>
      <main class="container mt-4">
        <router-view></router-view> <!-- 动态内容区域 -->
      </main>
    </div>
  </div>
</template>

<script>

export default {
  components: {
  },
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
