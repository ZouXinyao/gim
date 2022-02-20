package service

import (
	"context"
	"gim/internal/business/domain/user/model"
	"gim/internal/business/domain/user/repo"
	"gim/pkg/gerrors"
	"gim/pkg/pb"
	"gim/pkg/rpc"
	"time"
)

type authService struct{}

var AuthService = new(authService)

// SignIn 登录
func (*authService) SignIn(ctx context.Context, phoneNumber, code string, deviceId int64) (bool, int64, string, error) {
	// code应该在服务端也会存一份，根据phoneNumber查找DB中的code，和传入的code对比，相同就验证成功。
	if !Verify(phoneNumber, code) {
		return false, 0, "", gerrors.ErrBadCode
	}

	// 根据手机号获取用户信息，从DB中拉取user
	user, err := repo.UserRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return false, 0, "", err
	}

	// 手机号查不到用户信息，说明为新用户，新建user，存入DB
	var isNew = false
	if user == nil {
		user = &model.User{
			PhoneNumber: phoneNumber,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
		err := repo.UserRepo.Save(user)
		if err != nil {
			return false, 0, "", err
		}
		isNew = true
	}

	// 获取设备信息
	resp, err := rpc.LogicIntClient.GetDevice(ctx, &pb.GetDeviceReq{DeviceId: deviceId})
	if err != nil {
		return false, 0, "", err
	}

	// 方便测试
	token := "0"
	//token := util.RandString(40)
	// 设置token，存入cache中
	err = repo.AuthRepo.Set(user.Id, resp.Device.DeviceId, model.Device{
		Type:   resp.Device.Type,
		Token:  token,
		Expire: time.Now().AddDate(0, 3, 0).Unix(),
	})
	if err != nil {
		return false, 0, "", err
	}

	return isNew, user.Id, token, nil
}

func Verify(phoneNumber, code string) bool {
	// 假装他成功了
	return true
}

// Auth 验证用户是否登录
func (*authService) Auth(ctx context.Context, userId, deviceId int64, token string) error {
	// 根据user和deviceId获取设备信息，主要是token和超时时间Expire
	device, err := repo.AuthRepo.Get(userId, deviceId)
	if err != nil {
		return err
	}

	if device == nil {
		return gerrors.ErrUnauthorized
	}

	// 不满足条件，就表示准入失败，重新登录
	if device.Expire < time.Now().Unix() {
		return gerrors.ErrUnauthorized
	}

	if device.Token != token {
		return gerrors.ErrUnauthorized
	}
	return nil
}
