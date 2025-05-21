<template>
  <div 
    class="fixed right-5 top-5 transform transition-transform duration-300 ease-in-out"
    :class="[isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0 invisible']">
    <div class="bg-white shadow-lg rounded-lg p-4 flex items-start max-w-sm">
      <div class="flex-shrink-0" :class="iconColor">
        <i :class="['fas', iconClass, 'text-lg']"></i>
      </div>
      <div class="ml-3 flex-1">
        <p class="text-sm font-medium text-gray-900">{{ title }}</p>
        <p class="mt-1 text-sm text-gray-500">{{ message }}</p>
      </div>
      <button type="button" @click="hide" class="ml-4 flex-shrink-0 text-gray-400 hover:text-gray-500">
        <i class="fas fa-times"></i>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

const isVisible = ref(false);
const type = ref('success');
const title = ref('成功');
const message = ref('操作已完成');
let timeoutId = null;

// 计算图标和颜色
const iconClass = computed(() => {
  switch (type.value) {
    case 'success': return 'fa-check-circle';
    case 'error': return 'fa-exclamation-circle';
    case 'info': return 'fa-info-circle';
    default: return 'fa-check-circle';
  }
});

const iconColor = computed(() => {
  switch (type.value) {
    case 'success': return 'text-green-400';
    case 'error': return 'text-red-400';
    case 'info': return 'text-blue-400';
    default: return 'text-green-400';
  }
});

// 显示通知
function show(toastType, toastTitle, toastMessage, duration = 3000) {
  // 清除之前的超时
  if (timeoutId) {
    clearTimeout(timeoutId);
  }
  
  // 设置内容
  type.value = toastType;
  title.value = toastTitle;
  message.value = toastMessage;
  
  // 显示通知
  isVisible.value = true;
  
  // 设置自动隐藏
  timeoutId = setTimeout(() => {
    hide();
  }, duration);
}

// 隐藏通知
function hide() {
  isVisible.value = false;
}

// 暴露方法给父组件
defineExpose({
  show,
  hide
});
</script>