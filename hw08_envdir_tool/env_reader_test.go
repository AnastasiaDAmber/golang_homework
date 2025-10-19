package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// вспомогательная функция для записи файла.
func writeFile(t *testing.T, dir, name string, data []byte) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("не удалось записать файл %s: %v", path, err)
	}
}

func TestReadDir(t *testing.T) {
	dir := t.TempDir()

	// обычный файл
	writeFile(t, dir, "FOO", []byte("foo\nsecond"))
	// файл с null-байтом и пробелами/табами в конце
	writeFile(t, dir, "BAR", []byte("   foo\x00with new line  \t\nrest"))
	// файл без перевода строки
	writeFile(t, dir, "NO_NEWLINE", []byte("value\x00more"))
	// пустой файл → переменную надо удалить
	writeFile(t, dir, "EMPTY", []byte{})
	// неправильное имя (с '=') должно игнорироваться
	writeFile(t, dir, "BAD=NAME", []byte("ignored"))

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ошибка ReadDir: %v", err)
	}

	expect := Environment{
		"FOO":        EnvValue{Value: "foo", NeedRemove: false},
		"BAR":        EnvValue{Value: "   foo\nwith new line", NeedRemove: false},
		"NO_NEWLINE": EnvValue{Value: "value\nmore", NeedRemove: false},
		"EMPTY":      EnvValue{Value: "", NeedRemove: true},
	}

	// проверка, что ожидаемые переменные совпадают
	for k, v := range expect {
		got, ok := env[k]
		if !ok {
			t.Fatalf("нет переменной %s", k)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("несовпадение для %s: получили %#v, ожидали %#v", k, got, v)
		}
	}

	if _, ok := env["BAD=NAME"]; ok {
		t.Fatalf("переменная с именем BAD=NAME не должна быть загружена")
	}
}
