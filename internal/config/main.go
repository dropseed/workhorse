package config

type Config interface {
	Validate() error
	GetTargets() ([]string, error)
	ExecuteTargets([]string) error
}
