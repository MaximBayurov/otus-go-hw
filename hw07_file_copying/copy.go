package main

import (
	"errors"
	"io"
	"log"
	"math"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrFileCreate            = errors.New("enable to create file")
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrCannotReadFromFile    = errors.New("cannot read from file")
	ErrCannotWriteToFile     = errors.New("cannot write to file")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return ErrUnsupportedFile
	}
	defer closeFile(fromFile)

	fromFileInfo, err := fromFile.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}
	if fromFileInfo.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	toFile, err := os.Create(toPath)
	if err != nil {
		return ErrFileCreate
	}
	defer closeFile(toFile)

	bufSize := defineBufSize(limit, fromFileInfo)
	buf := make([]byte, bufSize)

	bar := pb.StartNew(100.0)
	if offset > 0 {
		err := setOffset(offset, fromFile)
		if err != nil {
			return err
		}
	} else {
		_, err = fromFile.Seek(offset, io.SeekStart)
	}

	if err != nil {
		return ErrOffsetExceedsFileSize
	}
	bar.SetCurrent(33)

	leftToRead := bufSize
	for leftToRead > 0 {
		readReal, err := fromFile.Read(buf)
		var read int64
		read = int64(readReal)
		if err != nil {
			return ErrCannotReadFromFile
		}
		if limit > 0 {
			read -= countNewLines(buf)
		}

		isOverflowed := int64(readReal) < leftToRead
		if isOverflowed {
			buf = buf[:readReal]
			leftToRead = 0
		}
		_, err = toFile.Write(buf)
		if err != nil {
			return ErrCannotWriteToFile
		}
		if !isOverflowed {
			buf = buf[read:]
			leftToRead -= read
		}

		bar.SetCurrent(int64(100 - math.Floor(float64(leftToRead/bufSize)*100)))
	}
	bar.Finish()

	return nil
}

// countNewLines подсчитывает количество переносов строки в буфере.
func countNewLines(_ []byte) int64 {
	var newLinesCount int64
	return newLinesCount
	//	for _, b := range buf {
	//		if b == '\n' {
	//			newLinesCount++
	//		}
	//	}
	//	return newLinesCount
}

// setOffset устанавливает сдвиг в файле.
func setOffset(offset int64, fromFile *os.File) error {
	buf := make([]byte, offset)
	for offset > 0 {
		read, err := fromFile.Read(buf)
		if err != nil {
			return ErrCannotReadFromFile
		}

		newLinesCount := countNewLines(buf)
		read -= int(newLinesCount)
		offset -= int64(read)
		buf = buf[read:]
	}
	return nil
}

// defineBufSize определяет размер буфера.
func defineBufSize(limit int64, info os.FileInfo) int64 {
	if limit <= 0 || limit > info.Size() {
		return info.Size()
	}
	return limit
}

// closeFile закрывает файл и логирует ошибку в случае её возникновения.
func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}
