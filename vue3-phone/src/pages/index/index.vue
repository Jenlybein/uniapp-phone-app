<template>
	<view class="container">
		<view class="header">
			<text class="title">{{ title }}</text>
		</view>
		
		<!-- 上半区：选项列表区 -->
		<view class="options-section">
			<view class="section-title">选项列表</view>
			
			<!-- 空状态提示 -->
			<view v-if="options.length === 0" class="empty-state">
				<image class="empty-icon" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='60' height='60' viewBox='0 0 24 24' fill='none' stroke='%238E8E93' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cline x1='21' y1='10' x2='3' y2='10'%3E%3C/line%3E%3Cline x1='21' y1='6' x2='3' y2='6'%3E%3C/line%3E%3Cline x1='21' y1='14' x2='3' y2='14'%3E%3C/line%3E%3Cline x1='21' y1='18' x2='3' y2='18'%3E%3C/line%3E%3C/svg%3E"></image>
				<text class="empty-text">暂无选项</text>
				<text class="empty-subtext">等待电脑端传输内容</text>
			</view>
			
			<!-- 选项列表 -->
			<view v-else class="options-list">
				<view v-for="(option, index) in options" :key="index" 
					class="option-item" :class="{ 'selected': selectedIndex === index }"
					@click="selectOption(index)">
					<view class="option-content">
						<text v-if="option.type === 'text'" class="option-text">{{ option.content }}</text>
						<image v-else class="option-image" :src="option.content"></image>
					</view>
					<view class="delete-btn" @click.stop="deleteOption(index)">
						<image class="delete-icon" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='%238E8E93' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cline x1='18' y1='6' x2='6' y2='18'%3E%3C/line%3E%3Cline x1='6' y1='6' x2='18' y2='18'%3E%3C/line%3E%3C/svg%3E"></image>
					</view>
				</view>
			</view>
		</view>
		
		<!-- 下半区：AI回复展示区 -->
		<view class="ai-section">
			<view class="section-title">AI回复</view>
			<view class="ai-content">
				<!-- 流式输出区域 - 始终可见 -->
				<text class="ai-response">{{ aiResponse }}</text>
				
				<!-- 加载状态 - 仅在loading且无内容时显示 -->
				<view v-if="loading && aiResponse === ''" class="loading-container">
					<view class="loading-spinner"></view>
					<text class="loading-text">正在思考...</text>
				</view>
				
				<!-- 空状态 - 仅在非loading且无内容时显示 -->
				<text v-else-if="aiResponse === ''" class="empty-ai">请选择一个选项获取AI回复</text>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import wsClient from '@/utils/websocket';

// 数据定义
const title = ref('Phone Answer');
const options = ref([]);
const selectedIndex = ref(-1);
const aiResponse = ref('');
const loading = ref(false);
const wsStatus = ref('disconnected'); // WebSocket 连接状态

// 处理接收到的内容
const handleReceivedContent = (data) => {
	if (!data || !data.type) {
		console.error('无效的 WebSocket 消息格式:', data);
		return;
	}

	// 根据内容类型添加到选项列表
	if (data.type === 'text' && data.content) {
		options.value.push({
			type: 'text',
			content: data.content
		});
	} else if (data.type === 'image' && data.content) {
		options.value.push({
			type: 'image',
			content: data.content
		});
	} else {
		console.error('不支持的内容类型:', data.type);
	}
};

