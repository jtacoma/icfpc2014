package lambdaman

const (
	up    = 0
	right = 1
	down  = 2
	left  = 3
)

func main(world, undocumented interface{}) {
	[]interface{}{
		NewMem(world, undocumented),
		step,
	}
}

func step(mem, world interface{}) {
	DBUG(get(mem, iturn))
	Response(NextMem(mem, world))
}

func Response(mem interface{}) {
	[]interface{}{
		mem,
		Direction(mem),
	}
}

func NewMem(world, undocumented interface{}) {
	DBUG(200)
	DBUG(get([]interface{}{
		world,
		undocumented,
		0,
	}, 2))
	[]interface{}{
		world,
		undocumented,
		0,
	}
}

// Indices into mem
const (
	iworld int = iota
	iundocumented
	iturn
)

// Indices into mem[iworld]
const (
	imap int = iota
	ilambdaman
	ighosts
	ifruits
)

// Indices into mem[iworld][ilambdaman]
const (
	// Lambda-Man's vitality is a number which is a countdown to the
	// expiry of the active power pill, if any. It is 0 when no
	// power pill is active.
	// * 0: standard mode;
	// * n > 0: power pill mode: the number of game ticks remaining
	// while the power pill will will be active
	ivitality int = iota
	// Lambda-Man's current location, as an (x,y) pair.
	ilocation
	idirection
	ilives
	iscore
)

// Indices into mem[iworld][ighosts][:][ivitality]
const (
	standard int = iota
	fright
	invisible
)

func NextMem(mem, world interface{}) {
	[]interface{}{
		world,
		get(mem, iundocumented),
		get(mem, iturn) + 1,
	}
}

func Direction(mem interface{}) {
	ScoreFour(
		mem,
		get(get(get(mem, iworld), ilambdaman), ilocation))
}

func ScoreFour(mem, location interface{}) {
	IndexOfGreatestIn4([]interface{}{
		0,
		ScoreAt(mem, get(location, 0)+1, get(location, 1)),
		0,
		ScoreAt(mem, get(location, 0)-1, get(location, 1)),
	})
}

func ScoreAt(mem, x, y interface{}) {
	DBUG(x)
	DBUG(y)
	get(get(get(get(mem, iworld), imap), y), x) == 2
}

func IndexOfGreatestIn4(lst []interface{}) {
	if (get(lst, 0) > get(lst, 1)) * (get(lst, 0) > get(lst, 2)) * (get(lst, 0) > get(lst, 3)) {
		0
	}
	if (get(lst, 0) < get(lst, 1)) * (get(lst, 1) > get(lst, 2)) * (get(lst, 1) > get(lst, 3)) {
		1
	}
	if (get(lst, 0) < get(lst, 2)) * (get(lst, 1) < get(lst, 2)) * (get(lst, 2) > get(lst, 3)) {
		2
	}
	3
}

func get(lst, index interface{}) {
	if ATOM(lst) {
		if index == 0 {
			DBUG(100)
			lst
		} else {
			DBUG(101)
			BRK()
		}
	} else if index == 0 {
		DBUG(102)
		CAR(lst)
	} else {
		DBUG(103)
		get(CDR(lst), index-1)
	}
}

func length(lst interface{}) {
	lengthPlus(lst, 0)
}

func lengthPlus(lst, plus interface{}) {
	if ATOM(lst) {
		plus
	} else {
		lengthPlus(CDR(lst), plus+1)
	}
}
