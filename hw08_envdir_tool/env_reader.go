package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Environment map[string]EnvValue

type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir читает указанный каталог и формирует карту переменных окружения.
// Каждая переменная окружения хранится в отдельном файле:
//
//	имя файла — это имя переменной,
//	первая строка файла — значение переменной.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment)

	for _, e := range entries {
		// пропуск директории
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// имя переменной не должно содержать символ "="
		if bytes.ContainsRune([]byte(name), '=') {
			continue
		}

		path := filepath.Join(dir, name)
		f, err := os.Open(path)
		if err != nil {
			// если файл исчез между ReadDir и Open, просто пропуск
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}

		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}

		// если файл пустой → переменная удаляется
		if len(data) == 0 {
			env[name] = EnvValue{Value: "", NeedRemove: true}
			continue
		}

		// берется только первая строка (до '\n')
		idx := bytes.IndexByte(data, '\n')
		var firstLine []byte
		if idx == -1 {
			firstLine = data
		} else {
			firstLine = data[:idx]
		}

		// замена всех нулевых байтов (0x00) на перевод строки
		firstLine = bytes.ReplaceAll(firstLine, []byte{0}, []byte{'\n'})

		// удаление пробелов и табуляций справа
		firstLine = bytes.TrimRight(firstLine, " \t")

		env[name] = EnvValue{Value: string(firstLine), NeedRemove: false}
	}

	return env, nil
}
