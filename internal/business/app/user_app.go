package app

import (
	"context"
	"gim/internal/business/domain/user/repo"
	"gim/pkg/pb"
	"time"
)

type userApp struct{}

var UserApp = new(userApp)

func (*userApp) Get(ctx context.Context, userId int64) (*pb.User, error) {
	// 根据userId获取用户信息
	user, err := repo.UserRepo.Get(userId)
	return user.ToProto(), err
}

func (*userApp) Update(ctx context.Context, userId int64, req *pb.UpdateUserReq) error {
	// 根据传入的信息更新user实例，存入DB，产出cache
	u, err := repo.UserRepo.Get(userId)
	if err != nil {
		return err
	}
	if u == nil {
		return nil
	}

	u.Nickname = req.Nickname
	u.Sex = req.Sex
	u.AvatarUrl = req.AvatarUrl
	u.Extra = req.Extra
	u.UpdateTime = time.Now()

	err = repo.UserRepo.Save(u)
	if err != nil {
		return err
	}
	return nil
}

func (*userApp) GetByIds(ctx context.Context, userIds []int64) (map[int64]*pb.User, error) {
	// 通过多个userIds获取多个user实例
	users, err := repo.UserRepo.GetByIds(userIds)
	if err != nil {
		return nil, err
	}

	pbUsers := make(map[int64]*pb.User, len(users))
	for i := range users {
		pbUsers[users[i].Id] = users[i].ToProto()
	}
	return pbUsers, nil
}

func (*userApp) Search(ctx context.Context, key string) ([]*pb.User, error) {
	// 模糊查询，key目前的实现是phone或者nickname，查询多个符合条件的User
	users, err := repo.UserRepo.Search(key)
	if err != nil {
		return nil, err
	}

	pbUsers := make([]*pb.User, len(users))
	for i, v := range users {
		pbUsers[i] = v.ToProto()
	}
	return pbUsers, nil
}
