package fetcher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	client "github.com/coreos/matchbox/matchbox/client"
	serverpb "github.com/coreos/matchbox/matchbox/server/serverpb"
	storagepb "github.com/coreos/matchbox/matchbox/storage/storagepb"

	"github.com/radhus/xenmatchbox/machine"
)

type Fetcher struct {
	dir string

	conn *client.Client
	url  *url.URL
}

func tlsConfig(crtPath, keyPath, caPath string) (*tls.Config, error) {
	crt, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		return nil, err
	}

	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("Couldn't parse CA cert as PEM")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      pool,
		Certificates: []tls.Certificate{crt},
	}, nil
}

func New(
	directory string,
	crt, key, ca string,
	server string,
	httpPort, grpcPort int,
) (*Fetcher, error) {
	f := &Fetcher{
		dir: directory,
	}

	tlsConf, err := tlsConfig(crt, key, ca)
	if err != nil {
		return nil, err
	}

	f.conn, err = client.New(&client.Config{
		Endpoints: []string{net.JoinHostPort(server, strconv.Itoa(grpcPort))},
		TLS:       tlsConf,
	})
	if err != nil {
		return nil, err
	}

	f.url, err = url.Parse(
		fmt.Sprintf("http://%s", net.JoinHostPort(server, strconv.Itoa(httpPort))),
	)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f *Fetcher) fetchBoot(lookup map[string]string) (*storagepb.NetBoot, error) {
	response, err := f.conn.Select.SelectProfile(
		context.TODO(),
		&serverpb.SelectProfileRequest{
			Labels: lookup,
		},
	)
	if err != nil {
		return nil, err
	}

	profile := response.GetProfile()
	if profile == nil {
		return nil, errors.New("Got no profile")
	}

	boot := profile.GetBoot()
	if boot == nil {
		return nil, errors.New("Got no boot")
	}

	return boot, nil
}

func (f *Fetcher) download(src *url.URL) error {
	dest := path.Join(f.dir, path.Base(src.Path))
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(src.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (f *Fetcher) downloadAssets(kernel, initrd string) error {
	relativeKernel, err := url.Parse(kernel)
	if err != nil {
		return err
	}
	kernelUrl := f.url.ResolveReference(relativeKernel)

	relativeInitrd, err := url.Parse(initrd)
	if err != nil {
		return err
	}
	initrdUrl := f.url.ResolveReference(relativeInitrd)

	log.Println("Would download kernel from:", kernelUrl)
	if err = f.download(kernelUrl); err != nil {
		return err
	}

	log.Println("Would download initrd from:", initrdUrl)
	if err = f.download(initrdUrl); err != nil {
		return err
	}

	return nil
}

func (f *Fetcher) Fetch(lookup map[string]string) (*machine.Configuration, error) {
	boot, err := f.fetchBoot(lookup)
	if err != nil {
		return nil, err
	}

	kernel := boot.GetKernel()
	if kernel == "" {
		return nil, errors.New("Got no kernel")
	}

	initrds := boot.GetInitrd()
	if initrds == nil || len(initrds) < 1 {
		// TODO: not always an error right...
		return nil, errors.New("Got no initrd")
	}
	initrd := initrds[0]

	argsLines := boot.GetArgs()
	if argsLines == nil || len(argsLines) < 1 {
		// TODO: not always an error right...
		return nil, errors.New("Got no args")
	}
	args := strings.Join(argsLines, " ")

	err = f.downloadAssets(kernel, initrd)
	if err != nil {
		return nil, err
	}

	return &machine.Configuration{
		Kernel:  path.Join(f.dir, path.Base(kernel)),
		Ramdisk: path.Join(f.dir, path.Base(initrd)),
		Args:    args,
	}, nil
}
