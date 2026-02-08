package shopdomain

import (
	"strings"
)

type Name string

func NewName(name string) (Name, error) {
	n := Name(name)
	if err := n.validate(); err != nil {
		return Name(""), err
	}

	return n, nil
}

func (n Name) validate() error {
	if strings.TrimSpace(n.String()) == "" {
		return ErrEmptyShopName
	}

	return nil
}

func (n Name) String() string {
	return string(n)
}
