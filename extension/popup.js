// Popup script - 管理擴展設定和狀態
document.addEventListener('DOMContentLoaded', async () => {
  const apiUrlInput = document.getElementById('apiUrl');
  const saveBtn = document.getElementById('saveBtn');
  const statusDot = document.getElementById('statusDot');
  const statusText = document.getElementById('statusText');
  const apiAddress = document.getElementById('apiAddress');
  const alertBox = document.getElementById('alertBox');
  const queueSection = document.getElementById('queueSection');
  const queueList = document.getElementById('queueList');
  const queueCount = document.getElementById('queueCount');

  let refreshInterval;
  let currentApiUrl;

  // 加載已保存的設定
  const result = await chrome.storage.sync.get(['apiUrl']);
  const savedApiUrl = result.apiUrl || 'http://localhost:18080';
  currentApiUrl = savedApiUrl;
  apiUrlInput.value = savedApiUrl;
  apiAddress.textContent = savedApiUrl;

  // 檢查服務器狀態
  async function checkServerStatus() {
    try {
      const response = await fetch(`${currentApiUrl}/api/health`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000)
      });

      if (response.ok) {
        statusDot.classList.remove('offline');
        statusText.textContent = '在線';
        return true;
      }
    } catch (error) {
      // Ignore error
    }

    statusDot.classList.add('offline');
    statusText.textContent = '離線';
    return false;
  }

  // 獲取下載隊列
  async function fetchQueue() {
    try {
      const response = await fetch(`${currentApiUrl}/api/tasks`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000)
      });

      if (response.ok) {
        const data = await response.json();
        renderQueue(data);
        return true;
      }
    } catch (error) {
      console.error('Failed to fetch queue:', error);
    }
    return false;
  }

  // 渲染隊列
  function renderQueue(data) {
    const tasks = data.tasks || [];
    
    if (tasks.length === 0) {
      queueSection.style.display = 'none';
      return;
    }

    queueSection.style.display = 'block';
    
    // 計算隊列中和下載中的任務數
    const activeCount = tasks.filter(t => 
      t.status === 'queued' || t.status === 'downloading'
    ).length;
    queueCount.textContent = activeCount;

    // 渲染任務列表
    if (tasks.length === 0) {
      queueList.innerHTML = '<div class="queue-empty">暫無下載任務</div>';
    } else {
      queueList.innerHTML = tasks.map(task => {
        const statusText = getStatusText(task.status);
        const statusClass = `status-${task.status}`;
        const url = extractVideoId(task.url);
        const time = formatTime(task.created_at);
        
        return `
          <div class="queue-item">
            <div class="queue-item-header">
              <span class="queue-item-status ${statusClass}">${statusText}</span>
              <span class="queue-item-time">${time}</span>
            </div>
            <div class="queue-item-url" title="${task.url}">${url}</div>
            ${task.error ? `<div style="color: #fca5a5; font-size: 11px; margin-top: 5px;">${task.error}</div>` : ''}
          </div>
        `;
      }).join('');
    }
  }

  // 獲取狀態文本
  function getStatusText(status) {
    const statusMap = {
      'queued': '⏳ 排隊中',
      'downloading': '⬇️ 下載中',
      'completed': '✅ 已完成',
      'failed': '❌ 失敗'
    };
    return statusMap[status] || status;
  }

  // 提取視頻 ID
  function extractVideoId(url) {
    const match = url.match(/\/([^/]+)\/?$/);
    return match ? match[1] : url;
  }

  // 格式化時間
  function formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000); // 秒

    if (diff < 60) return '剛刚';
    if (diff < 3600) return `${Math.floor(diff / 60)} 分鐘前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} 小時前`;
    return `${Math.floor(diff / 86400)} 天前`;
  }

  // 啟動自動刷新
  function startAutoRefresh() {
    if (refreshInterval) clearInterval(refreshInterval);
    
    refreshInterval = setInterval(async () => {
      const isOnline = await checkServerStatus();
      if (isOnline) {
        await fetchQueue();
      }
    }, 3000); // 每 3 秒刷新一次
  }

  // 停止自動刷新
  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }

  // 顯示提示
  function showAlert(message, type = 'success') {
    alertBox.innerHTML = `<div class="alert ${type}">${message}</div>`;
    setTimeout(() => {
      alertBox.innerHTML = '';
    }, 3000);
  }

  // 保存設定
  saveBtn.addEventListener('click', async () => {
    const apiUrl = apiUrlInput.value.trim();
    
    if (!apiUrl) {
      showAlert('請輸入 API 地址', 'error');
      return;
    }

    try {
      // 驗證 URL 格式
      new URL(apiUrl);
      
      // 保存到 storage
      await chrome.storage.sync.set({ apiUrl });
      currentApiUrl = apiUrl;
      apiAddress.textContent = apiUrl;
      
      showAlert('設定已保存', 'success');
      
      // 重新檢查狀態
      const isOnline = await checkServerStatus();
      if (isOnline) {
        await fetchQueue();
        startAutoRefresh();
      }
    } catch (error) {
      showAlert('無效的 URL 格式', 'error');
    }
  });

  // 初始狀態檢查
  const isOnline = await checkServerStatus();
  if (isOnline) {
    await fetchQueue();
    startAutoRefresh();
  }

  // 清理
  window.addEventListener('beforeunload', () => {
    stopAutoRefresh();
  });
});
