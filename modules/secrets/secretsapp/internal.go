package secretsapp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/quara-dev/beyond/modules/secrets"
)

func (a *App) getStore(name string) secrets.Store {
	return a.stores[name]
}

func (a *App) getStoreAndKey(key string) (secrets.Store, string, error) {
	parts := strings.Split(key, "@")
	var storename string
	var secretkey string
	switch len(parts) {
	case 1:
		secretkey = parts[0]
		storename = a.defaultStore
	case 2:
		secretkey = parts[0]
		storename = parts[1]
	default:
		return nil, "", errors.New("invalid key")
	}
	store := a.getStore(storename)
	if store == nil {
		return nil, "", fmt.Errorf("store not found: %s", storename)
	}
	return store, secretkey, nil
}
