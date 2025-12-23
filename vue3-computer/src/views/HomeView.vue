<template>
  <div class="home-container">
    <el-header class="home-header">
      <div class="header-left">
        <h1>设备消息发送系统</h1>
      </div>
      <div class="header-right">
        <el-badge :value="messageStore.connected ? '在线' : '离线'" :type="messageStore.connected ? 'success' : 'danger'" class="status-badge">
          <el-text>手机端状态</el-text>
        </el-badge>
        <el-button type="danger" @click="handleLogout" class="logout-button">
          退出登录
        </el-button>
      </div>
    </el-header>
    
    <el-main class="home-main">
      <el-card class="send-card">
        <template #header>
          <div class="card-header">
            <h2>发送消息</h2>
          </div>
        </template>
        
        <div class="send-content">
          <!-- 文本消息发送 -->
          <el-form :model="textForm" ref="textFormRef" label-width="80px">
            <el-form-item label="文本内容" prop="content">
              <el-input
                v-model="textForm.content"
                type="textarea"
                :rows="4"
                placeholder="请输入要发送的文本消息"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleSendText" :loading="messageStore.loading" class="send-button">
                发送文本消息
              </el-button>
            </el-form-item>
          </el-form>
          
          <div class="divider"></div>
          
          <!-- 图片消息发送 -->
          <el-form ref="imageFormRef" label-width="80px">
            <el-form-item label="图片上传">
              <el-upload
                ref="uploadRef"
                class="upload-demo"
                :auto-upload="false"
                :on-change="handleImageChange"
                :show-file-list="true"
                accept="image/*"
              >
                <el-button type="primary">选择图片</el-button>
                <template #tip>
                  <div class="el-upload__tip">
                    只能上传jpg/png文件，且不超过2MB
                  </div>
                </template>
              </el-upload>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleSendImage" :loading="messageStore.loading" class="send-button" :disabled="!selectedImage">
                发送图片消息
              </el-button>
            </el-form-item>
          </el-form>
          
          <!-- WebSocket连接状态 -->
          <div class="ws-status">
            <el-text>
              WebSocket连接状态: 
              <span :class="messageStore.connected ? 'connected' : 'disconnected'">
                {{ messageStore.connected ? '已连接' : '未连接' }}
              </span>
            </el-text>
          </div>
        </div>
      </el-card>
    </el-main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useMessageStore } from '../stores/message'

const router = useRouter()
const authStore = useAuthStore()
const messageStore = useMessageStore()

const textFormRef = ref()
const imageFormRef = ref()
const uploadRef = ref()

const textForm = reactive({
  content: ''
})

const selectedImage = ref<File | null>(null)

// 处理文本消息发送
const handleSendText = async () => {
  if (!textFormRef.value) return
  
  if (!textForm.content.trim()) {
    return
  }
  
  const success = await messageStore.sendTextMessage(textForm.content)
  if (success) {
    textForm.content = ''
  }
}

// 处理图片选择
const handleImageChange = (file: any) => {
  selectedImage.value = file.raw
}

// 处理图片消息发送
const handleSendImage = async () => {
  if (!selectedImage.value) return
  
  const success = await messageStore.sendImageMessage(selectedImage.value)
  if (success) {
    // 重置上传组件
    if (uploadRef.value) {
      uploadRef.value.clearFiles()
    }
    selectedImage.value = null
  }
}

// 处理退出登录
const handleLogout = () => {
  authStore.logout()
  messageStore.disconnectWebSocket()
  router.push('/')
}

// 组件挂载时连接WebSocket
onMounted(() => {
  messageStore.connectWebSocket()
})

// 组件卸载时断开WebSocket
onBeforeUnmount(() => {
  messageStore.disconnectWebSocket()
})
</script>

<style scoped>
.home-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.home-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #409eff;
  color: white;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.header-left h1 {
  margin: 0;
  font-size: 24px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.status-badge {
  margin-right: 20px;
}

.logout-button {
  margin-left: 20px;
}

.home-main {
  flex: 1;
  padding: 20px;
  background-color: #f5f7fa;
  overflow-y: auto;
}

.send-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: center;
  align-items: center;
}

.send-content {
  padding: 20px;
}

.divider {
  margin: 30px 0;
  border: 1px solid #e4e7ed;
}

.send-button {
  margin-right: 10px;
}

.ws-status {
  margin-top: 20px;
  text-align: center;
}

.connected {
  color: #67c23a;
  font-weight: bold;
}

.disconnected {
  color: #f56c6c;
  font-weight: bold;
}
</style>
