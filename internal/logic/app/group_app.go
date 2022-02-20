package app

import (
	"context"
	"gim/internal/logic/domain/group/model"
	"gim/internal/logic/domain/group/repo"
	"gim/pkg/pb"
)

type groupApp struct{}

var GroupApp = new(groupApp)

// CreateGroup 创建群组
func (*groupApp) CreateGroup(ctx context.Context, userId int64, in *pb.CreateGroupReq) (int64, error) {
	// 创建群组实例，将userId设为管理员，其他member添加到成员中。
	group := model.CreateGroup(userId, in)
	// group存到DB中，group和groupuser分开存储(根据创建表的SQL也可以看出来，group和groupuser是两个表)。
	err := repo.GroupRepo.Save(group)
	if err != nil {
		return 0, err
	}
	return group.Id, nil
}

// GetGroup 获取群组信息
func (*groupApp) GetGroup(ctx context.Context, groupId int64) (*pb.Group, error) {
	// group和groupuser中的内容都通过cache或者DB查询到。
	group, err := repo.GroupRepo.Get(groupId)
	if err != nil {
		return nil, err
	}

	return group.ToProto(), nil
}

// GetUserGroups 获取用户加入的群组列表
func (*groupApp) GetUserGroups(ctx context.Context, userId int64) ([]*pb.Group, error) {
	// 用SQL语句直接查询
	groups, err := repo.GroupUserRepo.ListByUserId(userId)
	if err != nil {
		return nil, err
	}

	// 转存到PB结构中
	pbGroups := make([]*pb.Group, len(groups))
	for i := range groups {
		pbGroups[i] = groups[i].ToProto()
	}
	return pbGroups, nil
}

// Update 更新群组
func (*groupApp) Update(ctx context.Context, userId int64, update *pb.UpdateGroupReq) error {
	group, err := repo.GroupRepo.Get(update.GroupId)
	if err != nil {
		return err
	}

	// 更新group实例
	err = group.Update(ctx, update)
	if err != nil {
		return err
	}

	// 将新的group存到DB中，删除cache
	err = repo.GroupRepo.Save(group)
	if err != nil {
		return err
	}

	// 向群组内推送群组更新的消息
	err = group.PushUpdate(ctx, userId)
	if err != nil {
		return err
	}
	return nil
}

// AddMembers 添加群组成员
func (*groupApp) AddMembers(ctx context.Context, userId, groupId int64, userIds []int64) ([]int64, error) {
	group, err := repo.GroupRepo.Get(groupId)
	if err != nil {
		return nil, err
	}
	// 更新group实例中的内容
	existIds, addedIds, err := group.AddMembers(ctx, userIds)
	if err != nil {
		return nil, err
	}
	// 更新DB中的内容，删除cache
	err = repo.GroupRepo.Save(group)
	if err != nil {
		return nil, err
	}

	// 向群组内推送添加成员的消息
	err = group.PushAddMember(ctx, userId, addedIds)
	if err != nil {
		return nil, err
	}
	return existIds, nil
}

// UpdateMember 更新群组用户
func (*groupApp) UpdateMember(ctx context.Context, in *pb.UpdateGroupMemberReq) error {
	group, err := repo.GroupRepo.Get(in.GroupId)
	if err != nil {
		return err
	}
	// 根据userid找到群组的成员，更新该user在group的信息
	err = group.UpdateMember(ctx, in)
	if err != nil {
		return err
	}
	// 更新DB中的内容，删除cache
	err = repo.GroupRepo.Save(group)
	if err != nil {
		return err
	}
	return nil
}

// DeleteMember 删除群组成员
func (*groupApp) DeleteMember(ctx context.Context, groupId int64, userId int64, optId int64) error {
	group, err := repo.GroupRepo.Get(groupId)
	if err != nil {
		return err
	}
	// userid成员标记为delete
	err = group.DeleteMember(ctx, userId)
	if err != nil {
		return err
	}
	// DB中group中的成员，删除cache
	err = repo.GroupRepo.Save(group)
	if err != nil {
		return err
	}

	// 向group推送移除成员的消息
	err = group.PushDeleteMember(ctx, optId, userId)
	if err != nil {
		return err
	}
	return nil
}

// GetMembers 获取群组成员
func (*groupApp) GetMembers(ctx context.Context, groupId int64) ([]*pb.GroupMember, error) {
	group, err := repo.GroupRepo.Get(groupId)
	if err != nil {
		return nil, err
	}
	// group实例中不仅包含了group信息，也包含了groupuser的信息
	return group.GetMembers(ctx)
}

// SendMessage 向群组发送消息
func (*groupApp) SendMessage(ctx context.Context, sender *pb.Sender, req *pb.SendMessageReq) (int64, error) {
	group, err := repo.GroupRepo.Get(req.ReceiverId)
	if err != nil {
		return 0, err
	}

	// 向group中推送消息
	return group.SendMessage(ctx, sender, req)
}
