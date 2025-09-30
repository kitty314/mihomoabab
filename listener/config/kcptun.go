package config

import "github.com/metacubex/clash/transport/kcptun"

type KcpTun struct {
	Enable        bool `json:"enable"`
	kcptun.Config `json:",inline"`
}
