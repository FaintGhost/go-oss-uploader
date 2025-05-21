<template>
  <div class="bg-white shadow-lg rounded-xl p-8 flex flex-col h-[600px]">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-medium text-gray-900">文件列表</h2>
      <p 
        class="text-sm px-2 py-1 rounded-md"
        :class="[fileStats.totalFiles > 0 ? 'bg-indigo-50 text-indigo-600 font-medium' : 'bg-gray-50 text-gray-500']">
        {{ fileStats.totalFiles > 0 ? `已选择 ${fileStats.totalFiles} 个文件 (${fileStats.totalSize})` : '未选择文件' }}
      </p>
    </div>
    
    <!-- 无文件时显示 -->
    <div v-if="files.size === 0" class="flex-grow flex items-center justify-center text-center text-gray-500">
      <div>
        <i class="fas fa-inbox text-4xl mb-2"></i>
        <p>暂无选择的文件</p>
      </div>
    </div>
    
    <!-- 文件列表容器 -->
    <div v-else class="flex-grow overflow-y-auto pr-2 space-y-4">
      <FileItem
        v-for="[fileName, fileItem] in fileItemsArray"
        :key="fileName"
        :file-name="fileName"
        :file-item="fileItem"
        @generate-link="handleGenerateLink"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import FileItem from './FileItem.vue';
import { useUploader } from '../composables/useUploader';

const props = defineProps({
  files: {
    type: Map,
    required: true
  }
});

const emit = defineEmits(['update-file-status', 'update-file-progress']);

// 文件统计信息
const fileStats = computed(() => {
  let totalFiles = props.files.size;
  let totalSize = 0;
  
  props.files.forEach(item => {
    if (item.file && item.file.size) {
      totalSize += item.file.size;
    }
  });
  
  // 格式化总大小
  let formattedSize = '0 B';
  if (totalSize === 0) {
    formattedSize = '0 B';
  } else if (totalSize < 1024) {
    formattedSize = totalSize.toFixed(2) + ' B';
  } else if (totalSize < 1024 * 1024) {
    formattedSize = (totalSize / 1024).toFixed(2) + ' KB';
  } else if (totalSize < 1024 * 1024 * 1024) {
    formattedSize = (totalSize / (1024 * 1024)).toFixed(2) + ' MB';
  } else {
    formattedSize = (totalSize / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
  }
  
  return {
    totalFiles,
    totalSize: formattedSize
  };
});

// 转换Map为数组以便在模板中循环
const fileItemsArray = computed(() => {
  return Array.from(props.files.entries());
});

// 获取上传功能
const { generateShareLink } = useUploader(
  props.files,
  (fileName, status, additionalInfo) => emit('update-file-status', fileName, status, additionalInfo),
  (fileName, progress, transferred, speed) => emit('update-file-progress', fileName, progress, transferred, speed)
);

// 处理生成链接
async function handleGenerateLink(fileName, fileUrl, expiration) {
  return await generateShareLink(fileName, fileUrl, expiration);
}
</script>