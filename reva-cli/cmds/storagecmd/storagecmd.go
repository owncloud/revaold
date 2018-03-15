package storagecmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"gitlab.com/labkode/reva/api"
	"gitlab.com/labkode/reva/reva-cli/util"

	"github.com/codegangsta/cli"
	"github.com/ryanuber/columnize"
)

var EmptyRecycleCommand = cli.Command{
	Name:      "recycle-purge",
	Usage:     "Purge recycle entries",
	ArgsUsage: "Usage: recycle-purge",
	Action:    emptyRecycle,
}

var ListRecycleCommand = cli.Command{
	Name:      "recycle-list",
	Usage:     "List recycle entries",
	ArgsUsage: "Usage: recycle-list",
	Action:    listRecycle,
}

var RestoreRecycleEntryCommand = cli.Command{
	Name:      "recycle-restore",
	Usage:     "Restore a recycle entry",
	ArgsUsage: "Usage: recycle-restore <restore-key>",
	Action:    restoreRecycleEntry,
}

var ListRevisionsCommand = cli.Command{
	Name:      "rev-list",
	Usage:     "List revisions of a file",
	ArgsUsage: "Usage: rev-list- <path>",
	Action:    listRevisions,
}

var DownloadRevisionCommand = cli.Command{
	Name:      "rev-download",
	Usage:     "Download a revision of a file",
	ArgsUsage: "Usage: rev-download <path> <rev-key>",
	Action:    downloadRevision,
}

var RestoreRevisionCommand = cli.Command{
	Name:      "rev-restore",
	Usage:     "Restore a revision of a file",
	ArgsUsage: "Usage: rev-restore <path> <rev-key>",
	Action:    restoreRevision,
}

var InspectCommand = cli.Command{
	Name:      "inspect",
	Aliases:   []string{"i", "in", "ins"},
	Usage:     "Inspect a namespace entry",
	ArgsUsage: "Usage: inspect <path>",
	Action:    inspect,
}

var ListFolderCommand = cli.Command{
	Name:      "list",
	Usage:     "List the contents of a folder",
	ArgsUsage: "Usage: list <path>",
	Action:    listFolder,
}

var DeleteCommand = cli.Command{
	Name:      "delete",
	Usage:     "Delete an entry",
	ArgsUsage: "Usage: delete <path>",
	Action:    deleteEntry,
}

var DownloadFileCommand = cli.Command{
	Name:      "download",
	Usage:     "Download a file",
	ArgsUsage: "Usage: download <path> <localpath>",
	Action:    download,
}

var UploadFileCommand = cli.Command{
	Name:      "upload",
	Usage:     "Upload a file",
	ArgsUsage: "Usage: upload <path> <localpath>",
	Action:    upload,
}

var MoveCommand = cli.Command{
	Name:      "move",
	Usage:     "Move a file or folder",
	ArgsUsage: "Usage: move <old-path> <new-path>",
	Action:    move,
}

