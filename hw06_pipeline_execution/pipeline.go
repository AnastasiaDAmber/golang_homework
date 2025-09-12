package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	cur := orDone(done, in)
	for _, s := range stages {
		cur = orDone(done, s(cur)) // каждый stage получает поток, обёрнутый в orDone
	}
	return cur
}

// orDone возвращает канал, который читает из in и перенаправляет значения в out,
// но прекращает работу и закрывает out, если closed(done).
func orDone(done In, in In) Out {
	if done == nil { // если done == nil, то просто пробрасывается канал как есть
		return in
	}

	out := make(Bi) // создание промежуточного канала out
	go func() {
		defer close(out)
		for {
			select {
			case <-done: // случай: канал отмены закрыт
				return
			case v, ok := <-in: // чтение из входного канала
				if !ok {
					return
				}
				select { // попытка отправить значение дальше или обработать отмену
				case out <- v: // успешная отправка значения в out
				case <-done: // отмена во время отправки
					return
				}
			}
		}
	}()
	return out
}
