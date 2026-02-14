// Content script - 在 Jable 視頻頁面注入下載按鈕
(function() {
  'use strict';

  // 檢查是否已經注入
  if (window.jableDownloaderInjected) {
    return;
  }
  window.jableDownloaderInjected = true;

  // 創建下載按鈕
  function createDownloadButton() {
    const button = document.createElement('button');
    button.id = 'jable-download-btn';
    button.className = 'jable-download-button';
    button.innerHTML = `
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
        <polyline points="7 10 12 15 17 10"></polyline>
        <line x1="12" y1="15" x2="12" y2="3"></line>
      </svg>
      <span>下載影片</span>
    `;
    
    button.addEventListener('click', handleDownload);
    return button;
  }

  // 處理下載
  async function handleDownload(event) {
    event.preventDefault();
    const button = event.currentTarget;
    const videoUrl = window.location.href;

    try {
      // 更新按鈕狀態
      button.disabled = true;
      button.innerHTML = `
        <span class="spinner"></span>
        <span>正在發送...</span>
      `;

      // 獲取 API 服務器地址
      const result = await chrome.storage.sync.get(['apiUrl']);
      const apiUrl = result.apiUrl || 'http://localhost:18080';

      // 發送下載請求
      const response = await fetch(`${apiUrl}/api/download`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: videoUrl,
          convert: false
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      
      if (data.success) {
        // 成功
        button.innerHTML = `
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          <span>已加入下載隊列</span>
        `;
        button.style.backgroundColor = '#10b981';

        // 3秒後恢復
        setTimeout(() => {
          button.disabled = false;
          button.style.backgroundColor = '';
          button.innerHTML = `
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
            <span>下載影片</span>
          `;
        }, 3000);
      } else {
        throw new Error(data.message || '下載失敗');
      }
    } catch (error) {
      // 顯示錯誤
      button.innerHTML = `
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="15" y1="9" x2="9" y2="15"></line>
          <line x1="9" y1="9" x2="15" y2="15"></line>
        </svg>
        <span>${error.message.includes('Failed to fetch') ? '服務器未啟動' : '下載失敗'}</span>
      `;
      button.style.backgroundColor = '#ef4444';

      // 5秒後恢復
      setTimeout(() => {
        button.disabled = false;
        button.style.backgroundColor = '';
        button.innerHTML = `
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
          <span>下載影片</span>
        `;
      }, 5000);
    }
  }

  // 注入按鈕到頁面
  function injectButton() {
    // 如果按鈕已存在，不重複添加
    if (document.getElementById('jable-download-btn')) {
      return;
    }

    // 優先放在 info-header 下的 div.models 後面（最佳位置）
    const infoHeader = document.querySelector('.info-header');
    if (infoHeader) {
      const modelsDiv = infoHeader.querySelector('div.models');
      if (modelsDiv) {
        const button = createDownloadButton();
        const container = document.createElement('div');
        container.className = 'jable-download-container';
        container.appendChild(button);
        
        // 插入到 div.models 後面
        modelsDiv.parentNode.insertBefore(container, modelsDiv.nextSibling);
        return;
      }
    }

    // 備用方案 1：嘗試其他位置
    const selectors = [
      '.info-header',
      '.video-info-header',
      'h4.title',
      '.video-title',
      'h1',
      '.video-detail h2',
      '.title-box',
      '.detail-box .title'
    ];

    let targetElement = null;
    for (const selector of selectors) {
      targetElement = document.querySelector(selector);
      if (targetElement) {
        break;
      }
    }

    if (!targetElement) {
      // 備用方案 2：懸浮按鈕（右上角）
      const button = createDownloadButton();
      const container = document.createElement('div');
      container.className = 'jable-download-container';
      container.style.cssText = 'position: fixed; top: 80px; right: 20px; z-index: 9999;';
      container.appendChild(button);
      document.body.insertBefore(container, document.body.firstChild);
      return;
    }

    // 正常插入到找到的元素後
    const button = createDownloadButton();
    const container = document.createElement('div');
    container.className = 'jable-download-container';
    container.appendChild(button);
    
    targetElement.parentNode.insertBefore(container, targetElement.nextSibling);
  }

  // 等待頁面加載完成
  function init() {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', () => {
        setTimeout(injectButton, 500);
      });
    } else {
      setTimeout(injectButton, 500);
    }

    // 監聽 DOM 變化（處理單頁應用）
    const observer = new MutationObserver(() => {
      if (!document.getElementById('jable-download-btn')) {
        injectButton();
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  // 啟動
  init();
})();
