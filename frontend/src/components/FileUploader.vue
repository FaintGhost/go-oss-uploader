<template>
  <div class="bg-white shadow-lg rounded-xl p-8 flex flex-col h-[600px]">
    <!-- 文件选择区域 -->
    <div class="mb-6">
      <label for="file" class="block text-sm font-medium text-gray-700 mb-2">选择要上传的文件</label>
      <div 
        class="mt-1 flex flex-col items-center justify-center px-6 py-6 border-2 border-dashed rounded-lg transition-colors h-48"
        :class="[isDragging ? 'bg-gray-100 border-indigo-300' : 'border-gray-300 hover:bg-gray-50']"
        @dragenter.prevent="handleDragEnter"
        @dragover.prevent="handleDragOver"
        @dragleave.prevent="handleDragLeave"
        @drop.prevent="handleDrop">
        
        <!-- 文件操作按钮 -->
        <div class="flex justify-center space-x-4 mb-2">
          <label for="file" class="relative cursor-pointer px-4 py-2 bg-indigo-600 rounded-md font-medium text-white hover:bg-indigo-700 focus-within:outline-none transition-colors">
            <i class="fas fa-file mr-1"></i><span>选择文件</span>
            <input id="file" ref="fileInput" name="file" type="file" class="sr-only" multiple required @change="handleFileChange">
          </label>
          <label for="folder" class="relative cursor-pointer px-4 py-2 bg-indigo-500 rounded-md font-medium text-white hover:bg-indigo-600 focus-within:outline-none transition-colors">
            <i class="fas fa-folder-open mr-1"></i><span>选择文件夹</span>
            <input id="folder" ref="folderInput" name="folder" type="file" webkitdirectory directory class="sr-only" @change="handleFolderChange">
          </label>
        </div>
        <!-- 拖拽提示文本 -->
        <p class="text-xs text-gray-500">或将文件拖拽到此处上传</p>
      </div>
    </div>
    
    <!-- 上传按钮区域 -->
    <div class="grid grid-cols-2 gap-4 mb-6">
      <button 
        @click="handleNormalUpload" 
        :disabled="isUploading || hasPendingFiles === false"
        :class="[isUploading || hasPendingFiles === false ? 'opacity-50 cursor-not-allowed' : '']"
        class="inline-flex justify-center items-center py-3 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 transition-colors">
        <i class="fas fa-upload mr-2"></i>
        普通上传
      </button>
      <button 
        @click="handlePresignUpload" 
        :disabled="isUploading || hasPendingFiles === false"
        :class="[isUploading || hasPendingFiles === false ? 'opacity-50 cursor-not-allowed' : '']"
        class="inline-flex justify-center items-center py-3 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
        <i class="fas fa-bolt mr-2"></i>
        快速上传
      </button>
    </div>
    
    <!-- 功能说明 -->
    <div class="mt-4 pt-6 border-t border-gray-200 flex-grow overflow-y-auto">
      <div class="flex border-b border-gray-200">
        <button 
          @click="activeTab = 'normal'"
          :class="[activeTab === 'normal' ? 'text-indigo-600 border-b-2 border-indigo-600' : 'text-gray-500 hover:text-gray-700']"
          class="px-4 py-2 text-sm font-medium">
          普通上传
        </button>
        <button 
          @click="activeTab = 'presign'"
          :class="[activeTab === 'presign' ? 'text-indigo-600 border-b-2 border-indigo-600' : 'text-gray-500 hover:text-gray-700']"
          class="px-4 py-2 text-sm font-medium">
          快速上传
        </button>
      </div>
      
      <div class="py-4">
        <div v-if="activeTab === 'normal'" class="text-sm text-gray-600">
          <h3 class="text-lg font-medium text-gray-900 mb-2">普通上传模式</h3>
          <p class="mb-2">文件先上传到服务器，然后由服务器上传至OSS。适用于小文件，服务器需要处理文件的场景。</p>
        </div>
        
        <div v-if="activeTab === 'presign'" class="text-sm text-gray-600">
          <h3 class="text-lg font-medium text-gray-900 mb-2">快速上传模式</h3>
          <p class="mb-2">客户端直接上传文件到OSS，不经过服务器中转。适用于大文件，减轻服务器负担。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue';
