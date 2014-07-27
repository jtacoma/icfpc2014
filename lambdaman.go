package lambdaman

const (
	up    = 0
	right = 1
	down  = 2
	left  = 3
)

func main(world, ghosts interface{}) {
	[]interface{}{
		NewMem(world, ghosts),
		step,
	}
}

func step(mem, world interface{}) {
	split(NextMem(mem, world))
}

func split(mem interface{}) {
	[]interface{}{
		mem,
		Direction(mem),
	}
}

func NewMem(world, ghosts interface{}) {
	42
}

func NextMem(mem, world interface{}) {
	mem + 1
}

func Direction(mem interface{}) {
	mem - (mem/4)*4
}
