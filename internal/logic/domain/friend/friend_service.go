package friend

import (
	"context"
	"gim/internal/logic/proxy"
	"gim/pkg/gerrors"
	"gim/pkg/pb"
	"gim/pkg/rpc"
	"time"
)

type friendService struct{}

var FriendService = new(friendService)

// List 获取好友列表
func (s *friendService) List(ctx context.Context, userId int64) ([]*pb.Friend, error) {
	// 获取friends列表
	friends, err := FriendRepo.List(userId, FriendStatusAgree)
	if err != nil {
		return nil, err
	}

	userIds := make(map[int64]int32, len(friends))
	for i := range friends {
		userIds[friends[i].FriendId] = 0
	}
	// 获取每个friend的用户信息
	resp, err := rpc.BusinessIntClient.GetUsers(ctx, &pb.GetUsersReq{UserIds: userIds})
	if err != nil {
		return nil, err
	}

	// 构建并发送friends的详细信息
	var infos = make([]*pb.Friend, len(friends))
	for i := range friends {
		friend := pb.Friend{
			UserId:  friends[i].FriendId,
			Remarks: friends[i].Remarks,
			Extra:   friends[i].Extra,
		}

		user, ok := resp.Users[friends[i].FriendId]
		if ok {
			friend.Nickname = user.Nickname
			friend.Sex = user.Sex
			friend.AvatarUrl = user.AvatarUrl
			friend.UserExtra = user.Extra
		}
		infos[i] = &friend
	}

	return infos, nil
}

// AddFriend 添加好友
func (*friendService) AddFriend(ctx context.Context, userId, friendId int64, remarks, description string) error {
	// 获取当前ID的好友信息，申请过就直接返回。
	friend, err := FriendRepo.Get(userId, friendId)
	if err != nil {
		return err
	}
	if friend != nil {
		if friend.Status == FriendStatusApply {
			return nil
		}
		if friend.Status == FriendStatusAgree {
			return gerrors.ErrAlreadyIsFriend
		}
	}

	// 保存好友实例
	now := time.Now()
	err = FriendRepo.Save(&Friend{
		UserId:     userId,
		FriendId:   friendId,
		Remarks:    remarks,
		Status:     FriendStatusApply,
		CreateTime: now,
		UpdateTime: now,
	})
	if err != nil {
		return err
	}

	// 获取当前用户信息
	resp, err := rpc.BusinessIntClient.GetUser(ctx, &pb.GetUserReq{UserId: userId})
	if err != nil {
		return err
	}

	// 将好友申请信息发送给friendId用户
	err = proxy.MessageProxy.PushToUser(ctx, friendId, pb.PushCode_PC_ADD_FRIEND, &pb.AddFriendPush{
		FriendId:    userId,
		Nickname:    resp.User.Nickname,
		AvatarUrl:   resp.User.AvatarUrl,
		Description: description,
	}, true)
	if err != nil {
		return err
	}
	return nil
}

// AgreeAddFriend 同意添加好友
func (*friendService) AgreeAddFriend(ctx context.Context, userId, friendId int64, remarks string) error {
	friend, err := FriendRepo.Get(friendId, userId)
	if err != nil {
		return err
	}
	if friend == nil {
		return gerrors.ErrBadRequest
	}
	if friend.Status == FriendStatusAgree {
		return nil
	}
	friend.Status = FriendStatusAgree
	err = FriendRepo.Save(friend)
	if err != nil {
		return err
	}

	now := time.Now()
	err = FriendRepo.Save(&Friend{
		UserId:     userId,
		FriendId:   friendId,
		Remarks:    remarks,
		Status:     FriendStatusAgree,
		CreateTime: now,
		UpdateTime: now,
	})
	if err != nil {
		return err
	}

	resp, err := rpc.BusinessIntClient.GetUser(ctx, &pb.GetUserReq{UserId: userId})
	if err != nil {
		return err
	}

	// 将同意添加好友的信息发送给friendId的用户
	err = proxy.MessageProxy.PushToUser(ctx, friendId, pb.PushCode_PC_AGREE_ADD_FRIEND, &pb.AgreeAddFriendPush{
		FriendId:  userId,
		Nickname:  resp.User.Nickname,
		AvatarUrl: resp.User.AvatarUrl,
	}, true)
	if err != nil {
		return err
	}
	return nil
}

// SendToFriend 消息发送至好友
func (*friendService) SendToFriend(ctx context.Context, sender *pb.Sender, req *pb.SendMessageReq) (int64, error) {
	// 发给发送者，因为发送消息后，也要在界面上显示自己发的内容。
	seq, err := proxy.MessageProxy.SendToUser(ctx, sender, sender.SenderId, req)
	if err != nil {
		return 0, err
	}

	// 发给接收者
	_, err = proxy.MessageProxy.SendToUser(ctx, sender, req.ReceiverId, req)
	if err != nil {
		return 0, err
	}

	return seq, nil
}
