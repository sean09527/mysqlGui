# 组件文档

## ConnectionForm.vue

连接配置表单组件，用于创建和编辑 MySQL 数据库连接配置。

### 功能特性

- **基本连接配置**：主机地址、端口、用户名、密码、数据库、字符集、超时时间
- **SSH 隧道支持**：可选的 SSH 隧道配置，支持密码和私钥两种认证方式
- **表单验证**：完整的表单验证规则，确保数据有效性
- **测试连接**：在保存前测试连接是否有效
- **创建/编辑模式**：支持创建新连接和编辑现有连接

### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| visible | boolean | false | 控制对话框显示/隐藏 |
| profile | ConnectionProfile \| null | null | 编辑模式时传入的连接配置对象 |

### Events

| 事件名 | 参数 | 说明 |
|--------|------|------|
| update:visible | value: boolean | 对话框显示状态变化时触发 |
| success | - | 保存成功后触发 |

### 使用示例

```vue
<template>
  <div>
    <el-button @click="showForm">新建连接</el-button>
    
    <ConnectionForm
      v-model:visible="formVisible"
      :profile="currentProfile"
      @success="handleSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import ConnectionForm from './ConnectionForm.vue';
import type { ConnectionProfile } from '../types/api';

const formVisible = ref(false);
const currentProfile = ref<ConnectionProfile | null>(null);

const showForm = () => {
  currentProfile.value = null; // 创建模式
  formVisible.value = true;
};

const handleSuccess = () => {
  console.log('连接配置已保存');
  // 刷新连接列表等操作
};
</script>
```

### 表单验证规则

#### 基本信息验证

- **连接名称**：必填，长度 1-50 个字符
- **主机地址**：必填
- **端口**：必填，范围 1-65535
- **用户名**：必填
- **密码**：必填
- **字符集**：必填，可选值：utf8mb4, utf8, latin1, gbk
- **超时时间**：必填，范围 1-60 秒

#### SSH 隧道验证（启用时）

- **SSH 主机**：必填
- **SSH 端口**：必填，范围 1-65535
- **SSH 用户名**：必填
- **SSH 密码/私钥路径**：至少填写一项

### 测试连接功能

点击"测试连接"按钮会：

1. 验证表单数据
2. 调用后端 `ConnectionAPI.testConnection()` 方法
3. 显示测试结果：
   - 成功：显示成功消息
   - 失败：区分 SSH 连接失败和数据库连接失败，显示具体错误信息

### 保存功能

点击"保存"按钮会：

1. 验证表单数据
2. 根据模式调用不同的 API：
   - 创建模式：调用 `ConnectionAPI.createProfile()`
   - 编辑模式：调用 `ConnectionAPI.updateProfile()`
3. 保存成功后触发 `success` 事件并关闭对话框

### 需求映射

该组件实现了以下需求：

- **需求 1.1**：允许用户创建新的连接配置
- **需求 1.2**：存储连接配置的主机地址、端口、用户名、密码和数据库名称
- **需求 1.3**：支持可选的 SSH 隧道配置
- **需求 2.1**：测试连接功能
- **需求 2.2**：SSH 隧道连接测试
- **需求 2.3**：显示连接成功消息
- **需求 2.4**：显示连接失败原因，区分 SSH 和数据库连接错误

## ConnectionManager.vue

连接管理器组件，显示所有连接配置列表并提供管理功能。

### 功能特性

- 显示所有连接配置列表
- 连接/断开数据库
- 创建新连接
- 编辑现有连接
- 删除连接配置
- 显示连接状态和 SSH 隧道状态

### 集成说明

ConnectionManager 组件已集成 ConnectionForm 组件：

- 点击"新建连接"按钮打开创建表单
- 点击"编辑"按钮打开编辑表单
- 表单保存成功后自动刷新连接列表

