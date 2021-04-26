package project

func (p *Project) docker() *Project {
	if p.Error() != nil {
		return p
	}

	return p
}
