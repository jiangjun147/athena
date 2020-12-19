package common

import "fmt"

type ParellelError struct {
	errors []error
}

func (pe *ParellelError) IsNil() bool {
	for _, err := range pe.errors {
		if err != nil {
			return false
		}
	}
	return true
}

func (pe *ParellelError) Add(err error) {
	pe.errors = append(pe.errors, err)
}

func (pe *ParellelError) Error() string {
	return fmt.Sprintf("%+v", pe.errors)
}

func Parellel(handlers ...func() error) func() error {
	return func() error {
		ch := make(chan error)
		for _, h := range handlers {
			go func(h func() error) {
				var err error
				defer func() {
					ch <- err
				}()

				err = h()
			}(h)
		}

		pe := ParellelError{}
		for i := 0; i < len(handlers); i++ {
			pe.Add(<-ch)
		}
		close(ch)

		if pe.IsNil() {
			return nil
		}
		return &pe
	}
}
