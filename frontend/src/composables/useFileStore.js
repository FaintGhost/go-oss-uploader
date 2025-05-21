import { ref, computed } from 'vue';
import { formatSize } from '../utils/formatters';

export function useFileStore() {
  // 文件对象存储
  const fileItems = ref(new Map());
  
  // 计算文件总数和总大小
  const fileStats = computed(() => {
    let totalFiles = fileItems.value.size;
    let totalSize = 0;
    let pendingFiles = 0;
    
    fileItems.value.forEach(item => {
      totalSize += item.file.size;
      if (item.status === 'pending') {
        pendingFiles++;
      }
    });
    
    return {
      totalFiles,
      totalSize: formatSize(totalSize),
      pendingFiles
    };
  });

  // 更新文件列表
  function updateFileList(files) {
    // 添加新选择的文件
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const fileName = file.name;
      
      // 检查文件是否已经在列表中
      if (!fileItems.value.has(fileName)) {
        // 存储文件信息
        fileItems.value.set(fileName, {
          file: file,
          status: 'pending', // pending, uploading, success, error
          progress: 0,
          uploadInfo: '-',
          url: ''
        });
      }
    }
  }

  // 更新文件状态
  function updateFileStatus(fileName, status, additionalInfo = {}) {
    const fileItem = fileItems.value.get(fileName);
    if (!fileItem) return;
    
    fileItem.status = status;
    
    // 根据状态更新附加信息
    if (status === 'success') {
      fileItem.url = additionalInfo.url || '';
      if (additionalInfo.uploadTime) {
        fileItem.uploadInfo = `上传用时: ${additionalInfo.uploadTime.toFixed(1)}秒`;
      }
    } else if (status === 'error') {
      fileItem.uploadInfo = additionalInfo.error || '发生错误';
    }
  }

  // 更新文件上传进度
  function updateFileProgress(fileName, progress, transferred, speed) {
    const fileItem = fileItems.value.get(fileName);
    if (!fileItem) return;
    
    fileItem.progress = progress;
    
    // 更新速度显示
    if (speed !== undefined && speed > 0) {
      fileItem.uploadInfo = `速度: ${formatSize(speed)}/s`;
    }
  }

  // 清除文件
  function clearFiles() {
    fileItems.value.clear();
  }

  // 删除特定文件
  function removeFile(fileName) {
    fileItems.value.delete(fileName);
  }

  return {
    fileItems,
    fileStats,
    updateFileList,
    updateFileStatus,
    updateFileProgress,
    clearFiles,
    removeFile
  };
}