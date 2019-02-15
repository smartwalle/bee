package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/smartwalle/bee"
	"math/big"
	"time"
)

func main() {
	listener, err := bee.ListenQUIC(":8889", generateTLSConfig(), &quic.Config{IdleTimeout: time.Second * 60})
	if err != nil {
		return
	}

	var hub = bee.NewHub()
	var handler = &handler{h: hub}

	for {
		c, err := listener.Accept()
		if err != nil {
			return
		}

		bee.NewSession(c, c.RemoteAddr().String(), "tag", 1024, handler)
	}
}

type handler struct {
	h bee.Hub
}

func (this *handler) DidOpenSession(s bee.Session) {
	this.h.AddSession(s)
	fmt.Println("open session", s.Identifier(), s.Tag())
	fmt.Println(this.h.Len())
}

func (this *handler) DidClosedSession(s bee.Session, err error) {
	this.h.RemoveSession(s)
	fmt.Println("close session")
	fmt.Println(this.h.Len())
}

func (this *handler) DidWrittenData(s bee.Session, data []byte) {
	fmt.Println("write data", s.Identifier(), string(data))
}

func (this *handler) DidReceivedData(s bee.Session, data []byte) {
	fmt.Println("receive data", s.Identifier(), string(data))
	s.WriteMessage([]byte("success"))
	//var cl = this.h.GetAllSessions()
	//for _, c := range cl {
	//	fmt.Println(c.WriteMessage(data))
	//}
	//s.Write([]byte(fmt.Sprintf("%s", time.Now())))
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
