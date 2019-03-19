package permissions

func isValidActions(actions Action) bool {
	if actions > ActionRead+ActionWrite+ActionDelete {
		return false
	}

	return true
}
