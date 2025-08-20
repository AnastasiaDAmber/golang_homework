package hw03frequencyanalysis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = false

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})
}

func TestTop10AdditionalAll(t *testing.T) {
	t.Run("single word repeated", func(t *testing.T) {
		text := "go go go"
		expected := []string{"go"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("lexicographic order tie", func(t *testing.T) {
		text := "cat bat apple bat cat dog"
		expected := []string{"bat", "cat", "apple", "dog"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("punctuation as part of word", func(t *testing.T) {
		text := "go! go, go."
		expected := []string{"go!", "go,", "go."}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("more than 10 words", func(t *testing.T) {
		text := "a b c d e f g h i j k l m n o"
		expected := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("case sensitivity", func(t *testing.T) {
		text := "Go go GO gO"
		expected := []string{"GO", "Go", "gO", "go"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("very long word", func(t *testing.T) {
		text := strings.Repeat("aVeryLongWord ", 100)
		text = strings.TrimSpace(text)
		expected := []string{"aVeryLongWord"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("words with various spaces", func(t *testing.T) {
		text := "one\ttwo  three\nfour"
		expected := []string{"four", "one", "three", "two"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("punctuation variants", func(t *testing.T) {
		text := "hello! hello!! hello!!!"
		expected := []string{"hello!", "hello!!", "hello!!!"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("more than 10 words with same freq", func(t *testing.T) {
		text := "a b c d e f g h i j k l"
		expected := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		require.Equal(t, expected, Top10(text))
	})

	t.Run("mixed case and punctuation", func(t *testing.T) {
		text := "Go go GO go! Go?"
		expected := []string{"GO", "Go", "Go?", "go", "go!"}
		require.Equal(t, expected, Top10(text))
	})
}