// 选择选项
const selectOption = async (index) => {
	if (loading.value) return; // 禁止重复触发

	selectedIndex.value = index;
	loading.value = true;
	aiResponse.value = '';

	try {
		const selectedOption = options.value[index];
		const token = uni.getStorageSync('token');
		
		// 检查是否支持fetch API（H5环境）
		if (typeof fetch !== 'undefined') {
			console.log('使用fetch API接收SSE流式响应');
			
			// 使用fetch API发送请求，设置Accept头为text/event-stream
			const response = await fetch('http://localhost:8080/api/ai/chat', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': token,
					'Accept': 'text/event-stream' // 声明支持SSE流式响应
				},
				body: JSON.stringify({
					type: selectedOption.type,
					content: selectedOption.content
				})
			});
			
			// 检查响应状态
			if (!response.ok) {
				throw new Error(`请求失败，状态码：${response.status}`);
			}
			
			// 检查响应是否支持流式处理
			if (!response.body) {
				throw new Error('响应不支持流式处理');
			}
			
			// 获取响应流
			const reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = '';
			
			console.log('开始接收SSE流式响应');
			
			// 隐藏loading状态，显示AI回复区域
			loading.value = false;
			
			// 处理流式响应
			while (true) {
				const { done, value } = await reader.read();
				if (done) {
					console.log('SSE流式响应结束');
					break;
				}
				
				// 解码并处理数据
				buffer += decoder.decode(value, { stream: true });
				
				// 解析SSE格式：data: 内容
				const lines = buffer.split('\n');
				buffer = lines.pop(); // 保存不完整的行
				
				for (const line of lines) {
					if (line.startsWith('data: ')) {
						// 提取AI回复内容
						const chunk = line.slice(6); // 去掉 'data: ' 前缀
						if (chunk) {
							console.log('收到AI回复片段:', chunk);
							aiResponse.value += chunk;
						}
					}
				}
			}
		} else {
			// 不支持fetch API的环境，使用兼容模式（普通HTTP请求）
			console.log('不支持fetch API，使用兼容模式');
			
			const response = await uni.request({
				url: 'http://localhost:8080/api/ai/chat',
				method: 'POST',
				data: {
					type: selectedOption.type,
					content: selectedOption.content
				},
				header: {
					'content-type': 'application/json',
					'Authorization': token
				}
			});
			
			if (response.statusCode === 200 && response.data.code === 200) {
				aiResponse.value = response.data.content;
			} else {
				throw new Error(response.data?.message || 'AI请求失败');
			}
			loading.value = false;
		}
	} catch (error) {
		// 显示iOS风格的错误提示弹窗
		console.error('AI请求出错:', error);
		
		// 显示错误信息
		aiResponse.value = `AI请求失败: ${error.message}`;
		loading.value = false;
	}
};

// 删除选项
const deleteOption = (index) => {
	// 显示iOS风格的确认弹窗
	uni.showModal({
		title: '确认删除',
		content: '确定要删除这个选项吗？',
		confirmText: '删除',
		cancelText: '取消',
		success: (res) => {
			if (res.confirm) {
				options.value.splice(index, 1);
				// 如果删除的是当前选中的选项，重置选中状态
				if (selectedIndex.value === index) {
					selectedIndex.value = -1;
					aiResponse.value = '';
				}
				// 如果删除的是前面的选项，更新选中索引
				else if (selectedIndex.value > index) {
					selectedIndex.value--;
				}
			}
		}
	});
};

// WebSocket 事件监听
const setupWebSocketListeners = () => {
	// 连接成功
	wsClient.on('connect', () => {
		console.log('WebSocket 连接成功');
		wsStatus.value = 'connected';
	});

	// 接收消息
	wsClient.on('message', handleReceivedContent);

	// 连接断开
	wsClient.on('disconnect', () => {
		console.log('WebSocket 连接断开');
		wsStatus.value = 'disconnected';
	});

	// 连接错误
	wsClient.on('error', (error) => {
		console.error('WebSocket 错误:', error);
		wsStatus.value = 'error';
	});
};

// 检查用户是否已登录
const checkLogin = () => {
	const token = uni.getStorageSync('token');
	if (!token) {
		// 未登录，跳转到登录页面
		uni.navigateTo({
			url: '/pages/login/login'
		});
		return false;
	}
	return true;
};

// 页面加载时初始化 WebSocket 连接
onMounted(() => {
	// 检查用户是否已登录
	if (checkLogin()) {
		// 设置 WebSocket 事件监听
		setupWebSocketListeners();
		// 连接 WebSocket 服务器
		wsClient.connect();
	}
});

// 页面卸载时关闭 WebSocket 连接
onUnmounted(() => {
	wsClient.disconnect();
});
</script>

<style scoped>
/* iOS风格全局样式 */
.container {
	width: 100%;
	height: 100vh;
	background-color: #f2f2f7;
	font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Text', 'Helvetica Neue', Helvetica, Arial, sans-serif;
	display: flex;
	flex-direction: column;
	-webkit-font-smoothing: antialiased;
	-moz-osx-font-smoothing: grayscale;
}

