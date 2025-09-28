package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// - возвращается код выхода запущенной программы.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 0
	}

	//nolint:gosec
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// берется текущее окружение
	baseEnv := os.Environ()
	envMap := map[string]string{}
	for _, kv := range baseEnv {
		parts := strings.SplitN(kv, "=", 2)
		k := parts[0]
		v := ""
		if len(parts) > 1 {
			v = parts[1]
		}
		envMap[k] = v
	}

	// применение изменений из env
	for k, v := range env {
		if v.NeedRemove {
			delete(envMap, k) // если файл пустой, то переменная удаляется
		} else {
			envMap[k] = v.Value // иначе задается новое значение
		}
	}

	// сбор окружение в слайс
	newEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		newEnv = append(newEnv, k+"="+v)
	}
	c.Env = newEnv

	// запуск команду
	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		// если команда завершилась с ошибкой, то достается код выхода
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
			return 1
		}
		return 1
	}

	return 0
}
