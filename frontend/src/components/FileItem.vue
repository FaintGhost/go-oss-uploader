<template>
  <div class="file-item border border-gray-200 rounded-md p-3 transition">
    <div class="flex justify-between items-start mb-2">
      <div class="flex-1 mr-2">
        <div class="text-sm font-medium text-gray-800 truncate" :title="fileTitle">{{ displayName }}</div>
        <div class="text-xs text-gray-500">{{ formattedSize }}</div>
      </div>
      <div 
        class="px-2 py-1 text-xs font-medium rounded"
        :class="statusClass">
        {{ statusText }}
      </div>
    </div>
    <div class="mb-2">
      <div class="w-full bg-gray-200 rounded-full h-1.5">
        <div 
          class="bg-blue-600 h-1.5 rounded-full" 
          :style="{ width: `${fileItem.progress}%` }">
        </div>
      </div>
      <div class="flex justify-between text-xs mt-1">
        <span>{{ fileItem.progress }}%</span>
        <span>{{ formattedTransferred }} / {{ formattedSize }}</span>
      </div>
    </div>
    <div class="flex justify-between items-center">
      <span class="text-xs text-gray-500">{{ fileItem.uploadInfo }}</span>
      <button 
        @click="toggleShareLink"
        class="text-xs px-2 py-1 rounded"
        :class="[fileItem.status === 'success' ? 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100' : 'bg-gray-50 text-gray-400 cursor-not-allowed']">
        <i class="fas fa-link mr-1"></i>生成链接
      </button>
    </div>
    
    <!-- 分享链接区域 -->
    <div v-if="showShareLink" class="share-link-area mt-2 pt-2 border-t border-gray-100">
      <div class="flex items-center space-x-2 mb-2">
        <label class="text-xs text-gray-600">有效期:</label>
        <select v-model="expirationOption" class="text-xs border border-gray-300 rounded px-1 py-0.5 bg-white">
          <option value="3h">3小时</option>
          <option value="24h">1天</option>
          <option value="72h">3天</option>
          <option value="168h">7天</option>
        </select>
      </div>
      <div class="flex rounded overflow-hidden border border-gray-300">
        <input 
          type="text" 
          v-model="shortLink"
          readonly 
          class="flex-1 text-xs px-2 py-1 border-none focus:ring-0" 
          :placeholder="linkPlaceholder">
        <button 
          @click="copyShareLink"
          class="px-2 py-1 text-xs bg-gray-100 text-gray-600 hover:bg-gray-200 transition"
          :disabled="!shortLink">
          <i class="fas fa-copy"></i>
        </button>
      </div>
      <div class="text-xs text-gray-500 mt-1">{{ expirationText }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue';
import { formatSize } from '../utils/formatters';

const props = defineProps({
  fileName: {
    type: String,
    required: true
  },
  fileItem: {
    type: Object,
    required: true
  }
});

const emit = defineEmits(['generate-link']);

// 使用toast通知
const toast = inject('toast');

// 显示分享链接区域
const showShareLink = ref(false);

// 有效期选项
const expirationOption = ref('24h');

// 短链接
const shortLink = ref('');

// 是否处于加载状态
const isLoading = ref(false);

// 过期时间文本
const expirationText = ref('链接有效期: -');

// 链接占位符文本
const linkPlaceholder = computed(() => {
  if (isLoading.value) return '正在生成链接...';
  if (!showShareLink.value) return '点击生成链接按钮生成分享链接';
  return '生成链接失败';
});

// 计算文件显示名称
const displayName = computed(() => {
  return props.fileName.split('/').pop();
});

// 计算文件完整路径显示
const fileTitle = computed(() => {
  return props.fileItem.file.webkitRelativePath 
    ? props.fileItem.file.webkitRelativePath 
    : props.fileName;
});

// 格式化文件大小
const formattedSize = computed(() => {
  return formatSize(props.fileItem.file.size);
});

// 格式化已上传大小
const formattedTransferred = computed(() => {
  // 使用progress百分比和文件大小估算已上传大小
  const transferred = props.fileItem.file.size * (props.fileItem.progress / 100);
  return formatSize(transferred);
});

// 计算状态显示文本
const statusText = computed(() => {
  switch (props.fileItem.status) {
    case 'pending': return '等待上传';
    case 'uploading': return '上传中';
    case 'success': return '上传成功';
    case 'error': return '上传失败';
    default: return '等待上传';
  }
});

// 计算状态class
const statusClass = computed(() => {
  switch (props.fileItem.status) {
    case 'pending': return 'bg-gray-100 text-gray-800';
    case 'uploading': return 'bg-blue-100 text-blue-800';
    case 'success': return 'bg-green-100 text-green-800';
    case 'error': return 'bg-red-100 text-red-800';
    default: return 'bg-gray-100 text-gray-800';
  }
});

// 切换分享链接区域显示
async function toggleShareLink() {
  // 只有成功上传的文件可以生成链接
  if (props.fileItem.status !== 'success') return;
  
  // 切换显示状态
  showShareLink.value = !showShareLink.value;
  
  // 如果是显示分享区域，则生成链接
  if (showShareLink.value) {
    await generateShareLink();
  }
}

// 生成分享链接
async function generateShareLink() {
  if (!props.fileItem.url) return;
  
  isLoading.value = true;
  shortLink.value = '';
  expirationText.value = '链接有效期: 加载中...';
  
  try {
    // 调用父组件方法生成链接
    const result = await emit('generate-link', props.fileName, props.fileItem.url, expirationOption.value);
    
    if (result && result[0]) { // 获取第一个结果
      const linkData = result[0];
      shortLink.value = linkData.shortURL;
      
      // 更新过期时间
      const expDate = new Date(linkData.expiration);
      expirationText.value = `链接有效期至: ${expDate.toLocaleString()}`;
    }
  } catch (error) {
    console.error('生成链接失败:', error);
    expirationText.value = `错误: ${error.message}`;
  } finally {
    isLoading.value = false;
  }
}

// 监听有效期变化
async function watchExpirationChange() {
  // 如果分享区域显示且已经有链接，则刷新链接
  if (showShareLink.value && shortLink.value) {
    await generateShareLink();
  }
}

// 复制分享链接
function copyShareLink() {
  if (!shortLink.value) return;
  
  navigator.clipboard.writeText(shortLink.value)
    .then(() => {
      toast.show('success', '已复制', '链接已复制到剪贴板');
    })
    .catch(err => {
      console.error('复制链接失败:', err);
      toast.show('error', '复制失败', '请手动复制链接');
    });
}

// 监听有效期变化
import { watch } from 'vue';
watch(expirationOption, watchExpirationChange);
</script>