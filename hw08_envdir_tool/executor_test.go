package main

import "testing"

func TestRunCmd(t *testing.T) {
	// задается окружение: удаление UNSET и установка FOO=bar
	env := Environment{
		"UNSET": EnvValue{Value: "", NeedRemove: true},
		"FOO":   EnvValue{Value: "bar", NeedRemove: false},
	}

	// проверка, что UNSET пустая, а FOO=bar
	cmd := []string{"/bin/sh", "-c", `if [ -z "${UNSET}" ] && [ "$FOO" = "bar" ]; then exit 42; else exit 1; fi`}

	code := RunCmd(cmd, env)
	if code != 42 {
		t.Fatalf("ожидали код выхода 42, получили %d", code)
	}
}
