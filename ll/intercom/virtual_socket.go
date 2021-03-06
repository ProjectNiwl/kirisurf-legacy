package intercom

import (
	"errors"
	"io"
	"kirisurf-legacy/ll/common"
)

type virtsock struct {
	reader *common.BufferedPipe
	writer *common.BufferedPipe
}

func (xaxa *virtsock) Read(p []byte) (int, error) {
	return xaxa.reader.Read(p)
}

func (xaxa *virtsock) Write(p []byte) (int, error) {
	return xaxa.writer.Write(p)
}

func (xaxa *virtsock) Close() error {
	xaxa.writer.Close()
	xaxa.reader.Close()
	return nil
}

func (xaxa *virtsock) Flipped() *virtsock {
	toret := new(virtsock)
	toret.reader = xaxa.writer
	toret.writer = xaxa.reader
	return toret
}

func new_vs() *virtsock {
	return &virtsock{common.NewBufferedPipe(), common.NewBufferedPipe()}
}

type VirtualServer struct {
	vs_ch chan *virtsock
}

func (vs *VirtualServer) Accept() (io.ReadWriteCloser, error) {
	toret, ok := <-vs.vs_ch
	if !ok {
		return nil, errors.New("Channel closed")
	}
	return toret, nil
}

func VSConnect(tgt *VirtualServer) (rw io.ReadWriteCloser, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("Channel closed")
		}
	}()
	toret := new_vs()
	tosend := toret.Flipped()
	tgt.vs_ch <- tosend
	return toret, nil
}

func VSListen() *VirtualServer {
	return &VirtualServer{make(chan *virtsock)}
}
