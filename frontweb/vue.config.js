const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: [],
  configureWebpack: {
    resolve: {
      alias: {
        vue$: 'vue/dist/vue.esm-bundler.js',
      },
    },
  },
  devServer: {
    proxy: {
      '/api': {
        target: 'http://10.86.32.70:8000',
        changeOrigin: true,
      },
    },
  }
})
