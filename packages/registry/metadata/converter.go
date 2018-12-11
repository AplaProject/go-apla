package metadata

import (
	"reflect"

	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/pkg/errors"
)

type converter struct{}

func (m converter) createFromParams(name string, params map[string]interface{}) (types.RegistryModel, error) {
	r := model.GetRegistries()
	for _, registry := range r {
		if registry.ModelName() == name {
			filled, err := registry.CreateFromData(params)
			if err != nil {
				return nil, err
			}

			return filled, nil
		}
	}

	return nil, ErrWrongRegistry
}

func (m converter) updateFromParams(name string, value types.RegistryModel, params map[string]interface{}) error {
	t := reflect.ValueOf(value)
	if t.Kind() != reflect.Ptr || t.IsNil() {
		return errors.New("value must be a pointer")
	}

	r := model.GetRegistries()
	for _, registry := range r {
		if registry.ModelName() == name {
			err := registry.UpdateFromData(value, params)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return ErrWrongRegistry
}
