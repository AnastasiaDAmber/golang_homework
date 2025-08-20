package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// ErrInvalidString возвращается, если входная строка некорректна.
var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var b strings.Builder
	var prevRune rune // предыдущая обработанная руна
	var hasPrev bool  // индикатор, что prevRune содержит символ

	for i, r := range s {
		if unicode.IsDigit(r) {
			// цифра без предыдущего символа → ошибка (пример: "3abc", "45")
			if !hasPrev {
				return "", ErrInvalidString
			}

			// конвертируем руну-цифру в число
			n, _ := strconv.Atoi(string(r))

			if n == 0 {
				// "aaa0b" => "aab": нужно удалить последний символ
				runes := []rune(b.String())
				b.Reset()
				b.WriteString(string(runes[:len(runes)-1]))
			} else {
				// символ уже добавлен один раз → добавляем ещё n-1 повторений
				b.WriteString(strings.Repeat(string(prevRune), n-1))
			}

			hasPrev = false // текущий символ "закрыт" цифрой
		} else {
			// обычная руна: записываем в результат
			b.WriteRune(r)
			prevRune = r
			hasPrev = true
		}

		// защита от многозначных чисел (например, "a10" или "aaa12b")
		if i > 0 && unicode.IsDigit(r) && unicode.IsDigit(rune(s[i-1])) {
			return "", ErrInvalidString
		}
	}

	return b.String(), nil
}
