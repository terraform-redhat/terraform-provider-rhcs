package helper

import "fmt"

func GetRegistry(port int) string {
	return fmt.Sprintf("10.0.0.0:%d", port)
}

func GetRegistries(ports ...int) (regs []string) {
	for _, p := range ports {
		regs = append(regs, GetRegistry(p))
	}
	return
}
