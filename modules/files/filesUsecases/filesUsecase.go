package filesUsecases

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirawann/arigato-shop/config"
	"github.com/sirawann/arigato-shop/modules/files"
)

type IFilesUsecase interface {
	UploadToGCP(req []*files.FileReq) ([]*files.FileRes, error)
	DeleteFileOnGCP(req []*files.DeleteFileReq) error
}

type filesUsecase struct {
	cfg config.IConfig
}

func FilesUsecase(cfg config.IConfig) IFilesUsecase {
	return &filesUsecase{
		cfg: cfg,
	}
}

type filePub struct {
	bucket      string
	destination string
	file        *files.FileRes
}

func (f *filePub) makePublic(ctx context.Context, client *storage.Client) error {
	acl := client.Bucket(f.bucket).Object(f.destination).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("ACLHandle.Set: %w", err)
	}
	fmt.Printf("Blob %v is now publicly accessible.\n", f.destination)
	return nil
}

func (u *filesUsecase) uploadWorkers(ctx context.Context, client *storage.Client, jobs <-chan *files.FileReq, results chan<- *files.FileRes, errs chan<- error) {
	for job := range jobs {
		cotainer, err := job.File.Open()
		if err != nil {
			errs <- err
			return
		}

		b, err := ioutil.ReadAll(cotainer)
		if err != nil {
			errs <- err
			return
		}

		buf := bytes.NewBuffer(b)

		// Upload an object with storage.Writer.
		wc := client.Bucket(u.cfg.App().GCPbucket()).Object(job.Destination).NewWriter(ctx)

		if _, err = io.Copy(wc, buf); err != nil {
			errs <- fmt.Errorf("io.Copy: %w", err)
			return
		}
		// Data can continue to be added to the file until the writer is closed.
		if err := wc.Close(); err != nil {
			errs <- fmt.Errorf("Writer.Close: %w", err)
			return
		}
		fmt.Printf("%v uploaded to %v.\n", job.FileName, job.Extension)

		newFile := &filePub{
			file: &files.FileRes{
				FileName: job.FileName,
				Url:      fmt.Sprintf("http://storage.googleapis.com/%s%s", u.cfg.App().GCPbucket(), job.Destination),
			},
			bucket:      u.cfg.App().GCPbucket(),
			destination: job.Destination,
		}

		if err := newFile.makePublic(ctx, client); err != nil {
			errs <- err
			return
		}

		errs <- nil
		results <- newFile.file
	}

}

func (u *filesUsecase) UploadToGCP(req []*files.FileReq) ([]*files.FileRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	jobsCh := make(chan *files.FileReq, len(req))
	resultsCh := make(chan *files.FileRes, len(req))
	errsCh := make(chan error, len(req))

	res := make([]*files.FileRes, 0)

	for _, r := range req {
		jobsCh <- r
	}
	close(jobsCh)

	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go u.uploadWorkers(ctx, client, jobsCh, resultsCh, errsCh)
	}

	for a := 0; a < len(req); a++ {
		err := <-errsCh
		if err != nil {
			return nil, err
		}
		result := <-resultsCh
		res = append(res, result)
	}

	return res, nil
}

func (u *filesUsecase) deleteFileWorkers(ctx context.Context, client *storage.Client, jobs <-chan *files.DeleteFileReq, errs chan<- error) {
	for job := range jobs {
		o := client.Bucket(u.cfg.App().GCPbucket()).Object(job.Destination)

		// Optional: set a generation-match precondition to avoid potential race
		// conditions and data corruptions. The request to delete the file is aborted
		// if the object's generation number does not match your precondition.
		attrs, err := o.Attrs(ctx)
		if err != nil {
			errs <- fmt.Errorf("object.Attrs: %v", err)
			return
		}
		o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

		if err := o.Delete(ctx); err != nil {
			errs <- fmt.Errorf("Object(%q).Delete: %v", job.Destination, err)
			return
		}
		fmt.Printf("Blob %v deleted.\n", job.Destination)

		errs <- nil
	}
}

func (u *filesUsecase) DeleteFileOnGCP(req []*files.DeleteFileReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	jobsCh := make(chan *files.DeleteFileReq, len(req))
	errsCh := make(chan error, len(req))

	for _, r := range req {
		jobsCh <- r
	}
	close(jobsCh)

	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go u.deleteFileWorkers(ctx, client, jobsCh, errsCh)
	}

	for a := 0; a < len(req); a++ {
		err := <-errsCh
		return err
	}
	return nil
}

func (u *filesUsecase) deleteFromStorageFileWorkers(ctx context.Context, jobs <-chan *files.DeleteFileReq, errs chan<- error) {
	for job := range jobs {
		if err := os.Remove("./assets/images/" + job.Destination); err != nil {
			errs <- fmt.Errorf("remove file: %s failed: %v", job.Destination, err)
			return
		}
		errs <- nil
	}
}

func (u *filesUsecase) DeleteFileOnStorage(req []*files.DeleteFileReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	jobsCh := make(chan *files.DeleteFileReq, len(req))
	errsCh := make(chan error, len(req))

	for _, r := range req {
		jobsCh <- r
	}
	close(jobsCh)

	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go u.deleteFromStorageFileWorkers(ctx, jobsCh, errsCh)
	}

	for a := 0; a < len(req); a++ {
		err := <-errsCh
		return err
	}
	return nil
}


