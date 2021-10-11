package agent

type GlobalInternetComputer interface {
	Agent() Agent
}

var (
	window GlobalInternetComputer
	global GlobalInternetComputer
	self   GlobalInternetComputer
)

func GetDefaultAgent() Agent {
	if self != nil {
		return self.Agent()
	} else if global != nil {
		return global.Agent()
	} else if window != nil {
		return window.Agent()
	}
	panic("No Agent could be found.');")
}
