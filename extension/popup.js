// Popup script - 管理擴展設定和狀態
document.addEventListener('DOMContentLoaded', async () => {
  const apiUrlInput = document.getElementById('apiUrl');
  const saveBtn = document.getElementById('saveBtn');
  const statusDot = document.getElementById('statusDot');
  const statusText = document.getElementById('statusText');
  const apiAddress = document.getElementById('apiAddress');
  const alertBox = document.getElementById('alertBox');

  // 加載已保存的設定
  const result = await chrome.storage.sync.get(['apiUrl']);
  const savedApiUrl = result.apiUrl || 'http://localhost:18080';
  apiUrlInput.value = savedApiUrl;
  apiAddress.textContent = savedApiUrl;

  // 檢查服務器狀態
  async function checkServerStatus() {
    try {
      const response = await fetch(`${savedApiUrl}/api/health`, {
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
      apiAddress.textContent = apiUrl;
      
      showAlert('設定已保存', 'success');
      
      // 重新檢查狀態
      await checkServerStatus();
    } catch (error) {
      showAlert('無效的 URL 格式', 'error');
    }
  });

  // 初始狀態檢查
  await checkServerStatus();
});
