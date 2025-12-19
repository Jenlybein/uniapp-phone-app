// WebSocket 测试服务器
// 用于测试手机应用接收电脑端发送的图片和信息

const WebSocket = require('ws');
const http = require('http');
const fs = require('fs');
const path = require('path');

// 创建 HTTP 服务器，用于提供 test-websocket.html
const server = http.createServer((req, res) => {
    if (req.url === '/') {
        // 读取 test-websocket.html 文件
        fs.readFile(path.join(__dirname, 'test-websocket.html'), (err, data) => {
            if (err) {
                res.writeHead(500);
                res.end('Error loading test-websocket.html');
                return;
            }
            res.writeHead(200, { 'Content-Type': 'text/html' });
            res.end(data);
        });
    } else {
        res.writeHead(404);
        res.end('Not found');
    }
});

// 创建 WebSocket 服务器
const wss = new WebSocket.Server({ server });

// 客户端连接集合
const clients = new Set();

// 监听 WebSocket 连接
wss.on('connection', (ws) => {
    console.log('客户端已连接');
    clients.add(ws);

    // 监听客户端消息
    ws.on('message', (message) => {
        console.log('收到消息:', message);
        
        // 广播消息给所有客户端
        clients.forEach((client) => {
            if (client !== ws && client.readyState === WebSocket.OPEN) {
                client.send(message);
            }
        });
    });

    // 监听连接关闭
    ws.on('close', () => {
        console.log('客户端已断开连接');
        clients.delete(ws);
    });

    // 监听错误
    ws.on('error', (error) => {
        console.error('WebSocket 错误:', error);
    });
});

// 启动服务器
const PORT = 8080;
server.listen(PORT, () => {
    console.log(`\n=== WebSocket 测试服务器已启动 ===`);
    console.log(`WebSocket 地址: ws://localhost:${PORT}`);
    console.log(`测试页面地址: http://localhost:${PORT}`);
    console.log(`\n使用说明:`);
    console.log(`1. 访问 http://localhost:${PORT} 打开测试页面`);
    console.log(`2. 点击"连接"按钮连接到 WebSocket 服务器`);
    console.log(`3. 输入文本或选择图片，点击发送按钮`);
    console.log(`4. 手机应用将实时接收并显示发送的内容`);
    console.log(`\n按 Ctrl+C 停止服务器`);
});
