package common

func Serial(handlers ...func() error) func() error {
	return func() error {
		for _, h := range handlers {
			if err := h(); err != nil {
				return err
			}
		}
		return nil
	}
}
