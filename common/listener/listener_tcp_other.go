//go:build !darwin

package listener

import "github.com/sagernet/sing/common/control"

func shareUIDControl() control.Func {
	return nil
}
