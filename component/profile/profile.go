package profile

import (
	"github.com/metacubex/clash/common/atomic"
)

// StoreSelected is a global switch for storing selected proxy to cache
var StoreSelected = atomic.NewBool(true)
