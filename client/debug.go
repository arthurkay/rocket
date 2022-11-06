//go:build !release
// +build !release

package client

var (
	rootCrtPaths = []string{"client/tls/rocketroot.crt", "client/tls/snakeoilca.crt"}
)

func useInsecureSkipVerify() bool {
	return true
}
