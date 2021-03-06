package circuitry

import (
	"crypto/subtle"
	"encoding/base32"
	"fmt"
	"io"
	"kirisurf-legacy/ll/dirclient"
	"kirisurf-legacy/ll/intercom"
	"kirisurf-legacy/ll/kiss"
	"strconv"
	"strings"

	"github.com/KirisurfProject/kilog"
)

func Old2New(addr string) string {
	port, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	naddr := fmt.Sprintf("kirisurf-legacy@%s:%d", strings.Split(addr, ":")[0], port+1)
	return naddr
}

var dialer = new(intercom.IntercomDialer)

var Dialer = dialer

func hash_base32(data []byte) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		kiss.KeyedHash(data, data)[:20]))
}

func BuildCircuit(slc []dirclient.KNode, subchannel int) (io.ReadWriteCloser, error) {
	// this returns a checker whether a public key is valid
	pubkey_checker := func(hsh string) func([]byte) bool {
		return func([]byte) bool { return true }

		return func(xaxa []byte) bool {
			hashed := hash_base32(xaxa)
			return subtle.ConstantTimeCompare([]byte(hashed), []byte(hsh)) == 1
		}
	}

	// circuit-building loop
	gwire, err := dialer.Dial(Old2New(slc[0].Address))
	if err != nil {
		return nil, err
	}
	wire, err := kiss.TransportHandshake(kiss.GenerateDHKeys(),
		gwire, pubkey_checker(slc[0].PublicKey))
	if err != nil {
		gwire.Close()
		return nil, err
	}
	for _, ele := range slc[1:] {
		kilog.Debug("Connecting to node %v...", string([]byte(ele.PublicKey)[:10]))
		// extend wire
		_, err = wire.Write(append([]byte{byte(len(ele.PublicKey))}, ele.PublicKey...))
		if err != nil {
			gwire.Close()
			return nil, err
		}

		verifier := pubkey_checker(ele.PublicKey)
		// at this point wire is raw (well unobfs) connection to next
		wire, err = kiss.TransportHandshake(kiss.GenerateDHKeys(), wire, verifier)
		if err != nil {
			kilog.Debug("Died when transport at %s", ele.PublicKey)
			gwire.Close()
			return nil, err
		}
		kilog.Debug("Connected to node %v!", string([]byte(ele.PublicKey)[:10]))
	}
	_, err = wire.Write([]byte{byte(subchannel)})
	if err != nil {
		gwire.Close()
		return nil, err
	}
	kilog.Debug("Opened subchannel %d", subchannel)
	return wire, nil
}

func init() {
	*dialer = *intercom.MakeIntercomDialer()
}
