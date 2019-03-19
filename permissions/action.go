package permissions

// Action represents an action type
type Action uint8

// Can will return if an action can peform an action request
func (a Action) Can(ar Action) (can bool) {
	return a&ar != 0
}

const (
	// ActionNone represents a zero value, no action
	ActionNone Action = 1 << iota
	// ActionRead represents a reading action
	ActionRead
	// ActionWrite represents a writing action
	ActionWrite
	// ActionDelete represents a deleting action
	ActionDelete
)
