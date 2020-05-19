package github

type Command interface {
	Run(gh *GitHub, owner string, repo string, number int, args ...interface{}) error
	Validate(args ...interface{}) error
}
