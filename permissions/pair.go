package permissions

// NewPair will return a new permissions pair
func NewPair(group string, actions Action) (p Pair) {
	p.Group = group
	p.Actions = actions
	return
}

// Pair represents a permissions pair
type Pair struct {
	Group   string
	Actions Action
}
