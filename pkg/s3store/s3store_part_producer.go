package s3store

import (
	"io"
	"io/ioutil"
	"os"
)

// s3PartProducer converts a stream of bytes from the reader into a stream of files on disk
type s3PartProducer struct {
	store *S3Store
	files chan<- *os.File
	done  chan struct{}
	err   error
	r     io.Reader
}

func (spp *s3PartProducer) produce(totalSize, partSize int64) {
	for {
		if totalSize <= 0 {
			close(spp.files)
			return
		}

		file, err := spp.nextPart(partSize)
		if err != nil {
			spp.err = err
			close(spp.files)
			return
		}

		select {
		case spp.files <- file:
			totalSize = totalSize - partSize
		case <-spp.done:
			close(spp.files)
			return
		}
	}
}

func (spp *s3PartProducer) nextPart(size int64) (*os.File, error) {
	// Create a temporary file to store the part
	file, err := ioutil.TempFile(spp.store.TemporaryDirectory, "tusd-s3-tmp-")
	if err != nil {
		return nil, err
	}

	limitedReader := io.LimitReader(spp.r, size)
	_, err = io.Copy(file, limitedReader)
	if err != nil {
		return nil, err
	}

	// Seek to the beginning of the file
	file.Seek(0, 0)

	return file, nil
}
