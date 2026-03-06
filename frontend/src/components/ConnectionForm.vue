<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditMode ? '编辑连接' : '新建连接'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      label-position="right"
    >
      <!-- 基本信息 -->
      <el-divider content-position="left">基本信息</el-divider>
      
      <el-form-item label="连接名称" prop="name">
        <el-input
          v-model="formData.name"
          placeholder="请输入连接名称"
          clearable
        />
      </el-form-item>

      <el-form-item label="主机地址" prop="host">
        <el-input
          v-model="formData.host"
          placeholder="例如: localhost 或 192.168.1.100"
          clearable
        />
      </el-form-item>

      <el-form-item label="端口" prop="port">
        <el-input-number
          v-model="formData.port"
          :min="1"
          :max="65535"
          :step="1"
          controls-position="right"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="用户名" prop="username">
        <el-input
          v-model="formData.username"
          placeholder="请输入数据库用户名"
          clearable
        />
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input
          v-model="formData.password"
          type="password"
          placeholder="请输入数据库密码"
          show-password
          clearable
        />
      </el-form-item>

      <el-form-item label="数据库" prop="database">
        <el-input
          v-model="formData.database"
          placeholder="默认数据库（可选）"
          clearable
        />
      </el-form-item>

      <el-form-item label="字符集" prop="charset">
        <el-select
          v-model="formData.charset"
          placeholder="请选择字符集"
          style="width: 100%"
        >
          <el-option label="utf8mb4" value="utf8mb4" />
          <el-option label="utf8" value="utf8" />
          <el-option label="latin1" value="latin1" />
          <el-option label="gbk" value="gbk" />
        </el-select>
      </el-form-item>

      <el-form-item label="超时时间" prop="timeout">
        <el-input-number
          v-model="formData.timeout"
          :min="1"
          :max="60"
          :step="1"
          controls-position="right"
          style="width: 100%"
        />
        <span style="margin-left: 10px; color: #909399; font-size: 12px">秒</span>
      </el-form-item>

      <!-- SSH 隧道配置 -->
      <el-divider content-position="left">
        <el-checkbox v-model="formData.sshEnabled" @change="handleSshToggle">
          SSH 隧道（可选）
        </el-checkbox>
      </el-divider>

      <template v-if="formData.sshEnabled">
        <el-form-item label="SSH 主机" prop="sshHost">
          <el-input
            v-model="formData.sshHost"
            placeholder="SSH 服务器地址"
            clearable
          />
        </el-form-item>

        <el-form-item label="SSH 端口" prop="sshPort">
          <el-input-number
            v-model="formData.sshPort"
            :min="1"
            :max="65535"
            :step="1"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="SSH 用户名" prop="sshUsername">
          <el-input
            v-model="formData.sshUsername"
            placeholder="SSH 用户名"
            clearable
          />
        </el-form-item>

        <el-form-item label="SSH 密码" prop="sshPassword">
          <el-input
            v-model="formData.sshPassword"
            type="password"
            placeholder="SSH 密码（如使用密钥则留空）"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item label="SSH 私钥路径" prop="sshKeyPath">
          <el-input
            v-model="formData.sshKeyPath"
            placeholder="私钥文件路径（如使用密码则留空）"
            clearable
          />
        </el-form-item>
      </template>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="info"
          :loading="testing"
          @click="handleTestConnection"
        >
          测试连接
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSave"
        >
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { ConnectionAPI } from '../api';
import type { ConnectionProfile } from '../types/api';

interface Props {
  visible: boolean;
  profile?: ConnectionProfile | null;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'success'): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  profile: null,
});

const emit = defineEmits<Emits>();

const router = useRouter();
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();
const testing = ref(false);
const saving = ref(false);

const isEditMode = ref(false);

// 表单数据
const formData = reactive<Partial<ConnectionProfile>>({
  id: '',
  name: '',
  host: 'localhost',
  port: 3306,
  username: '',
  password: '',
  database: '',
  charset: 'utf8mb4',
  timeout: 10,
  sshEnabled: false,
  sshHost: '',
  sshPort: 22,
  sshUsername: '',
  sshPassword: '',
  sshKeyPath: '',
  // Don't initialize timestamp fields - they should be undefined for new profiles
});

