<template>
  <div class="bg-gray-50 min-h-screen flex flex-col">
    <div class="max-w-7xl w-full mx-auto flex-grow flex flex-col justify-center">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-3xl font-bold text-gray-800">VaTransfer</h1>
        <span class="text-sm text-gray-500">简单高效的文件传输</span>
      </div>
      
      <!-- 双栏布局容器 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- 左侧：上传区域 -->
        <FileUploader 
          @update-files="updateFileList" 
          :files="fileItems"/>
        
        <!-- 右侧：文件列表 -->
        <FileList 
          :files="fileItems" 
          @update-file-status="updateFileStatus"
          @update-file-progress="updateFileProgress"/>
      </div>
    </div>
    
    <!-- 页脚 -->
    <footer class="text-center text-gray-500 text-sm py-4 mt-auto">
      <p>© 2025 VaTransfer. 简单、高效的文件传输工具。</p>
    </footer>

    <!-- 通知组件 -->
    <Toast ref="toast" />
  </div>
</template>

<script setup>
import { ref, provide } from 'vue';
import FileUploader from './components/FileUploader.vue';
import FileList from './components/FileList.vue';
import Toast from './components/Toast.vue';
import { useFileStore } from './composables/useFileStore';

const toast = ref(null);
const { fileItems, updateFileList, updateFileStatus, updateFileProgress } = useFileStore();

// 提供toast实例给所有组件
provide('toast', {
  show: (type, title, message, duration) => {
    toast.value.show(type, title, message, duration);
  }
});
</script>