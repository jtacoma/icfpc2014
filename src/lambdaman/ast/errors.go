package ast

type NotFoundError struct {
	BlockName string
	VarName   string
}

func (e NotFoundError) Error() string {
	return e.VarName + " not available in " + e.BlockName
}
