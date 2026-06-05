<template>
  <div class="auth-page">
    <div class="auth-panel">
      <div class="auth-brand">
        <span class="auth-logo">S</span>
        <h1>SamaraAI</h1>
        <p>登录你的账号</p>
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" label-position="top" class="auth-form">
        <el-form-item label="账号" prop="username">
          <el-input v-model="loginForm.username" placeholder="用户名或注册邮箱" size="large" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
            size="large"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-button type="primary" :loading="loading" class="auth-submit" size="large" @click="handleLogin">
          登录
        </el-button>
        <p class="auth-switch">
          还没有账号？
          <router-link to="/register">立即注册</router-link>
        </p>
      </el-form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'LoginView',
  setup() {
    const router = useRouter()
    const loginFormRef = ref()
    const loading = ref(false)
    const loginForm = ref({ username: '', password: '' })

    const loginRules = {
      username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码至少 6 位', trigger: 'blur' }
      ]
    }

    const handleLogin = async () => {
      try {
        await loginFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/login', loginForm.value)
        if (response.data.status_code === 1000) {
          localStorage.setItem('token', response.data.token)
          ElMessage.success('登录成功')
          router.push('/menu')
        } else {
          ElMessage.error(response.data.status_msg || '登录失败')
        }
      } catch (error) {
        if (error !== false) ElMessage.error('登录失败')
      } finally {
        loading.value = false
      }
    }

    return { loginFormRef, loading, loginForm, loginRules, handleLogin }
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: var(--ds-bg-muted);
}

.auth-panel {
  width: 100%;
  max-width: 400px;
  padding: 40px 36px;
  background: var(--ds-bg);
  border: 1px solid var(--ds-border);
  border-radius: 16px;
  box-shadow: var(--ds-shadow);
}

.auth-brand {
  text-align: center;
  margin-bottom: 32px;
}

.auth-logo {
  display: inline-flex;
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: var(--ds-primary);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.auth-brand h1 {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 6px;
}

.auth-brand p {
  font-size: 14px;
  color: var(--ds-text-secondary);
}

.auth-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--ds-text);
}

.auth-submit {
  width: 100%;
  margin-top: 8px;
  height: 44px;
  border-radius: 10px;
  background: var(--ds-primary) !important;
  border-color: var(--ds-primary) !important;
}

.auth-switch {
  margin-top: 20px;
  text-align: center;
  font-size: 14px;
  color: var(--ds-text-secondary);
}

.auth-switch a {
  color: var(--ds-primary);
  text-decoration: none;
  font-weight: 500;
}

.auth-switch a:hover {
  text-decoration: underline;
}
</style>
