package main

import (
	"log"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/radhus/xenmatchbox/expander"
	"github.com/radhus/xenmatchbox/fetcher"
	"github.com/radhus/xenmatchbox/xen"
)

var config struct {
	output *string
	dir    *string
	ca     *string
	crt    *string
	key    *string

	server   *string
	httpPort *int
	grpcPort *int

	lookup *map[string]string
}

func main() {
	config.output = kingpin.Flag("output", "Output stream").Default("").String()
	config.dir = kingpin.Flag("output-directory", "Output directory").Required().String()
	config.crt = kingpin.Flag("cert", "Path to client certificate").Default("client.crt").String()
	config.key = kingpin.Flag("key", "Path to client key").Default("client.key").String()
	config.ca = kingpin.Flag("ca", "Path to CA certificate").Default("ca.crt").String()

	config.server = kingpin.Flag("server", "Matchbox server hostname/address").Required().String()
	config.httpPort = kingpin.Flag("httpport", "Matchbox server HTTP port").Default("8080").Int()
	config.grpcPort = kingpin.Flag("grpcport", "Matchbox server gRPC port").Default("8081").Int()

	config.lookup = kingpin.Flag("lookup", "Key=Value to be used in lookup, e.g. mac=00:01:02:03:04:05").Required().StringMap()

	output := os.Stdout

	var (
		_ = kingpin.Flag("output-format", "Output format").Default("simple0").Enum("simple0")
		_ = kingpin.Flag("kernel", "(unused) Kernel in configuration").String()
		_ = kingpin.Flag("ramdisk", "(unused) Ramdisk in configuration").String()
		_ = kingpin.Flag("args", "(unused) Kernel args in configuration").String()
		_ = kingpin.Arg("disk", "(unused) Disk for bootloader").String()
	)

	kingpin.Parse()

	if *config.output != "" {
		var err error
		if output, err = os.Create(*config.output); err != nil {
			log.Fatal("Couldn't open log output destination:", err)
		}
		defer output.Close()
		log.SetOutput(os.Stderr)
	}

	f, err := fetcher.New(
		*config.dir,
		*config.crt, *config.key, *config.ca,
		*config.server,
		*config.httpPort, *config.grpcPort,
	)
	if err != nil {
		log.Fatal("Couldn't create fetcher:", err)
	}

	fetched, err := f.Fetch(*config.lookup)
	if err != nil {
		log.Fatal("Couldn't fetch information:", err)
	}

	fetched.Args = expander.ExpandArguments(
		fetched.Args,
		*config.lookup,
	)

	simple0 := xen.OutputSimple0(fetched)
	output.Write(simple0)
}
