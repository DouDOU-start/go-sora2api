import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      // react-hooks v5 compiler 规则，降为警告避免误报常规异步数据加载模式
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/immutability': 'warn',
      // 允许组件文件同时导出工具函数/常量
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
])
