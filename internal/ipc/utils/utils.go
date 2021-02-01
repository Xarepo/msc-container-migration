package utils

import "fmt"

func GetSocketAddr(containerId string) string {
	return fmt.Sprintf("/tmp/msc_%s.sock", containerId)
}
