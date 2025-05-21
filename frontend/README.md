# VaTransfer 前端

VaTransfer是一个简单高效的文件传输工具，使用Vue 3、Vite和Tailwind CSS构建。

## 功能特点

- 文件和文件夹上传
- 拖放支持
- 上传进度实时显示
- 支持普通上传和预签名直传两种模式
- 生成文件分享链接
- 响应式设计

## 快速开始

### 安装依赖

```bash
cd frontend
npm install
```

### 开发模式

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

## 目录结构

```
/frontend
  /public          # 静态资源
  /src
    /assets        # 样式资源
    /components    # Vue组件
    /composables   # 组合式API
    /utils         # 工具函数
    App.vue        # 主应用组件
    main.js        # 入口文件
  index.html       # HTML入口
  package.json     # 依赖
  vite.config.js   # Vite配置
  tailwind.config.js # Tailwind配置
```

## 主要组件

- **FileUploader**: 处理文件选择和上传
- **FileList**: 显示已选择的文件列表
- **FileItem**: 单个文件项组件，显示上传状态和进度
- **Toast**: 通知组件

## 技术栈

- Vue 3 (使用Composition API)
- Vite
- Tailwind CSS v3
- Font Awesome 图标