package swarm

type nullSwarmer struct{}

func (s *nullSwarmer) SwitchNode(addr string) error               { return nil }
func (s *nullSwarmer) GetInfo() (NodeInfo, error)                 { return NodeInfo{}, nil }
func (s *nullSwarmer) GetManagers() ([]NodeInfo, error)           { return nil, nil }
func (s *nullSwarmer) GetNodes() ([]NodeStatus, error)            { return nil, nil }
func (s *nullSwarmer) CreateSwarm(vms VMNodes) error              { return nil }
func (s *nullSwarmer) JoinToken(tokenType string) (string, error) { return "", nil }
