package device

import (
	"gim/pkg/pb"
	"time"
)

const (
	DeviceOnLine  = 1 // 设备在线
	DeviceOffLine = 0 // 设备离线
)

// Device 设备
type Device struct {
	Id            int64     // 设备id；(登录)
	UserId        int64     // 用户id；(登录)
	Type          int32     // 设备类型,1:Android；2：IOS；3：Windows; 4：MacOS；5：Web；(注册)
	Brand         string    // 手机厂商；(注册)
	Model         string    // 机型；(注册)
	SystemVersion string    // 系统版本；(注册)
	SDKVersion    string    // SDK版本；(注册)
	Status        int32     // 在线状态，0：离线；1：在线；(登录1，下线0)
	ConnAddr      string    // 连接层服务层地址；(登录)
	ClientAddr    string    // 客户端地址；(登录)
	CreateTime    time.Time // 创建时间；(注册、登录、下线都会更新)
	UpdateTime    time.Time // 更新时间；(注册、登录、下线都会更新)(问题：两个时间一直是同步的？)
}

func (d *Device) ToProto() *pb.Device {
	return &pb.Device{
		DeviceId:      d.Id,
		UserId:        d.UserId,
		Type:          d.Type,
		Brand:         d.Brand,
		Model:         d.Model,
		SystemVersion: d.SystemVersion,
		SdkVersion:    d.SDKVersion,
		Status:        d.Status,
		ConnAddr:      d.ConnAddr,
		ClientAddr:    d.ClientAddr,
		CreateTime:    d.CreateTime.Unix(),
		UpdateTime:    d.UpdateTime.Unix(),
	}
}

func (d *Device) IsLegal() bool {
	if d.Type == 0 || d.Brand == "" || d.Model == "" ||
		d.SystemVersion == "" || d.SDKVersion == "" {
		return false
	}
	return true
}

func (d *Device) Online(userId int64, connAddr string, clientAddr string) {
	d.UserId = userId
	d.ConnAddr = connAddr
	d.ClientAddr = clientAddr
	d.Status = DeviceOnLine
}

func (d *Device) Offline(userId int64, connAddr string, clientAddr string) {
	d.UserId = userId
	d.ConnAddr = connAddr
	d.ClientAddr = clientAddr
	d.Status = DeviceOnLine
}
