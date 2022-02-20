package app

import (
	"context"
	"gim/internal/logic/domain/message/service"
	"gim/pkg/pb"

	"google.golang.org/protobuf/proto"
)

type messageApp struct{}

var MessageApp = new(messageApp)

// SendToUser 发送消息给用户
func (*messageApp) SendToUser(ctx context.Context, sender *pb.Sender, toUserId int64, req *pb.SendMessageReq) (int64, error) {
	return service.MessageService.SendToUser(ctx, sender, toUserId, req)
}

// PushToUser 推送消息给用户，Push一般推送的是系统消息
func (*messageApp) PushToUser(ctx context.Context, userId int64, code pb.PushCode, message proto.Message, isPersist bool) error {
	return service.PushService.PushToUser(ctx, userId, code, message, isPersist)
}

// PushAll 全服推送
func (*messageApp) PushAll(ctx context.Context, req *pb.PushAllReq) error {
	return service.PushService.PushAll(ctx, req)
}

// Sync 消息同步
func (*messageApp) Sync(ctx context.Context, userId, seq int64) (*pb.SyncResp, error) {
	return service.MessageService.Sync(ctx, userId, seq)
}

// MessageAck 收到消息回执
func (*messageApp) MessageAck(ctx context.Context, userId, deviceId, ack int64) error {
	// 收到ACK后，缓存中存的seq++
	return service.DeviceAckService.Update(ctx, userId, deviceId, ack)
}

// SendMessage 发送消息
func (s *messageApp) SendMessage(ctx context.Context, sender *pb.Sender, req *pb.SendMessageReq) (int64, error) {
	// 如果发送者是用户，需要补充发送者用户的信息
	service.MessageService.AddSenderInfo(sender)

	switch req.ReceiverType {
	// 消息接收者为用户
	case pb.ReceiverType_RT_USER:
		// 发送者为用户，说明是两个好友发送信息
		if sender.SenderType == pb.SenderType_ST_USER {
			return FriendApp.SendToFriend(ctx, sender, req)
		} else {
			// 否则可以看做系统推送消息
			return s.SendToUser(ctx, sender, req.ReceiverId, req)
		}
	// 消息接收者是群组
	case pb.ReceiverType_RT_GROUP:
		return GroupApp.SendMessage(ctx, sender, req)
	}
	return 0, nil
}
