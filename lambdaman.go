package lambdaman

func main() int {
	return arithmetic(1, 2, 3)
}

func arithmetic(a, b, c int) int {
	return addtwo(a, b) * c
}

func addtwo(a, b int) int {
	return a + b
}
