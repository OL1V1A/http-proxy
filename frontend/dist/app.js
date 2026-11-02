// 网络代理 - 前端逻辑
// 通过 Wails 运行时 window.go / window.runtime 与后端通信

(function () {
  const App = () => window.go.main.App;
  const Runtime = () => window.runtime;

  const $ = (id) => document.getElementById(id);
  const portEl = $('port');
  const addrEl = $('proxyAddr');
  const startBtn = $('startBtn');
  const stopBtn = $('stopBtn');
  const clearBtn = $('clearBtn');
  const statusDot = $('statusDot');
  const statusText = $('statusText');
  const logs = $('logs');
  const msg = $('msg');

  function showMsg(text, type) {
    msg.textContent = text || '';
    msg.className = 'msg' + (type ? ' ' + type : '');
  }

  function addLog(level, text) {
    const line = document.createElement('div');
    line.className = 'log-line ' + (level || 'info');
    const t = document.createElement('span');
    t.className = 't';
    t.textContent = '[' + new Date().toLocaleTimeString() + ']';
    line.appendChild(t);
    line.appendChild(document.createTextNode(text));
    logs.appendChild(line);
    logs.scrollTop = logs.scrollHeight;
  }

  function setRunning(running) {
    if (running) {
      statusDot.className = 'dot running';
      statusText.textContent = '运行中';
      startBtn.disabled = true;
      stopBtn.disabled = false;
    } else {
      statusDot.className = 'dot stopped';
      statusText.textContent = '已停止';
      startBtn.disabled = false;
      stopBtn.disabled = true;
    }
  }

  async function start() {
    const req = {
      host: '',
      port: portEl.value.trim(),
      proxyAddr: addrEl.value.trim(),
    };
    if (!req.port) { showMsg('请填写本地端口', 'err'); return; }
    if (!req.proxyAddr) { showMsg('请填写转发地址', 'err'); return; }

    startBtn.disabled = true;
    showMsg('启动中…', 'info');
    try {
      const addr = await App().Start(req);
      showMsg('已启动: http://' + addr, 'ok');
      addLog('info', '代理服务已启动: http://' + addr);
      setRunning(true);
    } catch (e) {
      const errMsg = (e && e.message) ? e.message : String(e);
      showMsg('启动失败: ' + errMsg, 'err');
      addLog('error', '启动失败: ' + errMsg);
      setRunning(false);
    }
  }

  async function stop() {
    stopBtn.disabled = true;
    showMsg('停止中…', 'info');
    try {
      await App().Stop();
      showMsg('已停止', 'info');
      addLog('info', '代理服务已停止');
      setRunning(false);
    } catch (e) {
      const errMsg = (e && e.message) ? e.message : String(e);
      showMsg('停止失败: ' + errMsg, 'err');
      addLog('error', '停止失败: ' + errMsg);
      setRunning(false);
    }
  }

  startBtn.addEventListener('click', start);
  stopBtn.addEventListener('click', stop);
  clearBtn.addEventListener('click', () => { logs.innerHTML = ''; });

  // 订阅后端推送的日志与状态事件
  function subscribe() {
    try {
      Runtime().EventsOn('log', (data) => {
        if (data && data.level) addLog(data.level, data.msg);
      });
      Runtime().EventsOn('status', (running) => setRunning(!!running));
    } catch (e) { /* ignore */ }
  }

  // 全局错误兜底，避免“无反应”却看不到原因
  window.addEventListener('error', (ev) => {
    showMsg('界面异常: ' + (ev.message || ev.error), 'err');
    addLog('error', '界面异常: ' + (ev.message || ev.error));
  });
  window.addEventListener('unhandledrejection', (ev) => {
    showMsg('调用异常: ' + (ev.reason && ev.reason.message ? ev.reason.message : ev.reason), 'err');
    addLog('error', '调用异常: ' + (ev.reason && ev.reason.message ? ev.reason.message : ev.reason));
  });

  // 等待 Wails 运行时注入完成
  function whenReady(cb) {
    if (window.go && window.go.main && window.go.main.App && window.runtime) {
      cb();
    } else {
      setTimeout(() => whenReady(cb), 50);
    }
  }

  whenReady(async () => {
    subscribe();
    try {
      const cfg = await App().GetConfig();
      portEl.value = cfg.port || '3100';
      addrEl.value = cfg.proxyAddr || '';
      setRunning(await App().IsRunning());
      showMsg('就绪，填写后点击启动', 'info');
    } catch (e) {
      const errMsg = (e && e.message) ? e.message : String(e);
      showMsg('加载失败: ' + errMsg, 'err');
    }
  });
})();
