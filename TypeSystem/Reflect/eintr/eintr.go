package eintr

type Eintr struct {
	Name string
}

func (e *Eintr) Getname() (name string) {
	name = e.Name
	return
}

func (e *Eintr) Setname(name string) {
	e.Name = name
	return
}
