package utils

import (
	"github.com/MaroIshiku/dyniku/internal/models"
	"github.com/MaroIshiku/dyniku/pkg/publicip/ipversion"
)

func ToString(domain, owner string, provider models.Provider, ipVersion ipversion.IPVersion) string {
	return "[domain: " + domain + " | owner: " + owner + " | provider: " +
		string(provider) + " | ip: " + ipVersion.String() + "]"
}
