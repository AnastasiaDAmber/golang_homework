package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	// Проверка, что файл поддерживается (у него есть размер)
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	size := info.Size()
	if offset > size {
		return ErrOffsetExceedsFileSize
	}

	// Постановка курсора в исходном месте
	_, err = src.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// Открытие файла для записи
	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Определение реального размера копирования
	copyLimit := size - offset
	if limit > 0 && limit < copyLimit {
		copyLimit = limit
	}

	// Создание прогресс-бара
	bar := pb.Full.Start64(copyLimit)
	barReader := bar.NewProxyReader(io.LimitReader(src, copyLimit))

	_, err = io.CopyN(dst, barReader, copyLimit)
	if err != nil && err != io.EOF {
		bar.Finish()
		return err
	}

	bar.Finish()
	fmt.Println("Copy completed")

	return nil
}
