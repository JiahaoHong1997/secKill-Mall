package loadBalance

type LbType int

type LoadBalance interface {
	Add(...string) error
	Get(string) (string, error)
}

const (
	LbRandom LbType = iota
	LbRound
	LbWeightRound
	LbConsistent
)

func LoadBalanceFactory(lbType LbType, conVirtualNode int) LoadBalance {
	switch lbType {
	case LbRandom:
		return NewRandomBalance()
	case LbRound:
		return NewRoundBalance()
	case LbWeightRound:
		return NewWeightRound()
	case LbConsistent:
		return NewConsistent(conVirtualNode)
	default:
		return NewConsistent(conVirtualNode)
	}
}
