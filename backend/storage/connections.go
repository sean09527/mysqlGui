package storage

import (
	"mygui/backend/types"
	"sync"

	"gopkg.in/yaml.v3"
)

type ConnectionsStorage struct {
	stroage *LocalStorage
	mutex   sync.Mutex
}

func NewConnectionsStorage() *ConnectionsStorage {
	return &ConnectionsStorage{
		stroage: NewLocalStorage("connections.yaml"),
	}
}

func (c *ConnectionsStorage) DefaultConnections() types.Connections {
	return types.Connections{}
}

func (c *ConnectionsStorage) defaultConnectionsItem() types.ConnectionConfig {
	return types.ConnectionConfig{
		Host:              "127.0.0.1",
		Port:              3306,
		Username:          "root",
		Password:          "123456",
		Database:          "",
		MaxOpenConns:      10,
		MaxIdleConns:      10,
		ConnMaxLifetime:   3600,
		ParseTime:         true,
		Charset:           "utf8mb4",
		ReadOnly:          false,
		Loc:               "Local",
		HeartbeatInterval: 10,
		SSH: types.ConnectionSSH{
			Enable:    false,
			Host:      "",
			Port:      22,
			Username:  "",
			Password:  "",
			LoginType: "password",
			KeyPath:   "",
		},
	}
}

// 获取连接配置
func (c *ConnectionsStorage) GetConnections() (ret types.Connections) {
	conf, err := c.stroage.Load()
	ret = c.DefaultConnections()
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(conf, &ret); err != nil {
		ret = c.DefaultConnections()
		return
	}

	if len(ret) <= 0 {
		ret = c.DefaultConnections()
	}
	return
}

// 根据名称获取连接配置
func (c *ConnectionsStorage) GetConnectionByName(name string) types.ConnectionConfig {
	connections := c.GetConnections()
	for _, connection := range connections {
		if connection.Name == name {
			return connection
		}
	}

	return types.ConnectionConfig{}
}

// 保存连接配置
func (c *ConnectionsStorage) SaveConnections(connections types.Connections) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	conf, err := yaml.Marshal(connections)
	if err != nil {
		return err
	}
	return c.stroage.Save(conf)
}

// 新增连接配置
func (c *ConnectionsStorage) AddConnections(connections types.ConnectionConfig) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	name := connections.Name

	conf, err := yaml.Marshal(connections)
	if err != nil {
		return err
	}
	return c.stroage.Save(conf)
}
