<template>
  <div class="register-container">
    <div class="register-form">
      <div class="form-title">
        <h2>注册</h2>
      </div>
      
      <div class="form-content">
        <div class="input-group">
          <label class="label">用户名</label>
          <el-input v-model="registerForm.username" placeholder="请输入用户名" class="input" />
        </div>
        
        <div class="input-group">
          <label class="label">邮箱</label>
          <el-input v-model="registerForm.email" placeholder="请输入邮箱" class="input" />
        </div>
        
        <div class="input-group">
          <label class="label">密码</label>
          <el-input v-model="registerForm.password" type="password" placeholder="请输入密码" show-password class="input" />
        </div>
        
        <div class="input-group">
          <label class="label">确认密码</label>
          <el-input v-model="registerForm.confirmPassword" type="password" placeholder="请确认密码" show-password class="input" />
        </div>
        
        <div v-if="authStore.error" class="error-message">
          {{ authStore.error }}
        </div>
        
        <div class="button-group">
          <el-button type="primary" @click="handleRegister" :loading="authStore.loading" class="register-button">
            注册
          </el-button>
          <el-button @click="$router.push('/')" class="login-button">
            登录
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const handleRegister = async () => {
  // 表单验证
  if (!registerForm.username.trim()) {
    authStore.error = '请输入用户名'
    return
  }
  
  if (!registerForm.email.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(registerForm.email)) {
    authStore.error = '请输入有效的邮箱地址'
    return
  }
  
  if (!registerForm.password.trim() || registerForm.password.length < 6) {
    authStore.error = '密码长度不能少于 6 个字符'
    return
  }
  
  if (registerForm.password !== registerForm.confirmPassword) {
    authStore.error = '两次输入的密码不一致'
    return
  }
  
  // 调用注册方法
  const success = await authStore.register(registerForm.username, registerForm.email, registerForm.password)
  if (success) {
    router.push('/home')
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.register-form {
  width: 380px;
  background-color: #fff;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.3s ease;
}

.register-form:hover {
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.15);
}

.form-title {
  text-align: center;
  margin-bottom: 32px;
}

.form-title h2 {
  font-size: 28px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0;
}

.form-content {
  width: 100%;
}

.input-group {
  margin-bottom: 24px;
}

.label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
}

.input {
  width: 100%;
  height: 48px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 0 16px;
  font-size: 14px;
  color: #333;
  background-color: #fff;
  transition: all 0.3s ease;
}

.input:focus {
  outline: none;
  border-color: #007aff;
  box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
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
  gap: 16px;
}

.register-button {
  width: 100%;
  height: 48px;
  background-color: #007aff;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.register-button:hover {
  background-color: #0056b3;
  box-shadow: 0 4px 12px rgba(0, 122, 255, 0.3);
}

.register-button:active {
  background-color: #004085;
}

.login-button {
  width: 100%;
  height: 48px;
  background-color: #fff;
  color: #007aff;
  border: 1px solid #007aff;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.login-button:hover {
  background-color: rgba(0, 122, 255, 0.05);
  border-color: #0056b3;
}

.login-button:active {
  background-color: rgba(0, 122, 255, 0.1);
}
</style>
