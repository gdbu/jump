package permissions

// Groups represents a resource's group list (available actions keyed by group)
type Groups map[string]Action

func (g Groups) canRead(group string) bool {
	act := g[group]
	return act&ActionRead != 0
}

func (g Groups) canWrite(group string) bool {
	act := g[group]
	return act&ActionWrite != 0
}

func (g Groups) canDelete(group string) bool {
	act := g[group]
	return act&ActionDelete != 0
}

// Get will get the actions available to a given group
func (g Groups) Get(group string) (actions Action, ok bool) {
	actions, ok = g[group]
	return
}

// Has will return if a resource has a group
func (g Groups) Has(group string) (ok bool) {
	_, ok = g[group]
	return
}

// Can will check to see if a group can perform a given action
func (g Groups) Can(group string, action Action) (ok bool) {
	switch action {
	case ActionRead:
		return g.canRead(group)
	case ActionWrite:
		return g.canWrite(group)
	case ActionDelete:
		return g.canDelete(group)
	}

	return
}

// Set will set the actions available to a given group
func (g Groups) Set(group string, actions Action) (ok bool) {
	var currentActions Action
	currentActions, _ = g.Get(group)
	if currentActions|actions == currentActions {
		return false
	}

	g[group] = actions
	return true
}

// Remove will remove a group from a resource
func (g Groups) Remove(group string) (ok bool) {
	if _, ok = g.Get(group); ok {
		delete(g, group)
	}

	return
}
