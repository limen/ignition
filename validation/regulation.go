package validation

import (
	"fmt"
	"github.com/asaskevich/govalidator"
)

var (
	Username = username{}
	Password = password{}
)

type username struct{}
type password struct{}

func (username) Match(v interface{}) error {
	vv, ok := v.(string)
	if !ok {
		return fmt.Errorf("username should be a string")
	}
	if len(vv) >= 3 && len(vv) <= 30 {
		return nil
	}

	return fmt.Errorf("username should contains 3-30 characters")
}

func (password) Match(v interface{}) error {
	vv, ok := v.(string)
	if !ok {
		return fmt.Errorf("password should be a string")
	}

	if len(vv) >= 6 && !govalidator.HasWhitespace(vv) {
		return nil
	}

	return fmt.Errorf("password should contains at least 6 characters")
}
