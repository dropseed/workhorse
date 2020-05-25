package config

type Config interface {
	Validate() error
	GetTargets() ([]string, error)
	GetTargetsMarkdown() ([]string, error)
	ExecuteTargets([]string) error
}
