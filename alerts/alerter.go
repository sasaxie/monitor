package alerts

type Alerter interface {
	Load()
	Start()
	Alert()
}
