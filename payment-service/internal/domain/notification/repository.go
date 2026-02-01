package notification

import "context"

type NotificationRepository interface {
	SendMsg(context.Context, *Message) error
}
