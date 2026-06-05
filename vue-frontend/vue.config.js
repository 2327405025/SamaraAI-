module.exports = {
  devServer: {
    port: 8080,
    historyApiFallback: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9090',
        changeOrigin: true,
        pathRewrite: { '^/api': '/api/v1' }
      }
    }
  }
}