/* 区域标题 */
.section-title {
	font-size: 13px;
	font-weight: 600;
	color: #86868b;
	margin: 16px 20px 8px;
	text-transform: uppercase;
	letter-spacing: 0.5px;
	line-height: 1.2;
}

/* 选项列表区 */
.options-section {
	flex: 1;
	overflow: hidden;
}

/* 空状态样式 */
.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 300rpx;
	color: #86868b;
	padding: 0 20px;
}

.empty-icon {
	width: 80px;
	height: 80px;
	margin-bottom: 20px;
	opacity: 0.4;
}

.empty-text {
	font-size: 17px;
	font-weight: 600;
	margin-bottom: 8px;
	color: #1d1d1f;
}

.empty-subtext {
	font-size: 14px;
	opacity: 0.6;
	text-align: center;
	line-height: 1.5;
}

/* 选项列表 */
.options-list {
	background-color: #ffffff;
	border-radius: 12px;
	margin: 0 16px;
	overflow: hidden;
	box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* 选项项样式 */
.option-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 16px;
	border-bottom: 1px solid #f2f2f7;
	transition: background-color 0.15s cubic-bezier(0.4, 0, 0.2, 1);
	position: relative;
	-webkit-tap-highlight-color: transparent;
}

.option-item:last-child {
	border-bottom: none;
}

/* 选中状态 */
.option-item.selected {
	background-color: #e8f0fe;
}

/* 点击反馈 */
.option-item:active {
	background-color: #f2f2f7;
}

.option-item.selected:active {
	background-color: #d0e2fc;
}

/* 选项内容 */
.option-content {
	flex: 1;
	min-width: 0;
}

.option-text {
	font-size: 17px;
	color: #1d1d1f;
	line-height: 1.47059;
	font-weight: 400;
	word-wrap: break-word;
	white-space: pre-wrap;
}

.option-image {
	width: 100px;
	height: 100px;
	border-radius: 12px;
	object-fit: cover;
}

/* 删除按钮 */
.delete-btn {
	width: 44px;
	height: 44px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	transition: background-color 0.15s cubic-bezier(0.4, 0, 0.2, 1);
	margin-left: 8px;
	-webkit-tap-highlight-color: transparent;
}

.delete-btn:active {
	background-color: rgba(0, 0, 0, 0.08);
}

.delete-icon {
	width: 20px;
	height: 20px;
	opacity: 0.5;
	transition: opacity 0.15s ease;
}

.delete-btn:active .delete-icon {
	opacity: 0.8;
}

/* AI回复展示区 */
.ai-section {
	flex: 1;
	overflow: hidden;
	background-color: #ffffff;
	border-top-left-radius: 20px;
	border-top-right-radius: 20px;
	box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.03);
	position: relative;
}

/* AI内容容器 */
.ai-content {
	padding: 20px;
	height: calc(100% - 40px);
	overflow-y: auto;

	/* iOS滚动条样式 */
	::-webkit-scrollbar {
		width: 6px;
	}

	::-webkit-scrollbar-track {
		background-color: transparent;
	}

	::-webkit-scrollbar-thumb {
		background-color: rgba(0, 0, 0, 0.2);
		border-radius: 3px;
	}

	::-webkit-scrollbar-thumb:hover {
		background-color: rgba(0, 0, 0, 0.3);
	}
}

/* 加载动画 */
.loading-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 200rpx;
}

.loading-spinner {
	width: 36px;
	height: 36px;
	border: 3px solid rgba(0, 0, 0, 0.1);
	border-top-color: #007aff;
	border-radius: 50%;
	animation: spin 1s linear infinite;
	margin-bottom: 16px;
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}

.loading-text {
	font-size: 14px;
	color: #86868b;
	font-weight: 400;
	line-height: 1.42857;
}

/* AI回复文本 */
.ai-response {
	font-size: 16px;
	color: #1d1d1f;
	line-height: 1.625;
	font-weight: 400;
	letter-spacing: 0.2px;
	white-space: pre-wrap;
	word-wrap: break-word;
}

/* 空AI回复 */
.empty-ai {
	font-size: 14px;
	color: #86868b;
	text-align: center;
	display: block;
	margin-top: 60px;
	line-height: 1.42857;
	font-weight: 400;
}
</style>
