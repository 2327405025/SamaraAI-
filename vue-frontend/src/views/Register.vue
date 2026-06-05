<template>
  <div class="auth-page">
    <div class="auth-panel auth-panel-wide">
      <div class="auth-brand">
        <span class="auth-logo">S</span>
        <h1>创建账号</h1>
        <p>注册后验证码将发送至邮箱</p>
      </div>

      <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" label-position="top" class="auth-form">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="registerForm.email" type="email" placeholder="your@email.com" size="large" />
        </el-form-item>
        <el-form-item label="验证码" prop="captcha">
          <div class="captcha-row">
            <el-input v-model="registerForm.captcha" placeholder="6 位验证码" size="large" />
            <el-button
              type="primary"
              plain
              :loading="codeLoading"
              :disabled="countdown > 0"
              size="large"
              @click="sendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="registerForm.password" type="password" show-password placeholder="至少 6 位" size="large" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="registerForm.confirmPassword" type="password" show-password placeholder="再次输入" size="large" />
        </el-form-item>
        <el-button type="primary" :loading="loading" class="auth-submit" size="large" @click="handleRegister">
          注册
        </el-button>
        <p class="auth-switch">
          已有账号？
          <router-link to="/login">去登录</router-link>
        </p>
      </el-form>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'RegisterView',
  setup() {
    const router = useRouter()
    const registerFormRef = ref()
    const loading = ref(false)
    const codeLoading = ref(false)
    const countdown = ref(0)

    const registerForm = reactive({
      email: '',
      captcha: '',
      password: '',
      confirmPassword: ''
    })

    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== registerForm.password) callback(new Error('两次密码不一致'))
      else callback()
    }

    const registerRules = {
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
      ],
      captcha: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '至少 6 位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请确认密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ]
    }

    const sendCode = async () => {
      if (!registerForm.email) {
        ElMessage.warning('请先填写邮箱')
        return
      }
      try {
        codeLoading.value = true
        const response = await api.post('/user/captcha', { email: registerForm.email })
        if (response.data.status_code === 1000) {
          ElMessage.success('验证码已发送')
          countdown.value = 60
          const timer = setInterval(() => {
            countdown.value--
            if (countdown.value <= 0) clearInterval(timer)
          }, 1000)
        } else {
          ElMessage.error(response.data.status_msg || '发送失败')
        }
      } catch {
        ElMessage.error('发送失败')
      } finally {
        codeLoading.value = false
      }
    }

    const handleRegister = async () => {
      try {
        await registerFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/register', {
          email: registerForm.email,
          captcha: registerForm.captcha,
          password: registerForm.password
        })
        if (response.data.status_code === 1000) {
          ElMessage.success('注册成功，请登录')
          router.push('/login')
        } else {
          ElMessage.error(response.data.status_msg || '注册失败')
        }
      } catch (error) {
        if (error !== false) ElMessage.error('注册失败')
      } finally {
        loading.value = false
      }
    }

    return {
      registerFormRef,
      loading,
      codeLoading,
      countdown,
      registerForm,
      registerRules,
      sendCode,
      handleRegister
    }
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

.auth-panel-wide {
  max-width: 440px;
}

.auth-brand {
  text-align: center;
  margin-bottom: 28px;
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

.captcha-row {
  display: flex;
  gap: 10px;
  width: 100%;
}

.captcha-row .el-input {
  flex: 1;
}

.auth-form :deep(.el-form-item__label) {
  font-weight: 500;
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
</style>
