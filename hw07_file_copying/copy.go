package main

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
)

const DefaultBufSize = 1024

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer closeFile(fromFile)

	fromFileInfo, err := fromFile.Stat()
	if err != nil {
		return err
	}
	if fromFileInfo.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer closeFile(toFile)

	_, err = fromFile.Seek(offset, io.SeekStart)
	if err != nil {
		return ErrOffsetExceedsFileSize
	}

	if fromFileInfo.Size() == 0 {
		return ErrUnsupportedFile
	}

	leftToRead := defineWillReadBytes(limit, offset, fromFileInfo.Size())
	totalToRead := int(leftToRead)
	bufSize := defineBufSize(leftToRead)
	buf := make([]byte, bufSize)
	bar := pb.StartNew(totalToRead)
	for leftToRead > 0 {
		read, err := fromFile.Read(buf)
		if err != nil {
			return err
		}

		_, err = toFile.Write(buf)
		if err != nil {
			return err
		}
		leftToRead -= int64(read)
		if leftToRead < bufSize {
			buf = buf[:leftToRead]
		}
		bar.Add(read)
	}
	bar.Finish()

	return nil
}

// defineWillReadBytes определяет сколько байт будет прочитано.
func defineWillReadBytes(limit, offset, fileSize int64) (willReadBytes int64) {
	if limit > 0 {
		willReadBytes = limit
	} else {
		willReadBytes = fileSize
	}

	leftToRead := fileSize - offset
	if leftToRead < willReadBytes {
		willReadBytes = leftToRead
	}
	return
}

// defineBufSize определяет размер буфера.
func defineBufSize(willReadBytes int64) (bufSize int64) {
	bufSize = DefaultBufSize
	if willReadBytes <= bufSize {
		bufSize = willReadBytes
	}
	return
}

// closeFile закрывает файл и логирует ошибку в случае её возникновения.
func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}
