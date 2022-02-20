package app

import (
	"context"
	devicedomain "gim/internal/logic/domain/device"
	"gim/pkg/gerrors"
	"gim/pkg/pb"
)

type deviceApp struct{}

var DeviceApp = new(deviceApp)

// Register 注册设备
func (*deviceApp) Register(ctx context.Context, in *pb.RegisterDeviceReq) (int64, error) {
	device := devicedomain.Device{
		Type:          in.Type,
		Brand:         in.Brand,
		Model:         in.Model,
		SystemVersion: in.SystemVersion,
		SDKVersion:    in.SdkVersion,
	}

	// 判断设备信息是否合法
	if !device.IsLegal() {
		return 0, gerrors.ErrBadRequest
	}

	// 将设备信息存到DB中。
	err := devicedomain.DeviceRepo.Save(&device)
	if err != nil {
		return 0, err
	}

	return device.Id, nil
}

// SignIn 登录
func (*deviceApp) SignIn(ctx context.Context, userId, deviceId int64, token string, connAddr string, clientAddr string) error {
	return devicedomain.DeviceService.SignIn(ctx, userId, deviceId, token, connAddr, clientAddr)
}

// Offline 设备离线
func (*deviceApp) Offline(ctx context.Context, deviceId int64, clientAddr string) error {
	// 获取设备信息
	device, err := devicedomain.DeviceRepo.Get(deviceId)
	if err != nil {
		return err
	}
	if device == nil {
		return nil
	}

	if device.ClientAddr != clientAddr {
		return nil
	}
	// 将设备的状态设置为offline
	device.Status = devicedomain.DeviceOffLine

	// 保存（更新）设备信息
	err = devicedomain.DeviceRepo.Save(device)
	if err != nil {
		return err
	}
	return nil
}

// ListOnlineByUserId 获取用户所有在线设备
func (*deviceApp) ListOnlineByUserId(ctx context.Context, userId int64) ([]*pb.Device, error) {
	return devicedomain.DeviceService.ListOnlineByUserId(ctx, userId)
}

// GetDevice 获取设备信息
func (*deviceApp) GetDevice(ctx context.Context, deviceId int64) (*pb.Device, error) {
	device, err := devicedomain.DeviceRepo.Get(deviceId)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, gerrors.ErrDeviceNotExist
	}

	return device.ToProto(), err
}

// ServerStop connect服务停止
func (*deviceApp) ServerStop(ctx context.Context, connAddr string) error {
	return devicedomain.DeviceService.ServerStop(ctx, connAddr)
}
