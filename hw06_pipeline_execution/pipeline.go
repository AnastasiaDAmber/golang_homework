package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func chanCloser(in In, done In) In { // chanCloser оборачивает канал in, чтобы учитывать done → удаление → изменение
	if done != nil { // если done задан → удаление → изменение
		newIn := make(Bi)
		go func() {
			defer close(newIn) // закрываем новый канал при завершении → удаление → изменение
			for {
				select {
				case <-done: // если канал done закрыт → завершение → удаление → изменение
					return
				default: // иначе продолжаем → удаление → изменение
					select { // ещё один select для проверки done и чтения из in → удаление → изменение
					case <-done: // если done закрыт → выход → удаление → изменение
						return
					case val, ok := <-in: // читаем значение из in → удаление → изменение
						if !ok { // если in закрыт → выходим → удаление → изменение
							return
						}
						select { // ещё один select для отправки значения в newIn или проверки done → удаление → изменение
						case <-done: // если done закрыт → выход → удаление → изменение
							return
						case newIn <- val: // отправляем значение дальше → удаление → изменение
						}
					}
				}
			}
		}()
		return newIn // возвращаем новый обёрнутый канал → удаление → изменение
	}
	return in // если done == nil, возвращаем исходный канал → удаление → изменение
}

func ExecutePipeline(in In, done In, stages ...Stage) Out { // запуск пайплайна → удаление → изменение
	if in == nil { // проверка на nil → удаление → изменение
		panic("Input channel is nil") // паника, если канал nil → удаление → изменение
	}
	for _, stage := range stages { // проходим по всем стадиям → удаление → изменение
		in = stage(chanCloser(in, done)) // передаём в stage обёрнутый канал с done → удаление → изменение
	}
	return in // возвращаем канал последней стадии → удаление → изменение
}
