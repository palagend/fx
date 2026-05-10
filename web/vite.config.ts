import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { compression } from 'vite-plugin-compression2'

export default defineConfig({
  plugins: [
    vue(),
    // Gzip 压缩
    compression({
      algorithms: ['gzip'],
      exclude: [/\.(br)$/, /\.(gz)$/],
      threshold: 1024, // 大于 1KB 的文件才压缩
    }),
    // Brotli 压缩（更好的压缩率）
    compression({
      algorithms: ['brotliCompress'],
      exclude: [/\.(br)$/, /\.(gz)$/],
      threshold: 1024,
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  build: {
    // 代码分割配置
    rollupOptions: {
      output: {
        // 手动代码分割 - 优化策略：减少 chunk 数量，合并相关依赖
        manualChunks: (id) => {
          if (id.includes('node_modules')) {
            // 核心框架打包在一起
            if (id.includes('vue') || id.includes('vue-router') || id.includes('pinia')) {
              return 'vendor-core'
            }
            // UI 相关库合并
            if (id.includes('@iconify') || id.includes('chart.js') || id.includes('vue-chartjs')) {
              return 'vendor-ui'
            }
            // 其他第三方库统一打包
            return 'vendor-others'
          }
        },
        // chunk 文件命名（避免以 _ 开头，GitHub Pages 会忽略）
        chunkFileNames: 'js/chunk-[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const fileName = assetInfo.names?.[0] || ''
          if (/\.(png|jpe?g|gif|svg|webp|ico)$/i.test(fileName)) {
            return 'img/[name]-[hash][extname]'
          }
          if (/\.(css)$/i.test(fileName)) {
            return 'css/[name]-[hash][extname]'
          }
          return 'assets/[name]-[hash][extname]'
        }
      }
    },
    // 压缩配置 - 使用 esbuild 压缩（Vite 默认）
    minify: 'esbuild',
    // 增加资源内联阈值（小于 8KB 的资源内联，减少 HTTP 请求）
    assetsInlineLimit: 8192,
    // 预渲染资源大小警告阈值
    chunkSizeWarningLimit: 500,
    // 启用 CSS 代码分割
    cssCodeSplit: true,
    // 生成 sourcemap（生产环境可关闭）
    sourcemap: false
  },
  // 优化依赖预构建
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'chart.js', 'vue-chartjs', '@iconify/vue'],
    force: true
  }
})
