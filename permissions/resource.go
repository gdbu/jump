package permissions

// Resource represents an entry resource (available actions keyed by group)
type Resource map[string]Action

func (r Resource) canRead(group string) bool {
	act := r[group]
	return act&ActionRead != 0
}

func (r Resource) canWrite(group string) bool {
	act := r[group]
	return act&ActionWrite != 0
}

func (r Resource) canDelete(group string) bool {
	act := r[group]
	return act&ActionDelete != 0
}

// Get will get the actions available to a given group
func (r Resource) Get(group string) (actions Action, ok bool) {
	actions, ok = r[group]
	return
}

// Has will return if a resource has a group
func (r Resource) Has(group string) (ok bool) {
	_, ok = r[group]
	return
}

// Can will check to see if a group can perform a given action
func (r Resource) Can(group string, action Action) (ok bool) {
	switch action {
	case ActionRead:
		return r.canRead(group)
	case ActionWrite:
		return r.canWrite(group)
	case ActionDelete:
		return r.canDelete(group)
	}

	return
}

// Set will set the actions available to a given group
func (r Resource) Set(group string, actions Action) (ok bool) {
	var currentActions Action
	currentActions, _ = r.Get(group)
	if currentActions|actions == currentActions {
		return false
	}

	r[group] = actions
	return true
}

// Remove will remove a group from a resource
func (r Resource) Remove(group string) (ok bool) {
	if _, ok = r.Get(group); ok {
		delete(r, group)
	}

	return
}
