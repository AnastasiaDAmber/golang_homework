package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		// uncomment if task with asterisk completed
		// {input: `qwe\4\5`, expected: `qwe45`},
		// {input: `qwe\45`, expected: `qwe44444`},
		// {input: `qwe\\5`, expected: `qwe\\\\\`},
		// {input: `qwe\\\3`, expected: `qwe\3`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b"}
	for _, tc := range invalidStrings {
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

func TestUnpackComprehensive(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		// обычные повторения
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"aaa0b", "aab", false},
		{"a0b0c", "c", false},   // несколько нулей
		{"a1b1c", "abc", false}, // повтор 1

		// UTF-8 символы
		{"🙂2🙃3", "🙂🙂🙃🙃🙃", false},
		{"ф2я3", "ффяяя", false},
		{"ф0я", "я", false},

		// спецсимволы
		{"a\n3b", "a\n\n\nb", false},
		{"\t2x", "\t\tx", false},

		// пустая строка
		{"", "", false},

		// ошибки
		{"3abc", "", true},   // цифра в начале
		{"45", "", true},     // только цифры
		{"aaa10b", "", true}, // многозначная цифра
		{"a12b", "", true},   // две цифры подряд
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			if tc.expectError {
				require.Error(t, err)
				require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}
