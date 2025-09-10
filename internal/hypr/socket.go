package hypr

import "fmt"

func GetHyprEventsSocket(xdgRuntimeDir, instanceSignature string) string {
	return fmt.Sprintf("%s/hypr/%s/.socket2.sock", xdgRuntimeDir, instanceSignature)
}
