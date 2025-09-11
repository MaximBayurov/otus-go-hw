package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		out := make(Bi)
		close(out)
		return out
	}

	out := in
	out = decorateChanWithDone(out, done)
	for _, stage := range stages {
		out = stage(out)
		out = decorateChanWithDone(out, done)
	}

	return out
}

// decorateChanWithDone декорирует канал добавляя возможность закрыть его.
func decorateChanWithDone(in In, done In) Out {
	res := make(Bi)
	go func() {
		defer flashChan(in)
		defer close(res)
		for {
			select {
			case <-done:
				return
			case r, ok := <-in:
				if !ok {
					return
				}
				res <- r
			}
		}
	}()
	return res
}

// flashChan дочитывает из незакрытого канала воизбежание блокировки
// поскольку сами не можем закрыть канал.
func flashChan(in In) {
	for {
		_, ok := <-in
		if !ok {
			return
		}
	}
}
