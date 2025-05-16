package rfsActionService

import (
	"context"
	"i9rfs/src/models/rfsActionModel"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	return rfsActionModel.Ls(ctx, clientUsername, directoryId)
}

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (map[string]any, error) {
	return rfsActionModel.Mkdir(ctx, clientUsername, parentDirectoryId, directoryName)
}

func deleteFilesInCS(fileIds []any) {

}

func Del(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	done, fileIds, err := rfsActionModel.Del(ctx, clientUsername, parentDirectoryId, objectIds)

	if done {
		go deleteFilesInCS(fileIds)
	}

	return done, err
}

func Trash(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	return rfsActionModel.Trash(ctx, clientUsername, parentDirectoryId, objectIds)
}

func Restore(ctx context.Context, clientUsername string, objectIds []string) (bool, error) {
	return rfsActionModel.Restore(ctx, clientUsername, objectIds)
}

func ShowTrash(ctx context.Context, clientUsername string) ([]any, error) {
	return rfsActionModel.ShowTrash(ctx, clientUsername)
}

func Rename(ctx context.Context, clientUsername, parentDirectoryId, objectId, newName string) (bool, error) {
	return rfsActionModel.Rename(ctx, clientUsername, parentDirectoryId, objectId, newName)
}

func Move(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId string, objectIds []string) (bool, error) {
	return rfsActionModel.Move(ctx, clientUsername, fromParentDirectoryId, toParentDirectoryId, objectIds)
}

func copyFilesInCS(fileCopyIdMaps []any) {

}

func Copy(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId string, objectIds []string) (bool, error) {
	done, fileCopyIdMaps, err := rfsActionModel.Copy(ctx, clientUsername, fromParentDirectoryId, toParentDirectoryId, objectIds)

	go copyFilesInCS(fileCopyIdMaps)

	return done, err
}