// 表单验证规则
const rules = reactive<FormRules>({
  name: [
    { required: true, message: '请输入连接名称', trigger: 'blur' },
    { min: 1, max: 50, message: '长度在 1 到 50 个字符', trigger: 'blur' },
  ],
  host: [
    { required: true, message: '请输入主机地址', trigger: 'blur' },
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围 1-65535', trigger: 'blur' },
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
  ],
  charset: [
    { required: true, message: '请选择字符集', trigger: 'change' },
  ],
  timeout: [
    { required: true, message: '请输入超时时间', trigger: 'blur' },
    { type: 'number', min: 1, max: 60, message: '超时时间范围 1-60 秒', trigger: 'blur' },
  ],
  sshHost: [
    {
      validator: (rule, value, callback) => {
        if (formData.sshEnabled && !value) {
          callback(new Error('启用 SSH 隧道时，SSH 主机地址为必填项'));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
  sshPort: [
    {
      validator: (rule, value, callback) => {
        if (formData.sshEnabled) {
          if (!value) {
            callback(new Error('启用 SSH 隧道时，SSH 端口为必填项'));
          } else if (value < 1 || value > 65535) {
            callback(new Error('SSH 端口号范围 1-65535'));
          } else {
            callback();
          }
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
  sshUsername: [
    {
      validator: (rule, value, callback) => {
        if (formData.sshEnabled && !value) {
          callback(new Error('启用 SSH 隧道时，SSH 用户名为必填项'));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
  sshPassword: [
    {
      validator: (rule, value, callback) => {
        if (formData.sshEnabled && !value && !formData.sshKeyPath) {
          callback(new Error('SSH 密码和私钥路径至少填写一项'));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
  sshKeyPath: [
    {
      validator: (rule, value, callback) => {
        if (formData.sshEnabled && !value && !formData.sshPassword) {
          callback(new Error('SSH 密码和私钥路径至少填写一项'));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
});

// 监听 visible 变化
watch(
  () => props.visible,
  (newVal) => {
    dialogVisible.value = newVal;
    if (newVal) {
      initForm();
    }
  },
  { immediate: true }
);

// 监听 dialogVisible 变化
watch(dialogVisible, (newVal) => {
  if (!newVal) {
    emit('update:visible', false);
  }
});

// 初始化表单
const initForm = () => {
  isEditMode.value = !!props.profile;
  
  if (props.profile) {
    // 编辑模式：填充现有数据
    Object.assign(formData, props.profile);
  } else {
    // 创建模式：重置为默认值
    resetForm();
  }
  
  // 重置验证状态
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

// 重置表单
const resetForm = () => {
  formData.id = '';
  formData.name = '';
  formData.host = 'localhost';
  formData.port = 3306;
  formData.username = '';
  formData.password = '';
  formData.database = '';
  formData.charset = 'utf8mb4';
  formData.timeout = 10;
  formData.sshEnabled = false;
  formData.sshHost = '';
  formData.sshPort = 22;
  formData.sshUsername = '';
  formData.sshPassword = '';
  formData.sshKeyPath = '';
  // Don't set timestamp fields - they should be undefined
  delete formData.createdAt;
  delete formData.updatedAt;
};

// SSH 隧道开关切换
const handleSshToggle = (enabled: boolean) => {
  if (!enabled) {
    // 清空 SSH 相关字段
    formData.sshHost = '';
    formData.sshPort = 22;
    formData.sshUsername = '';
    formData.sshPassword = '';
    formData.sshKeyPath = '';
  }
  // 触发验证
  nextTick(() => {
    formRef.value?.clearValidate([
      'sshHost',
      'sshPort',
      'sshUsername',
      'sshPassword',
      'sshKeyPath',
    ]);
  });
};

// 测试连接
const handleTestConnection = async () => {
  if (!formRef.value) return;

  try {
    // 验证表单
    await formRef.value.validate();

    testing.value = true;

    // 调用后端测试连接 API
    await ConnectionAPI.testConnection(formData as ConnectionProfile);

    ElMessage.success({
      message: formData.sshEnabled
        ? '连接测试成功！SSH 隧道和数据库连接均正常'
        : '连接测试成功！',
      duration: 3000,
    });
  } catch (error: any) {
    if (error !== false) {
      // 区分 SSH 连接失败和数据库连接失败
      const errorMsg = error.message || error;
      if (errorMsg.includes('SSH') || errorMsg.includes('ssh')) {
        ElMessage.error(`SSH 连接失败: ${errorMsg}`);
      } else if (errorMsg.includes('database') || errorMsg.includes('MySQL')) {
        ElMessage.error(`数据库连接失败: ${errorMsg}`);
      } else {
        ElMessage.error(`连接测试失败: ${errorMsg}`);
      }
    }
  } finally {
    testing.value = false;
  }
};

// 保存连接配置
const handleSave = async () => {
  if (!formRef.value) return;

  try {
    // 验证表单
    await formRef.value.validate();

    saving.value = true;

    if (isEditMode.value) {
      // 更新现有连接
      await ConnectionAPI.updateProfile(formData.id!, formData as ConnectionProfile);
      ElMessage.success('连接配置已更新');
    } else {
      // 创建新连接
      // 生成新的 ID（实际应该由后端生成）
      formData.id = `conn_${Date.now()}`;
      // Don't set timestamps - backend will handle them
      delete formData.createdAt;
      delete formData.updatedAt;
      
      await ConnectionAPI.createProfile(formData as ConnectionProfile);
      ElMessage.success('连接配置已创建');
    }

    // 通知父组件刷新列表
    emit('success');
    
    // 关闭对话框
    handleClose();
  } catch (error: any) {
    if (error !== false) {
      ElMessage.error(`保存失败: ${error.message || error}`);
    }
  } finally {
    saving.value = false;
  }
};

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false;
  // 延迟重置表单，避免关闭动画时看到表单变化
  setTimeout(() => {
    resetForm();
    formRef.value?.clearValidate();
  }, 300);
};
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-divider__text) {
  font-weight: 600;
  color: #303133;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}
</style>
