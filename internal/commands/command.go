package commands

type Command interface {
	callback(...string) error
}
