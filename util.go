package main

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, a_elt := range a {
		b_elt := b[i]
		if a_elt != b_elt {
			return false
		}
	}

	return true
}
