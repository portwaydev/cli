package validate

var checkRegistry = CheckRegistry{
	Checks: make(map[string]ValidationCheck),
}

type CheckRegistry struct {
	Checks map[string]ValidationCheck
}

func (r *CheckRegistry) Register(check ValidationCheck) {
	r.Checks[check.Code] = check
}

func (r *CheckRegistry) Get(code string) ValidationCheck {
	return r.Checks[code]
}

func (r *CheckRegistry) GetAll() map[string]ValidationCheck {
	return r.Checks
}
