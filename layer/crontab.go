package layer

type ICrontab interface {
	IFlow
	Run() error
}

type Crontab struct {
	Flow
}

func (entity *Crontab) Run(args []string) error {
	return nil
}