import { useUploader } from '../composables/useUploader';
import { formatSize } from '../utils/formatters';

const props = defineProps({
  files: {
    type: Map,
    required: true
  }
});

const emit = defineEmits(['update-files']);

// 文件输入引用
const fileInput = ref(null);
const folderInput = ref(null);

// 拖拽状态
const isDragging = ref(false);

// 活动标签
const activeTab = ref('normal');

// 上传状态
const isUploading = ref(false);

// 有无待上传文件
const hasPendingFiles = computed(() => {
  return Array.from(props.files.values()).some(item => item.status === 'pending');
});

// 获取toast
const toast = inject('toast');

// 使用上传组合式API
const { handleNormalUpload: startNormalUpload, handlePresignUpload: startPresignUpload } = useUploader(
  props.files,
  (fileName, status, additionalInfo) => emit('update-file-status', fileName, status, additionalInfo),
  (fileName, progress, transferred, speed) => emit('update-file-progress', fileName, progress, transferred, speed)
);

// 文件选择处理
function handleFileChange(event) {
  if (event.target.files.length > 0) {
    emit('update-files', event.target.files);
    
    // 显示通知
    const files = event.target.files;
    let totalSize = 0;
    for (let i = 0; i < files.length; i++) {
      totalSize += files[i].size;
    }
    
    toast.show('info', '文件已选择', `已添加 ${files.length} 个文件 (${formatSize(totalSize)})`);
  }
}

// 文件夹选择处理
function handleFolderChange(event) {
  if (event.target.files.length > 0) {
    // 过滤掉文件夹项
    const files = Array.from(event.target.files).filter(file => !file.name.endsWith('/'));
    
    if (files.length === 0) {
      toast.show('error', '文件夹为空', '选择的文件夹中没有可上传的文件');
      return;
    }
    
    // 获取文件夹名称
    const folderPath = files[0].webkitRelativePath || '';
    const folderName = folderPath.split('/')[0];
    
    // 计算总大小
    let totalSize = 0;
    for (let i = 0; i < files.length; i++) {
      totalSize += files[i].size;
    }
    
    // 添加到文件列表
    emit('update-files', files);
    
    // 显示通知
    toast.show('info', '文件夹选择成功', `已从文件夹"${folderName}"中添加 ${files.length} 个文件 (${formatSize(totalSize)})`);
    
    // 清空文件夹选择器，以便可以再次选择同一文件夹
    folderInput.value.value = '';
  }
}

// 拖放处理
function handleDragEnter(e) {
  isDragging.value = true;
}

function handleDragOver(e) {
  isDragging.value = true;
}

function handleDragLeave(e) {
  isDragging.value = false;
}

function handleDrop(e) {
  isDragging.value = false;
  
  const dt = e.dataTransfer;
  const files = dt.files;
  
  if (files.length > 0) {
    emit('update-files', files);
    
    // 显示通知
    let totalSize = 0;
    for (let i = 0; i < files.length; i++) {
      totalSize += files[i].size;
    }
    
    toast.show('info', '文件已添加', `已拖放添加 ${files.length} 个文件 (${formatSize(totalSize)})`);
  }
}

// 上传处理
async function handleNormalUpload() {
  if (isUploading.value || !hasPendingFiles.value) return;
  
  isUploading.value = true;
  try {
    await startNormalUpload();
  } finally {
    isUploading.value = false;
  }
}

async function handlePresignUpload() {
  if (isUploading.value || !hasPendingFiles.value) return;
  
  isUploading.value = true;
  try {
    await startPresignUpload();
  } finally {
    isUploading.value = false;
  }
}
</script>