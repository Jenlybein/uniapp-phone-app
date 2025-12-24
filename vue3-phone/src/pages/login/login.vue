<template>
  <view class="login-container">
    <view class="login-form">
      <view class="form-title">
        <text class="title-text">登录</text>
      </view>
      
      <view class="form-content">
        <view class="input-group">
          <text class="label">用户名</text>
          <input 
            v-model="username" 
            class="input" 
            placeholder="请输入用户名" 
            placeholder-style="color: #999"
          />
        </view>
        
        <view class="input-group">
          <text class="label">密码</text>
          <input 
            v-model="password" 
            class="input" 
            type="password" 
            placeholder="请输入密码" 
            placeholder-style="color: #999"
          />
        </view>
        
        <view v-if="error" class="error-message">
          {{ error }}
        </view>
        
        <view class="button-group">
          <button 
            type="primary" 
            class="login-button" 
            @click="handleLogin" 
            :loading="loading"
          >
            登录
          </button>
          <button 
            type="default" 
            class="register-button" 
            @click="navigateToRegister"
          >
            注册
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  if (!username.value.trim() || !password.value.trim()) {
    error.value = '请输入用户名和密码'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    const response = await uni.request({
      url: 'http://localhost:8080/api/auth/login',
      method: 'POST',
      data: {
        username: username.value,
        password: password.value
      },
      header: {
        'content-type': 'application/json'
      }
    })
    
    if (response.statusCode === 200 && response.data.code === 200) {
      // 登录成功，保存token和用户信息
      uni.setStorageSync('token', response.data.data.token)
      uni.setStorageSync('user', JSON.stringify(response.data.data.user))
      
      // 跳转到首页
      uni.redirectTo({
        url: '/pages/index/index'
      })
    } else {
      error.value = response.data.message || '登录失败，请检查用户名和密码'
    }
  } catch (err) {
    error.value = '网络错误，请稍后重试'
    console.error('登录失败:', err)
  } finally {
    loading.value = false
  }
}

const navigateToRegister = () => {
  uni.navigateTo({
    url: '/pages/register/register'
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;
}

.login-form {
  width: 80%;
  max-width: 400px;
  background-color: #fff;
  border-radius: 10px;
  padding: 30px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.form-title {
  text-align: center;
  margin-bottom: 30px;
}

.title-text {
  font-size: 24px;
  font-weight: bold;
  color: #333;
}

.form-content {
  width: 100%;
}

.input-group {
  margin-bottom: 20px;
}

.label {
  display: block;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

.input {
  width: 100%;
  height: 45px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 0 15px;
  font-size: 14px;
  color: #333;
  background-color: #fff;
  box-sizing: border-box;
}

.input:focus {
  outline: none;
  border-color: #409eff;
}

.error-message {
  color: #f56c6c;
  font-size: 14px;
  margin-bottom: 20px;
  text-align: center;
}

.button-group {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.login-button {
  width: 100%;
  height: 45px;
  background-color: #409eff;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
}

.register-button {
  width: 100%;
  height: 45px;
  background-color: #fff;
  color: #409eff;
  border: 1px solid #409eff;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
}
</style>
