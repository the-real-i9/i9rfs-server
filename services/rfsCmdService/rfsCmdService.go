package rfsCmdService

import (
	"context"
	"i9rfs/models/rfsCmdModel"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	return rfsCmdModel.Ls(ctx, clientUsername, directoryId)
}

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (any, error) {

	return nil, nil
}
