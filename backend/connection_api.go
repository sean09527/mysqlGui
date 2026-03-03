package backend

import (
	"fmt"

	"mygui/backend/types"
)

// ConnectionAPI 提供连接管理相关的 API 方法，供前端调用

// CreateProfile 创建新的连接配置
func (a *App) CreateProfile(profile types.ConnectionProfile) error {
	err := a.connectionManager.CreateProfile(profile)
	if err != nil {
		a.logger.Error("Failed to create profile", err, map[string]interface{}{
			"profile_name": profile.Name,
		})
		return fmt.Errorf("创建连接配置失败: %w", err)
	}
	
	a.logger.Info("Profile created", map[string]interface{}{
		"profile_id":   profile.ID,
		"profile_name": profile.Name,
	})
	
	// 发送事件通知前端
	a.emitEvent("connection:profile:created", map[string]interface{}{
		"profileId": profile.ID,
		"name":      profile.Name,
	})
	
	return nil
}

// UpdateProfile 更新现有的连接配置
func (a *App) UpdateProfile(id string, profile types.ConnectionProfile) error {
	err := a.connectionManager.UpdateProfile(id, profile)
	if err != nil {
		a.logger.Error("Failed to update profile", err, map[string]interface{}{
			"profile_id":   id,
			"profile_name": profile.Name,
		})
		return fmt.Errorf("更新连接配置失败: %w", err)
	}
	
	a.logger.Info("Profile updated", map[string]interface{}{
		"profile_id":   id,
		"profile_name": profile.Name,
	})
	
	// 清理该连接的所有管理器实例
	a.cleanupManagersForProfile(id)
	
	// 发送事件通知前端
	a.emitEvent("connection:profile:updated", map[string]interface{}{
		"profileId": id,
		"name":      profile.Name,
	})
	
	return nil
}

// DeleteProfile 删除连接配置
func (a *App) DeleteProfile(id string) error {
	err := a.connectionManager.DeleteProfile(id)
	if err != nil {
		a.logger.Error("Failed to delete profile", err, map[string]interface{}{
			"profile_id": id,
		})
		return fmt.Errorf("删除连接配置失败: %w", err)
	}
	
	a.logger.Info("Profile deleted", map[string]interface{}{
		"profile_id": id,
	})
	
	// 清理该连接的所有管理器实例
	a.cleanupManagersForProfile(id)
	
	// 发送事件通知前端
	a.emitEvent("connection:profile:deleted", map[string]interface{}{
		"profileId": id,
	})
	
	return nil
}

// ListProfiles 获取所有连接配置列表
func (a *App) ListProfiles() ([]types.ConnectionProfile, error) {
	profiles, err := a.connectionManager.ListProfiles()
	if err != nil {
		a.logger.Error("Failed to list profiles", err, nil)
		return nil, fmt.Errorf("获取连接配置列表失败: %w", err)
	}
	
	return profiles, nil
}

// TestConnection 测试连接配置是否有效
func (a *App) TestConnection(profile types.ConnectionProfile) error {
	a.logger.Info("Testing connection", map[string]interface{}{
		"profile_name": profile.Name,
		"host":         profile.Host,
		"port":         profile.Port,
		"ssh_enabled":  profile.SSHEnabled,
	})
	
	err := a.connectionManager.TestConnection(profile)
	if err != nil {
		a.logger.Error("Connection test failed", err, map[string]interface{}{
			"profile_name": profile.Name,
			"host":         profile.Host,
			"port":         profile.Port,
		})
		return fmt.Errorf("连接测试失败: %w", err)
	}
	
	a.logger.Info("Connection test successful", map[string]interface{}{
		"profile_name": profile.Name,
	})
	
	return nil
}

// Connect 建立到数据库的连接
func (a *App) Connect(profileID string) error {
	a.logger.Info("Connecting to database", map[string]interface{}{
		"profile_id": profileID,
	})
	
	_, err := a.connectionManager.Connect(profileID)
	if err != nil {
		a.logger.Error("Failed to connect", err, map[string]interface{}{
			"profile_id": profileID,
		})
		
		// 发送连接失败事件
		a.emitEvent("connection:status:changed", map[string]interface{}{
			"profileId": profileID,
			"status":    "failed",
			"error":     err.Error(),
		})
		
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	
	a.logger.Info("Connected successfully", map[string]interface{}{
		"profile_id": profileID,
	})
	
	// 发送连接成功事件
	a.emitEvent("connection:status:changed", map[string]interface{}{
		"profileId": profileID,
		"status":    "connected",
	})
	
	return nil
}

// Disconnect 断开数据库连接
func (a *App) Disconnect(profileID string) error {
	err := a.connectionManager.Disconnect(profileID)
	if err != nil {
		a.logger.Error("Failed to disconnect", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return fmt.Errorf("断开连接失败: %w", err)
	}
	
	a.logger.Info("Disconnected", map[string]interface{}{
		"profile_id": profileID,
	})
	
	// 清理该连接的所有管理器实例
	a.cleanupManagersForProfile(profileID)
	
	// 发送断开连接事件
	a.emitEvent("connection:status:changed", map[string]interface{}{
		"profileId": profileID,
		"status":    "disconnected",
	})
	
	return nil
}

// GetConnectionStatus 获取连接状态
func (a *App) GetConnectionStatus(profileID string) (string, error) {
	_, err := a.connectionManager.GetConnection(profileID)
	if err != nil {
		return "disconnected", nil
	}
	
	return "connected", nil
}
