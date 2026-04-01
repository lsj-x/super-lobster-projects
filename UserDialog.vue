<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑用户' : '新增用户'"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="userFormRef"
      :model="formData"
      :rules="rules"
      label-width="80px"
      status-icon
    >
      <el-form-item label="用户名" prop="username">
        <el-input v-model="formData.username" placeholder="请输入用户名" />
      </el-form-item>

      <el-form-item label="邮箱" prop="email">
        <el-input v-model="formData.email" placeholder="请输入邮箱" />
      </el-form-item>

      <el-form-item label="角色" prop="role">
        <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%">
          <el-option label="管理员" value="admin" />
          <el-option label="普通用户" value="user" />
          <el-option label="访客" value="guest" />
        </el-select>
      </el-form-item>

      <el-form-item label="状态" prop="status">
        <el-switch v-model="formData.status" active-text="启用" inactive-text="禁用" />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'

// 定义 Props
const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  userData: {
    type: Object,
    default: () => ({})
  },
  isEdit: {
    type: Boolean,
    default: false
  }
})

// 定义 Emits
const emit = defineEmits(['update:visible', 'submit'])

// 响应式状态
const dialogVisible = ref(props.visible)
const userFormRef = ref(null)

// 表单数据
const formData = reactive({
  username: '',
  email: '',
  role: '',
  status: true
})

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: ['blur', 'change'] }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 监听 visible 变化以同步状态
watch(
  () => props.visible,
  (val) => {
    dialogVisible.value = val
    if (val) {
      resetForm()
    }
  }
)

// 监听 userData 变化以填充表单（编辑模式）
watch(
  () => props.userData,
  (newData) => {
    if (props.isEdit && newData) {
      formData.username = newData.username || ''
      formData.email = newData.email || ''
      formData.role = newData.role || ''
      formData.status = newData.status !== undefined ? newData.status : true
    }
  },
  { immediate: true }
)

// 重置表单
const resetForm = () => {
  if (!userFormRef.value) return
  
  userFormRef.value.resetFields()
  // 如果是编辑模式，重新填充数据
  if (props.isEdit && props.userData) {
    formData.username = props.userData.username || ''
    formData.email = props.userData.email || ''
    formData.role = props.userData.role || ''
    formData.status = props.userData.status !== undefined ? props.userData.status : true
  } else {
    // 新增模式重置默认值
    formData.username = ''
    formData.email = ''
    formData
