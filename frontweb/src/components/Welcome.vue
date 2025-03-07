<template>
  <div class="welcome-page">
    <div class="container py-5">
      <div class="text-center mb-5">
        <h1 class="display-4 fw-bold text-primary">Go代码分析平台</h1>
        <p class="lead">欢迎使用Go代码分析平台，一个强大的Go程序分析工具</p>
      </div>

      <div class="row justify-content-center mb-5">
        <div class="col-md-8">
          <div class="card shadow-lg border-0">
            <div class="card-body p-5">
              <h2 class="card-title text-center mb-4"><i class="bi bi-info-circle me-2"></i>关于本平台</h2>
              <p class="card-text">
                Go代码分析平台是一个专为Go语言开发者设计的工具，旨在帮助您更好地理解和分析Go程序的运行机制和调用关系。
                通过可视化的方式展示程序的执行流程和函数调用关系，帮助您更高效地进行代码审查、性能优化和问题排查。
              </p>
            </div>
          </div>
        </div>
      </div>

      <h2 class="text-center mb-4"><i class="bi bi-tools me-2"></i>主要功能</h2>
      <div class="row g-4 mb-5">
        <div class="col-md-6">
          <div class="card h-100 highlighted-container">
            <div class="card-body p-4">
              <div class="d-flex align-items-center mb-3">
                <div class="feature-icon bg-primary text-white me-3">
                  <i class="bi bi-activity"></i>
                </div>
                <h3 class="card-title h5 mb-0">程序运行分析</h3>
              </div>
              <p class="card-text">
                通过收集程序运行时的调用信息，生成详细的函数调用链和执行路径，帮助您了解程序的实际执行流程。
                支持查看函数参数、返回值和执行耗时等详细信息。
              </p>
              <router-link to="/allgids" class="btn btn-outline-primary mt-3">
                <i class="bi bi-arrow-right me-1"></i>开始分析
              </router-link>
            </div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card h-100 highlighted-container">
            <div class="card-body p-4">
              <div class="d-flex align-items-center mb-3">
                <div class="feature-icon bg-success text-white me-3">
                  <i class="bi bi-diagram-3"></i>
                </div>
                <h3 class="card-title h5 mb-0">程序调用静态分析</h3>
              </div>
              <p class="card-text">
                通过静态分析Go程序的源代码，生成函数调用关系图，帮助您理解程序的结构和组件间的依赖关系。
                支持包级别和函数级别的调用关系分析，方便您进行代码重构和架构优化。
              </p>
              <router-link to="/static-analysis" class="btn btn-outline-success mt-3">
                <i class="bi bi-arrow-right me-1"></i>查看分析
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <div class="text-center">
        <h2 class="mb-4"><i class="bi bi-rocket me-2"></i>开始使用</h2>
        <div class="d-grid gap-3 d-sm-flex justify-content-sm-center">
          <router-link to="/allgids" class="btn btn-primary btn-lg px-4 gap-3">
            <i class="bi bi-activity me-2"></i>程序运行分析
          </router-link>
          <router-link to="/static-analysis" class="btn btn-success btn-lg px-4">
            <i class="bi bi-diagram-3 me-2"></i>程序调用静态分析
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'WelcomePage'
}
</script>

<style scoped>
.welcome-page {
  padding-top: 2rem;
  padding-bottom: 4rem;
}

.feature-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  font-size: 1.5rem;
}

.highlighted-container {
  transition: var(--transition);
  border: none;
  box-shadow: var(--box-shadow);
}

.highlighted-container:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
}
</style> 