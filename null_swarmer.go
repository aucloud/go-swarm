package swarm

type nullSwarmer struct {
	*baseSwarmer
}

func (s *nullSwarmer) SwitchNode(addr string) error { return nil }
