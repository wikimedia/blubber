package build

// CacheDisabler returns whether caching for a named target should be disabled
// or not.
type CacheDisabler func(name string) bool