func inspect(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	md, err := client.Inspect(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	_type := "File"
	if md.IsDir {
		_type = "Directory"
	}

	dateTime := time.Unix(int64(md.Mtime), 0).Format(time.RFC3339)

	fmt.Fprintf(c.App.Writer, "%s: %s\nID: %s\nSize: %d\nModify: %s Timestamp: %d\nETag: %s\n", _type, md.Path, md.Id, md.Size, dateTime, md.Mtime, md.Etag)
	return nil
}

func listFolder(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	stream, err := client.ListFolder(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	lines := []string{"#Type|Id|Size|MTime|ETag|Path"}
	for {
		md, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		_type := "file"
		if md.IsDir {
			_type = "dir"
		}
		line := fmt.Sprintf("%s|%s|%d|%d|%s|%s", _type, md.Id, md.Size, md.Mtime, md.Etag, md.Path)
		lines = append(lines, line)
	}
	fmt.Fprintln(c.App.Writer, columnize.SimpleFormat(lines))
	return nil
}

func download(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	localFile := c.Args().Get(1)
	if localFile == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	fd, err := os.OpenFile(localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)
	defer fd.Close()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	stream, err := client.ReadFile(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	var reader io.Reader
	for {
		dc, err := stream.Recv()
		if dc != nil {
			if dc.Length > 0 {
				reader = bytes.NewReader(dc.Data)
				_, err := io.CopyN(fd, reader, int64(dc.Length))
				if err != nil {
					return cli.NewExitError(err, 1)
				}
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}

	}
	return nil
}

func upload(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	localFile := c.Args().Get(1)
	if localFile == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	fd, err := os.Open(localFile)
	defer fd.Close()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	ctx := util.GetContextWithAuth()
	txInfo, err := client.StartWriteTx(ctx, &api.Empty{})
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	stream, err := client.WriteChunk(ctx)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	// send data chunks of maximum 3 MiB
	buffer := make([]byte, 1024*1024*3)
	offset := uint64(0)
	numChunks := uint64(0)
	for {
		n, err := fd.Read(buffer)
		if n > 0 {
			dc := &api.TxChunk{
				TxId:   txInfo.TxId,
				Length: uint64(n),
				Data:   buffer,
				Offset: offset,
			}
			if err := stream.Send(dc); err != nil {
				return cli.NewExitError(err, 1)
			}
			numChunks++
			offset += uint64(n)

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	// all the chunks have been sent, we need to close the tx
	_, err = client.FinishWriteTx(ctx, &api.TxEnd{Path: path, TxId: txInfo.TxId})
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

func move(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	oldPath := c.Args().First()
	if oldPath == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	newPath := c.Args().Get(1)
	if newPath == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.MoveReq{OldPath: oldPath, NewPath: newPath}
	_, err = client.Move(util.GetContextWithAuth(), req)
	if err != nil {
		return err
	}
	return nil
}

func deleteEntry(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	_, err = client.Delete(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

func emptyRecycle(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	_, err = client.EmptyRecycle(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

func listRecycle(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	stream, err := client.ListRecycle(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	lines := []string{"#Type|RestoreKey|Deleted|Size|RestorePath"}
	for {
		re, err := stream.Recv()
		if re != nil {
			_type := "file"
			if re.IsDir {
				_type = "directory"
			}
			line := fmt.Sprintf("%s|%s|%d|%d|%s", _type, re.RestoreKey, re.DelMtime, re.Size, re.RestorePath)
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	fmt.Fprintln(c.App.Writer, columnize.SimpleFormat(lines))
	return nil
}

func restoreRecycleEntry(c *cli.Context) error {
	restoreKey := c.Args().First()
	if restoreKey == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.RecycleEntryReq{RestoreKey: restoreKey}
	_, err = client.RestoreRecycleEntry(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil

}

func listRevisions(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.PathReq{Path: path}
	stream, err := client.ListRevisions(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	lines := []string{"#Type|RevisionKey|Modified|Size"}
	for {
		rev, err := stream.Recv()
		if rev != nil {
			_type := "file"
			if rev.IsDir {
				_type = "directory"
			}
			line := fmt.Sprintf("%s|%s|%d|%d", _type, rev.RevKey, rev.Mtime, rev.Size)
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	fmt.Fprintln(c.App.Writer, columnize.SimpleFormat(lines))
	return nil
}

func downloadRevision(c *cli.Context) error {
	if len(c.Args()) < 3 {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	revKey := c.Args().Get(1)
	if revKey == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	localFile := c.Args().Get(2)
	if localFile == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	fd, err := os.OpenFile(localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)
	defer fd.Close()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.RevisionReq{Path: path, RevKey: revKey}
	stream, err := client.ReadRevision(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	var reader io.Reader
	for {
		dc, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if dc.Length > 0 {
			reader = bytes.NewReader(dc.Data)
			_, err := io.CopyN(fd, reader, int64(dc.Length))
			if err != nil {
				return cli.NewExitError(err, 1)
			}
		}
	}
	return nil
}

func restoreRevision(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}
	path := c.Args().First()
	if path == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	revKey := c.Args().Get(1)
	if revKey == "" {
		return cli.NewExitError(c.Command.ArgsUsage, 1)
	}

	client, err := util.GetStorageClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	req := &api.RevisionReq{Path: path, RevKey: revKey}
	_, err = client.RestoreRevision(util.GetContextWithAuth(), req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}
