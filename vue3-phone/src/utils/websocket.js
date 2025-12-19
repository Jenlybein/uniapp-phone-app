/**
 * WebSocket 客户端工具
 * 用于与电脑端建立实时通信连接
 */

// WebSocket 配置
const WS_CONFIG = {
  // WebSocket 服务器地址（实际项目中需要替换为真实地址）
  SERVER_URL: "ws://localhost:8080",
  // 重连间隔时间（毫秒）
  RECONNECT_INTERVAL: 3000,
  // 最大重连次数
  MAX_RECONNECT_ATTEMPTS: 5,
};

class WebSocketClient {
  constructor() {
    this.socket = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.eventListeners = {
      connect: [],
      message: [],
      disconnect: [],
      error: [],
    };
  }

  /**
   * 连接到 WebSocket 服务器
   */
  connect() {
    try {
      // 关闭现有连接
      if (this.socket) {
        this.socket.close();
      }

      // 创建新的 WebSocket 连接
      this.socket = new WebSocket(WS_CONFIG.SERVER_URL);

      // 连接打开事件
      this.socket.onopen = () => {
        console.log("WebSocket 连接已建立");
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.notifyListeners("connect");
      };

      // 接收消息事件
      this.socket.onmessage = (event) => {
        try {
          // 检查消息类型，处理字符串和Blob
          const processMessage = (messageData) => {
            const data = JSON.parse(messageData);
            console.log("收到 WebSocket 消息:", data);
            this.notifyListeners("message", data);
          };

          if (typeof event.data === "string") {
            // 直接处理字符串消息
            processMessage(event.data);
          } else if (event.data instanceof Blob) {
            // 处理Blob消息
            const reader = new FileReader();
            reader.onload = (e) => {
              processMessage(e.target.result);
            };
            reader.readAsText(event.data);
          } else {
            throw new Error(`不支持的消息类型: ${typeof event.data}`);
          }
        } catch (error) {
          console.error("解析 WebSocket 消息失败:", error);
          this.notifyListeners("error", error);
        }
      };

      // 连接关闭事件
      this.socket.onclose = () => {
        console.log("WebSocket 连接已关闭");
        this.isConnected = false;
        this.notifyListeners("disconnect");
        this.attemptReconnect();
      };

      // 连接错误事件
      this.socket.onerror = (error) => {
        console.error("WebSocket 连接错误:", error);
        this.notifyListeners("error", error);
      };
    } catch (error) {
      console.error("创建 WebSocket 连接失败:", error);
      this.notifyListeners("error", error);
      this.attemptReconnect();
    }
  }

  /**
   * 尝试重新连接
   */
  attemptReconnect() {
    if (this.reconnectAttempts < WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
      this.reconnectAttempts++;
      console.log(
        `尝试重新连接 (${this.reconnectAttempts}/${WS_CONFIG.MAX_RECONNECT_ATTEMPTS})...`
      );
      setTimeout(() => {
        this.connect();
      }, WS_CONFIG.RECONNECT_INTERVAL);
    } else {
      console.error("WebSocket 重连失败，已达到最大重试次数");
    }
  }

  /**
   * 发送消息到服务器
   * @param {Object} data - 要发送的数据
   * @returns {boolean} - 是否发送成功
   */
  send(data) {
    if (this.isConnected && this.socket) {
      try {
        this.socket.send(JSON.stringify(data));
        return true;
      } catch (error) {
        console.error("发送 WebSocket 消息失败:", error);
        this.notifyListeners("error", error);
        return false;
      }
    } else {
      console.error("WebSocket 未连接，无法发送消息");
      return false;
    }
  }

  /**
   * 关闭 WebSocket 连接
   */
  disconnect() {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
      this.isConnected = false;
      this.reconnectAttempts = 0;
    }
  }

  /**
   * 注册事件监听器
   * @param {string} event - 事件名称
   * @param {Function} callback - 回调函数
   */
  on(event, callback) {
    if (this.eventListeners[event]) {
      this.eventListeners[event].push(callback);
    }
  }

  /**
   * 移除事件监听器
   * @param {string} event - 事件名称
   * @param {Function} callback - 回调函数
   */
  off(event, callback) {
    if (this.eventListeners[event]) {
      this.eventListeners[event] = this.eventListeners[event].filter(
        (listener) => listener !== callback
      );
    }
  }

  /**
   * 通知所有监听器
   * @param {string} event - 事件名称
   * @param {*} data - 事件数据
   */
  notifyListeners(event, data) {
    if (this.eventListeners[event]) {
      this.eventListeners[event].forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error("执行 WebSocket 事件监听器失败:", error);
        }
      });
    }
  }

  /**
   * 获取连接状态
   * @returns {boolean} - 是否已连接
   */
  getConnectionStatus() {
    return this.isConnected;
  }
}

// 创建单例实例
const wsClient = new WebSocketClient();

export default wsClient;
