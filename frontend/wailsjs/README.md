# Wails 绑定目录

此目录包含 Wails 自动生成的前后端绑定代码。

## 重要说明

⚠️ **不要手动编辑此目录中的文件！**

当你运行 `wails dev` 或 `wails build` 时，Wails 会自动：
1. 扫描后端的 Go 代码
2. 生成对应的 JavaScript/TypeScript 绑定
3. 覆盖此目录中的所有文件

## 目录结构

```
wailsjs/
├── go/
│   └── backend/
│       ├── App.js          # 后端方法的 JavaScript 绑定（自动生成）
│       └── App.d.ts        # TypeScript 类型定义（自动生成）
└── runtime/
    ├── runtime.js          # Wails 运行时（自动生成）
    └── runtime.d.ts        # 运行时类型定义（自动生成）
```

## 当前状态

当前目录包含临时占位文件，用于避免开发时的 TypeScript 错误。

运行 `wails dev` 后，这些文件会被 Wails 生成的真实绑定文件替换。

## 如何生成绑定

```bash
# 开发模式（推荐）
wails dev

# 或者仅生成绑定
wails generate module
```

## 故障排除

如果绑定文件没有正确生成：

1. 删除此目录：
   ```bash
   rm -rf frontend/wailsjs/go
   ```

2. 重新运行 Wails：
   ```bash
   wails dev
   ```

3. 检查后端代码是否有语法错误：
   ```bash
   cd backend
   go build
   ```
