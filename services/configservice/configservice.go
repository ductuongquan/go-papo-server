
package configservice

import (
	"crypto/ecdsa"
	"bitbucket.org/enesyteam/papo-server/model"
)

// An interface representing something that contains a Config, such as the app.App struct
type ConfigService interface {
	Config() *model.Config
	AddConfigListener(func(old, current *model.Config)) string
	RemoveConfigListener(string)
	AsymmetricSigningKey() *ecdsa.PrivateKey
}
