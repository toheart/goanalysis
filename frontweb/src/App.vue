<template>
  <div id="app">
    <div>
      <nav class="navbar navbar-expand-lg navbar-dark">
        <div class="container-fluid">
          <router-link to="/" class="navbar-brand">
            <i class="bi bi-code-square me-2"></i>Go代码分析平台
          </router-link>
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav"
            aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
          </button>
          <div class="collapse navbar-collapse" id="navbarNav">
            <ul class="navbar-nav">
              <li class="nav-item">
                <router-link to="/" class="nav-link" exact-active-class="active">
                  <i class="bi bi-house-door me-1"></i>首页
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/allgids" class="nav-link" active-class="active">
                  <i class="bi bi-activity me-1"></i>程序运行分析
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/static-analysis" class="nav-link" active-class="active">
                  <i class="bi bi-diagram-3 me-1"></i>程序调用静态分析
                </router-link>
              </li>
            </ul>
          </div>
        </div>
      </nav>
      <main class="container mt-4 fade-in">
        <router-view></router-view> <!-- 动态内容区域 -->
      </main>
      <footer class="footer mt-5 py-3 bg-light">
        <div class="container text-center">
          <span class="text-muted">Go代码分析平台 &copy; 2024</span>
        </div>
      </footer>
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
  font-family: 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.footer {
  margin-top: 60px;
  border-top: 1px solid #eaeaea;
}
</style>
