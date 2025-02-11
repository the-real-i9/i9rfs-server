package rfsCmdService

import (
	"context"
	"i9rfs/models/rfsCmdModel"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	return rfsCmdModel.Ls(ctx, clientUsername, directoryId)
}

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (map[string]any, error) {
	return rfsCmdModel.Mkdir(ctx, clientUsername, parentDirectoryId, directoryName)
}

func deleteFilesInCS(fileIds []any) {

}

func Del(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	done, fileIds, err := rfsCmdModel.Del(ctx, clientUsername, parentDirectoryId, objectIds)

	go deleteFilesInCS(fileIds)

	return done, err
}

func Trash(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	return rfsCmdModel.Trash(ctx, clientUsername, parentDirectoryId, objectIds)
}

func Restore(ctx context.Context, clientUsername string, objectIds []string) (bool, error) {
	return rfsCmdModel.Restore(ctx, clientUsername, objectIds)
}

func ShowTrash(ctx context.Context, clientUsername string) ([]any, error) {
	return rfsCmdModel.ShowTrash(ctx, clientUsername)
}

func Rename(ctx context.Context, clientUsername, parentDirectoryId, objectId, newName string) (bool, error) {
	return rfsCmdModel.Rename(ctx, clientUsername, parentDirectoryId, objectId, newName)
}
