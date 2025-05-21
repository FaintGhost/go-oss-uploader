import { inject } from 'vue';
import { formatSize } from '../utils/formatters';

export function useUploader(fileItems, updateFileStatus, updateFileProgress) {
  const toast = inject('toast');
  
  // 普通上传
  async function handleNormalUpload() {
    // 将Map转换为数组并过滤出待上传文件
    const pendingFiles = Array.from(fileItems.value.values()).filter(item => item.status === 'pending');
    
    if (pendingFiles.length === 0) {
      toast.show('error', '错误', '没有待上传的文件');
      return;
    }
    
    try {
      // 逐个上传文件
      for (const fileItem of pendingFiles) {
        const file = fileItem.file;
        const fileName = file.name;
        const fileStartTime = new Date();
        
        // 更新状态为上传中
        updateFileStatus(fileName, 'uploading');
        
        try {
          // 生成唯一的上传ID
          const uploadID = 'upload_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
          
          // 连接WebSocket获取进度更新
          const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
          const wsUrl = `${protocol}//${window.location.host}/ws/progress/${uploadID}`;
          const socket = new WebSocket(wsUrl);
          
          // 创建一个Promise来处理WebSocket进度更新
          const uploadPromise = new Promise((resolve, reject) => {
            let lastTransferred = 0;
            
            socket.onopen = function() {
              console.log('WebSocket连接已建立', uploadID);
              // 初始化显示，在连接建立时就显示"准备上传"
              const fileItem = fileItems.value.get(fileName);
              if (fileItem) {
                fileItem.uploadInfo = "准备上传...";
              }
            };
            
            socket.onmessage = function(event) {
              const progress = JSON.parse(event.data);
              const increment = progress.transferred - lastTransferred;
              lastTransferred = progress.transferred;
              
              // 更新文件进度
              const percent = Math.min(100, Math.round((progress.transferred / progress.total) * 100));
              updateFileProgress(fileName, percent, progress.transferred, progress.speed);
            };
            
            socket.onerror = function(error) {
              console.error('WebSocket错误:', error);
              reject(new Error('WebSocket连接错误'));
            };
            
            // 准备表单数据
            const formData = new FormData();
            formData.append('file', file);
            formData.append('uploadID', uploadID);
            formData.append('originalFileName', fileName); // 添加原始文件名
            
            // 发送上传请求
            fetch('/upload', {
              method: 'POST',
              body: formData
            })
            .then(response => response.json())
            .then(data => {
              socket.close();
              resolve({
                result: data,
                uploadTime: (new Date() - fileStartTime) / 1000
              });
            })
            .catch(error => {
              socket.close();
              reject(error);
            });
          });
          
          // 等待当前文件上传完成
          const result = await uploadPromise;
          const uploadTime = result.uploadTime;
          
          // 更新状态为成功
          updateFileStatus(fileName, 'success', {
            uploadTime: uploadTime,
            url: result.result.url || ''
          });
          
          toast.show('success', '上传成功', `${fileName} 上传完成`);
        } catch (error) {
          // 更新状态为失败
          updateFileStatus(fileName, 'error', {
            error: error.message || '上传失败'
          });
          
          toast.show('error', '上传失败', `${fileName}: ${error.message || '上传失败'}`);
        }
      }
    } catch (error) {
      console.error('上传过程中发生错误:', error);
      toast.show('error', '上传失败', error.message || '上传过程中发生错误');
    }
  }
  
  // 预签名上传
  async function handlePresignUpload() {
    // 将Map转换为数组并过滤出待上传文件
    const pendingFiles = Array.from(fileItems.value.values()).filter(item => item.status === 'pending');
    
    if (pendingFiles.length === 0) {
      toast.show('error', '错误', '没有待上传的文件');
      return;
    }
    
    try {
      // 逐个上传文件
      for (const fileItem of pendingFiles) {
        const file = fileItem.file;
        const fileName = file.name;
        const fileStartTime = new Date();
        
        // 更新状态为上传中
        updateFileStatus(fileName, 'uploading');
        
        // 显示准备上传
        const item = fileItems.value.get(fileName);
        if (item) {
          item.uploadInfo = "准备上传...";
        }

        try {
          // 步骤1：获取预签名URL
          const formData = new FormData();
          formData.append('fileName', file.name);
          formData.append('fileSize', file.size);
          
          const presignResponse = await fetch('/presign', {
            method: 'POST',
            body: formData
          });
          
          if (!presignResponse.ok) {
            throw new Error('获取预签名URL失败');
          }
          
          const presignData = await presignResponse.json();
          
          // 步骤2：使用预签名URL直接上传到OSS
          // 创建XHR对象以便跟踪进度
          const xhr = new XMLHttpRequest();
          
          // 使用Promise包装XHR
          const uploadPromise = new Promise((resolve, reject) => {
            // 速度计算变量
            let lastLoaded = 0;
            let lastTime = Date.now();
            let speed = 0;
            
            // 监听上传进度
            xhr.upload.onprogress = function(e) {
              if (e.lengthComputable) {
                // 计算速度
                const now = Date.now();
                const elapsedMs = now - lastTime;
                if (elapsedMs > 100) { // 至少100毫秒更新一次速度
                  const loadedDiff = e.loaded - lastLoaded;
                  speed = (loadedDiff / elapsedMs) * 1000; // 字节/秒
                  lastLoaded = e.loaded;
                  lastTime = now;
                }
                
                const percent = Math.round((e.loaded / e.total) * 100);
                updateFileProgress(fileName, percent, e.loaded, speed);
              }
            };
            
            xhr.open('PUT', presignData.url);
            
            // 设置必要的请求头
            for (const [key, value] of Object.entries(presignData.headers)) {
              xhr.setRequestHeader(key, value);
            }
            
            xhr.onload = function() {
              if (xhr.status >= 200 && xhr.status < 300) {
                resolve({
                  success: true,
                  status: xhr.status,
                  url: `https://${presignData.headers['Host']}/${file.name}`
                });
              } else {
                reject(new Error(`上传失败，状态码: ${xhr.status}`));
              }
            };
            
            xhr.onerror = function() {
              reject(new Error('网络错误'));
            };
            
            xhr.send(file);
          });
          
          // 等待当前文件上传完成
          const result = await uploadPromise;
          const uploadTime = (new Date() - fileStartTime) / 1000;
          
          // 更新状态为成功
          updateFileStatus(fileName, 'success', {
            uploadTime: uploadTime,
            url: result.url || ''
          });
          
          toast.show('success', '上传成功', `${fileName} 上传完成`);
        } catch (error) {
          // 更新状态为失败
          updateFileStatus(fileName, 'error', {
            error: error.message || '上传失败'
          });
          
          toast.show('error', '上传失败', `${fileName}: ${error.message || '上传失败'}`);
        }
      }
    } catch (error) {
      console.error('上传过程中发生错误:', error);
      toast.show('error', '上传失败', error.message || '上传过程中发生错误');
    }
  }
  
  // 生成分享链接
  async function generateShareLink(fileName, fileUrl, expiration) {
    try {
      // 请求预签名下载URL
      const presignResponse = await fetch(`/download/${fileName}?expiration=${expiration}`);
      
      if (!presignResponse.ok) {
        throw new Error('获取下载链接失败');
      }
      
      const presignData = await presignResponse.json();
      
      // 请求将预签名URL转换为短链接
      const formData = new FormData();
      formData.append('url', presignData.url);
      formData.append('fileName', fileName);
      formData.append('expiration', presignData.expiration);
      
      const shortLinkResponse = await fetch('/short-link', {
        method: 'POST',
        body: formData
      });
      
      if (!shortLinkResponse.ok) {
        throw new Error('生成短链接失败');
      }
      
      const shortLinkData = await shortLinkResponse.json();
      
      return {
        shortURL: shortLinkData.shortURL,
        expiration: shortLinkData.expiration
      };
    } catch (error) {
      toast.show('error', '生成链接失败', error.message);
      throw error;
    }
  }
  
  return {
    handleNormalUpload,
    handlePresignUpload,
    generateShareLink
  };
}