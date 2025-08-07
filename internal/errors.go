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

// *******
// NETWORK
// *******

type NetworkError struct {
	Network string
	Message string
}

func (e *NetworkError) Error() string { return e.Message }

type NetworkAlreadyExistsError struct {
	// network name
	Network string
	// message to implement Error interface
	Message string
}

func (e *NetworkAlreadyExistsError) Error() string { return e.Message }

type NetworkNotActiveError struct {
	Network string
	Message string
}

func (e *NetworkNotActiveError) Error() string { return e.Message }
