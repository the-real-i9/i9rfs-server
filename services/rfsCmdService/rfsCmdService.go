package rfsCmdService

import (
	"context"
	"i9rfs/server/models/rfsCmdModel"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	return rfsCmdModel.Ls(ctx, clientUsername, directoryId)
}
