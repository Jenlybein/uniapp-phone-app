import { defineStore } from 'pinia'
import axios from 'axios'
import { useAuthStore } from './auth'

interface Message {
  id: number
  user_id: number
  type: 'text' | 'image'
  content: string
  sender: 'pc' | 'server'
  is_selected: boolean
  created_at: string
}

interface MessageState {
  messages: Message[]
  loading: boolean
  error: string | null
  ws: WebSocket | null
  connected: boolean
}

export const useMessageStore = defineStore('message', {
  state: (): MessageState => ({
    messages: [],
    loading: false,
    error: null,
    ws: null,
    connected: false
  }),

  actions: {
    async sendTextMessage(content: string) {
      this.loading = true
      this.error = null
      
      const authStore = useAuthStore()
      
      try {
        await axios.post('http://localhost:8080/api/message', {
          content
        }, {
          headers: {
            Authorization: authStore.token as string
          }
        })
        
        return true
      } catch (error: any) {
        this.error = error.response?.data?.message || '发送失败，请稍后重试'
        return false
      } finally {
        this.loading = false
      }
    },

    async sendImageMessage(image: File) {
      this.loading = true
      this.error = null
      
      const authStore = useAuthStore()
      
      try {
        const formData = new FormData()
        formData.append('image', image)
        
        await axios.post('http://localhost:8080/api/image', formData, {
          headers: {
            Authorization: authStore.token as string,
            'Content-Type': 'multipart/form-data'
          }
        })
        
        return true
      } catch (error: any) {
        this.error = error.response?.data?.message || '发送失败，请稍后重试'
        return false
      } finally {
        this.loading = false
      }
    },

    connectWebSocket() {
      const authStore = useAuthStore()
      
      if (!authStore.token) {
        this.error = '请先登录'
        return
      }
      
      // 创建WebSocket连接
      const ws = new WebSocket(`ws://localhost:8080/ws?token=${authStore.token}`)
      
      ws.onopen = () => {
        console.log('WebSocket连接已建立')
        this.connected = true
        this.ws = ws
      }
      
      ws.onmessage = (event) => {
        console.log('收到消息:', event.data)
        // 处理收到的消息
        try {
          const message = JSON.parse(event.data)
          this.messages.push(message)
        } catch (error) {
          console.error('解析消息失败:', error)
        }
      }
      
      ws.onclose = () => {
        console.log('WebSocket连接已关闭')
        this.connected = false
        this.ws = null
        // 尝试重连
        setTimeout(() => {
          this.connectWebSocket()
        }, 3000)
      }
      
      ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        this.error = 'WebSocket连接错误'
      }
    },

    disconnectWebSocket() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
        this.connected = false
      }
    }
  }
})
