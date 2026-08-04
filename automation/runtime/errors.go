package runtime

import "fmt"

func errMissing(name string) error {
	return fmt.Errorf("runtime: missing %s", name)
}
