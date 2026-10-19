package httpclient

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/client"
)

type Notifier struct{ Client *client.Client }

func (n *Notifier) Publish(ctx context.Context, owner, eventID, title, message string) error {
	return n.Client.Do(ctx, owner, "notification", "POST", "/internal/notifications", map[string]string{"event_id": eventID, "title": title, "message": message}, nil)
}
