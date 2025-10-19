package main

import (
	"fmt"
	"os"
)

func main() {
	// проверка количества аргументов
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "использование: go-envdir <dir> <command> [args...]")
		os.Exit(1)
	}

	dir := os.Args[1]
	cmd := os.Args[2:]

	// чтение переменных окружения из каталога
	env, err := ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ошибка чтения каталога с переменными:", err)
		os.Exit(1)
	}

	// запкск команды с подготовленным окружением
	code := RunCmd(cmd, env)
	os.Exit(code)
}
