package server

// Command .
type Command interface {
	exec() (string, error) // (rp, error)
}
