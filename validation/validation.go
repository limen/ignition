package validation

type Regulator interface {
	Match(v interface{}) error
}

type Rules map[string]Regulator
type Requires map[string]bool

func (r Rules) Has(field interface{}) bool {
	if r == nil {
		return false
	}
	vv, ok := field.(string)
	if !ok {
		return false
	}
	_, has := r[vv]
	return has
}
