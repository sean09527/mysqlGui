package backend

import (
	"context"
	"fmt"
	"sync"

	"mygui/backend/internal/connection"
	"mygui/backend/internal/data"
	"mygui/backend/internal/importexport"
	"mygui/backend/internal/logger"
	"mygui/backend/internal/query"
	"mygui/backend/internal/schema"
	"mygui/backend/internal/security"
	"mygui/backend/internal/storage"
	syncengine "mygui/backend/internal/sync"
	"mygui/backend/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 是 Wails 应用的主结构体，包含所有服务组件
type App struct {
	ctx               context.Context
	connectionManager *connection.ConnectionManager
	configStorage     *storage.ConfigStorage
	encryptor         *security.Encryptor
	logger            *logger.Logger
	
	// 当前活动连接的管理器实例（按 profileID 索引）
	schemaManagers    map[string]*schema.Manager
	dataManagers      map[string]*data.Manager
	queryExecutors    map[string]*query.Executor
	syncEngines       map[string]*syncengine.SyncEngine
	importExportServices map[string]*importexport.Service
	
	mu sync.RWMutex
}

// NewApp 创建新的 App 实例
func NewApp() *App {
	return &App{
		schemaManagers:       make(map[string]*schema.Manager),
		dataManagers:         make(map[string]*data.Manager),
		queryExecutors:       make(map[string]*query.Executor),
		syncEngines:          make(map[string]*syncengine.SyncEngine),
		importExportServices: make(map[string]*importexport.Service),
	}
}

// Greet 测试方法 - 用于验证 Wails 绑定是否工作
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

// Startup 在应用启动时调用，初始化所有服务组件
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	
	// 初始化加密器（使用固定的密钥，实际应用中应该从安全的地方获取）
	// TODO: 在生产环境中，应该使用更安全的密钥管理方式
	encryptor := security.NewEncryptor("mysql-management-tool-secret-key-2024")
	a.encryptor = encryptor
	
	// 初始化配置存储
	configStorage, err := storage.NewConfigStorage()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config storage: %v", err))
	}
	a.configStorage = configStorage
	
	// 初始化日志服务
	logger, err := logger.NewLogger(types.LogLevelInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	a.logger = logger
	
	// 初始化连接管理器
	a.connectionManager = connection.NewConnectionManager(configStorage, encryptor)
	
	// 记录启动日志
	a.logger.Info("Application started", map[string]interface{}{
		"version": "1.0.0",
	})
}

// Shutdown 在应用关闭时调用，清理所有资源
func (a *App) Shutdown(ctx context.Context) {
	// 断开所有连接
	if a.connectionManager != nil {
		if err := a.connectionManager.DisconnectAll(); err != nil {
			a.logger.Error("Failed to disconnect all connections", err, nil)
		}
	}
	
	// 关闭日志服务
	if a.logger != nil {
		a.logger.Info("Application shutting down", nil)
		if err := a.logger.Close(); err != nil {
			fmt.Printf("failed to close logger: %v\n", err)
		}
	}
	
	// 关闭配置存储
	if a.configStorage != nil {
		if err := a.configStorage.Close(); err != nil {
			fmt.Printf("failed to close config storage: %v\n", err)
		}
	}
}

// getOrCreateSchemaManager 获取或创建指定连接的 SchemaManager
func (a *App) getOrCreateSchemaManager(profileID string) (*schema.Manager, error) {
	a.mu.RLock()
	manager, exists := a.schemaManagers[profileID]
	a.mu.RUnlock()
	
	if exists {
		return manager, nil
	}
	
	// 获取数据库连接
	db, err := a.connectionManager.GetConnection(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	
	// 创建新的 SchemaManager
	manager = schema.NewManager(db)
	
	a.mu.Lock()
	a.schemaManagers[profileID] = manager
	a.mu.Unlock()
	
	return manager, nil
}

// getOrCreateDataManager 获取或创建指定连接的 DataManager
func (a *App) getOrCreateDataManager(profileID string) (*data.Manager, error) {
	a.mu.RLock()
	manager, exists := a.dataManagers[profileID]
	a.mu.RUnlock()
	
	if exists {
		return manager, nil
	}
	
	// 获取数据库连接
	db, err := a.connectionManager.GetConnection(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	
	// 创建新的 DataManager
	manager = data.NewManager(db)
	
	a.mu.Lock()
	a.dataManagers[profileID] = manager
	a.mu.Unlock()
	
	return manager, nil
}

// getOrCreateQueryExecutor 获取或创建指定连接的 QueryExecutor
func (a *App) getOrCreateQueryExecutor(profileID string) (*query.Executor, error) {
	a.mu.RLock()
	executor, exists := a.queryExecutors[profileID]
	a.mu.RUnlock()
	
	if exists {
		return executor, nil
	}
	
	// 获取数据库连接
	db, err := a.connectionManager.GetConnection(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	
	// 创建新的 QueryExecutor
	executor = query.NewExecutor(db)
	
	a.mu.Lock()
	a.queryExecutors[profileID] = executor
	a.mu.Unlock()
	
	return executor, nil
}

// getOrCreateSyncEngine 获取或创建指定连接的 SyncEngine
func (a *App) getOrCreateSyncEngine(sourceProfileID, targetProfileID string) (*syncengine.SyncEngine, error) {
	key := fmt.Sprintf("%s:%s", sourceProfileID, targetProfileID)
	
	a.mu.RLock()
	engine, exists := a.syncEngines[key]
	a.mu.RUnlock()
	
	if exists {
		return engine, nil
	}
	
	// 获取源和目标数据库连接
	sourceDB, err := a.connectionManager.GetConnection(sourceProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source connection: %w", err)
	}
	
	targetDB, err := a.connectionManager.GetConnection(targetProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target connection: %w", err)
	}
	
	// 创建新的 SyncEngine
	engine = syncengine.NewSyncEngine(sourceDB, targetDB)
	
	a.mu.Lock()
	a.syncEngines[key] = engine
	a.mu.Unlock()
	
	return engine, nil
}

// getOrCreateImportExportService 获取或创建指定连接的 ImportExportService
func (a *App) getOrCreateImportExportService(profileID string) (*importexport.Service, error) {
	a.mu.RLock()
	service, exists := a.importExportServices[profileID]
	a.mu.RUnlock()
	
	if exists {
		return service, nil
	}
	
	// 获取数据库连接
	db, err := a.connectionManager.GetConnection(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	
	// 创建新的 ImportExportService
	service = importexport.NewService(db)
	
	a.mu.Lock()
	a.importExportServices[profileID] = service
	a.mu.Unlock()
	
	return service, nil
}

// cleanupManagersForProfile 清理指定连接的所有管理器实例
func (a *App) cleanupManagersForProfile(profileID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	delete(a.schemaManagers, profileID)
	delete(a.dataManagers, profileID)
	delete(a.queryExecutors, profileID)
	delete(a.importExportServices, profileID)
	
	// 清理与此 profile 相关的 SyncEngine
	for key := range a.syncEngines {
		if contains(key, profileID) {
			delete(a.syncEngines, key)
		}
	}
}

// emitEvent 向前端发送事件
func (a *App) emitEvent(eventName string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, eventName, data)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)
}
