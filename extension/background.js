// Background service worker
chrome.runtime.onInstalled.addListener(() => {
  // Extension installed
});

// 處理來自 content script 的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'download') {
    handleDownload(request.url)
      .then(result => sendResponse(result))
      .catch(error => sendResponse({ success: false, error: error.message }));
    return true; // 保持消息通道開放
  }
});

async function handleDownload(url) {
  try {
    const result = await chrome.storage.sync.get(['apiUrl']);
    const apiUrl = result.apiUrl || 'http://localhost:18080';

    const response = await fetch(`${apiUrl}/api/download`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url, convert: false })
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    throw new Error('無法連接到下載服務器');
  }
}
