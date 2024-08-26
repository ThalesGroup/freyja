package internal

var ErrUserInput *UserInputError

type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

type ConfigurationVersionError struct {
	Message string
}

func (e *ConfigurationVersionError) Error() string {
	return e.Message
}

type UserInputError struct {
	Message string
}

func (e *UserInputError) Error() string { return e.Message }
