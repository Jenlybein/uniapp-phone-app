<template>
  <view class="register-container">
    <view class="register-form">
      <view class="form-title">
        <text class="title-text">注册</text>
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
          <text class="label">邮箱</text>
          <input 
            v-model="email" 
            class="input" 
            type="email" 
            placeholder="请输入邮箱" 
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
        
        <view class="input-group">
          <text class="label">确认密码</text>
          <input 
            v-model="confirmPassword" 
            class="input" 
            type="password" 
            placeholder="请确认密码" 
            placeholder-style="color: #999"
          />
        </view>
        
        <view v-if="error" class="error-message">
          {{ error }}
        </view>
        
        <view class="button-group">
          <button 
            type="primary" 
            class="register-button" 
            @click="handleRegister" 
            :loading="loading"
          >
            注册
          </button>
          <button 
            type="default" 
            class="login-button" 
            @click="navigateToLogin"
          >
            登录
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')

const handleRegister = async () => {
  // 表单验证
  if (!username.value.trim()) {
    error.value = '请输入用户名'
    return
  }
  
  if (!email.value.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value)) {
    error.value = '请输入有效的邮箱地址'
    return
  }
  
  if (!password.value.trim() || password.value.length < 6) {
    error.value = '密码长度不能少于6个字符'
    return
  }
  
  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    const response = await uni.request({
      url: 'http://localhost:8080/api/auth/register',
      method: 'POST',
      data: {
        username: username.value,
        email: email.value,
        password: password.value
      },
      header: {
        'content-type': 'application/json'
      }
    })
    
    if (response.statusCode === 200 && response.data.code === 200) {
      // 注册成功，保存token
      uni.setStorageSync('token', response.data.data.token)
      
      // 跳转到首页
      uni.redirectTo({
        url: '/pages/index/index'
      })
    } else {
      error.value = response.data.message || '注册失败，请稍后重试'
    }
  } catch (err) {
    error.value = '网络错误，请稍后重试'
    console.error('注册失败:', err)
  } finally {
    loading.value = false
  }
}

const navigateToLogin = () => {
  uni.navigateTo({
    url: '/pages/login/login'
  })
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;
}

.register-form {
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

.register-button {
  width: 100%;
  height: 45px;
  background-color: #409eff;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
}

.login-button {
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
