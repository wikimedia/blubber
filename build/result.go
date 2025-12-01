package build

// Result contains a built [Target] and its dependency [Target]s.
type Result struct {
	Target       *Target
	Dependencies []*Target
}
