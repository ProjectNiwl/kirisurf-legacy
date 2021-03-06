package kiss

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/binary"

	"code.google.com/p/go.crypto/blowfish"
)

// This file implements the Grand Central Chugger, which handles stream authentication
// Stream cipher is used.

type chugger struct {
	streamer  cipher.Stream
	key       []byte
	read_num  uint64
	write_num uint64
}

func (ctx *chugger) Seal(pt []byte) []byte {
	seq := make([]byte, 8)
	binary.LittleEndian.PutUint64(seq, ctx.write_num)
	ctx.write_num++

	toret := make([]byte, 20+len(pt))

	xaxa := hmac.New(sha1.New, ctx.key)
	xaxa.Write(pt)
	xaxa.Write(seq)
	tag := xaxa.Sum(nil)
	pt = append(tag, pt...)

	ctx.streamer.XORKeyStream(toret, pt)
	return toret
}

func (ctx *chugger) Open(ct []byte) ([]byte, error) {
	if len(ct) < 20 {
		return nil, ErrPacketTooShort
	}
	seq := make([]byte, 8)
	binary.LittleEndian.PutUint64(seq, ctx.read_num)
	ctx.read_num++

	pt := make([]byte, len(ct))
	ctx.streamer.XORKeyStream(pt, ct)
	xaxa := hmac.New(sha1.New, ctx.key)
	xaxa.Write(pt[20:])
	xaxa.Write(seq)
	actual_sum := xaxa.Sum(nil)
	hypo_sum := pt[:20]

	if subtle.ConstantTimeCompare(actual_sum, hypo_sum) == 1 {
		return pt[20:], nil
	}
	return nil, ErrMacNoMatch
}

func make_chugger(key []byte) *chugger {
	state, _ := blowfish.NewCipher(key)
	streamer := cipher.NewCTR(state, make([]byte, state.BlockSize()))
	return &chugger{streamer, key, 0, 0}
}
