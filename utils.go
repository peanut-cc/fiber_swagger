package fiberx

import (
	"os"
	"path/filepath"
	"reflect"
)

func ensureDirectory(filePath string) error {
	fileAbsDir, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	_, err = os.Stat(fileAbsDir)
	if err != nil {
		err := os.Mkdir(fileAbsDir, 0766)
		if err != nil {
			return err
		}
	}
	return nil
}

func isBasicType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true
	default:
		return false
	}
}
