package types

type ErrorResponseStructure struct {
	Message string `json:"message"`
}

type ExistingUsersError struct {
	Err error
}

func (r *ExistingUsersError) Error() string {
	return r.Err.Error()
}

type InputError struct {
	Err error
}

func (r *InputError) Error() string {
	return r.Err.Error()
}

type MissingResourceError struct {
	Err error
}

func (r *MissingResourceError) Error() string {
	return r.Err.Error()
}

type MissingUserIdError struct {
	Err error
}

func (r *MissingUserIdError) Error() string {
	return r.Err.Error()
}

type UnauthorizedError struct {
	Err error
}

func (r *UnauthorizedError) Error() string {
	return r.Err.Error()
}

type ConflictError struct {
	Err error
}

func (r *ConflictError) Error() string {
	return r.Err.Error()
}

type InternalError struct{}

func (r *InternalError) Error() string {
	return "Internal Server Error"
}
