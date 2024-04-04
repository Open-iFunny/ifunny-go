package ifunny

type Result[T any] struct {
	V   T
	Err error
}

type Iterator[T any] struct {
	Iter func() <-chan Result[T]
	Stop func()
}

func iterFrom[T any](next func() (T, bool, error)) Iterator[T] {
	done := make(chan struct{})
	inner, outer := make(chan Result[T]), make(chan Result[T])
	return Iterator[T]{
		Stop: func() {
			close(done)
		},
		Iter: func() <-chan Result[T] {
			go func() {
				for {
					v, more, err := next()
					if err != nil {
						inner <- Result[T]{Err: err}
						return
					}

					if !more {
						close(done)
						return
					}

					inner <- Result[T]{V: v}
				}
			}()

			go func() {
				select {
				case v := <-inner:
					outer <- v
				case <-done:
					close(outer)
					return
				}
			}()

			return outer
		},
	}
}
