package check

import "fmt"

type CheckerImpl struct {
}

type IChecker interface {
	CheckMethod(b []byte) bool
	CheckAuth(b []byte) (bool, uint8)
}

func (c *CheckerImpl) CheckMethod(b []byte) bool {
	if b[0] != 0x05 || (b[1] == 0x01 && b[2] != 0x02) {
		return false
	}
	return true
}

func (c *CheckerImpl) CheckAuth(b []byte) (bool, uint8) {
	b0 := b[0]

	nameLens := int(b[1])
	name := string(b[2 : 2+nameLens])

	passLens := int(b[2+nameLens])
	pass := string(b[2+nameLens+1 : 2+nameLens+1+passLens])

	fmt.Println(name, pass)
	if name != "abc" || pass != "123" {
		return false, b0
	} else {
		return true, b0
	}
}
