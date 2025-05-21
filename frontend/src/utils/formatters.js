/**
 * 格式化文件大小
 * @param {number} bytes 文件大小（字节）
 * @returns {string} 格式化后的字符串
 */
export function formatSize(bytes) {
  if (bytes === undefined || isNaN(bytes) || bytes === null) {
    return '0 B';
  }
  
  if (bytes === 0) {
    return '0 B';
  } else if (bytes < 1024) {
    return bytes.toFixed(2) + ' B';
  } else if (bytes < 1024 * 1024) {
    return (bytes / 1024).toFixed(2) + ' KB';
  } else if (bytes < 1024 * 1024 * 1024) {
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  } else {
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
  }
}