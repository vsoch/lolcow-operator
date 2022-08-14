package lolcow

// Greeter is the cow that is going to greet us
type Greeter struct {
	Greeting string
}

func NewGreeter(greeting string) *Greeter {
	return &Greeter{Greeting: greeting}
}
