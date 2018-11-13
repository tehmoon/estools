package flags

import (
	"log"
	"github.com/spf13/pflag"
	"github.com/tehmoon/errors"
	"time"
	"net"
	"os"
	"fmt"
	"net/url"
)

type Flags struct {
	Server string
	Owners []string
	Dir string
	Exec string
	QueryDelay time.Duration
	PublicURL string
	Listen string
	Index string
}

func GetOutboundIP() (outbound string, err error){
	conn, err := net.Dial("udp", "255.255.255.255:1")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if ! ok {
		return "", errors.New("Error asserting net.Dial to *net.UDPAddr")
	}

	return localAddr.IP.String(), nil
}

func Parse() (flags *Flags, err error) {
	flags = &Flags{}

	outbound, err := GetOutboundIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the default ip address from this OS")
	}

	listenURL := &url.URL{
		Host: ":7769",
	}

	publicURL := &url.URL{
		Scheme: "http",
		Host: fmt.Sprintf("%s:%s", outbound, listenURL.Port()),
	}

	pflag.StringVar(&flags.Server, "server", "http://localhost:9200", "Specify elasticsearch server to query")
	pflag.StringVar(&flags.Index, "index", "", "Specify the elasticsearch index to query")
	pflag.StringVar(&flags.Dir, "dir", "", "Directory where the .json files are")
	pflag.StringVar(&flags.Exec, "exec", "", "Execute a command when alerting")
	pflag.StringVar(&flags.Listen, "listen", listenURL.Host, "Start HTTP server and listen in ip:port")
	pflag.StringArrayVar(&flags.Owners, "owners", make([]string, 0), "List of default owners to notify")
	pflag.StringVar(&flags.PublicURL, "public-url", publicURL.String(), "Public facing URL")
	pflag.DurationVar(&flags.QueryDelay, "query-delay", time.Second, "When using \"now\", delay the query to allow index time")

	pflag.Parse()

	if flags.Index == "" {
		return nil, errors.Wrap(ErrFlagRequired, "index")
	}

	if flags.Dir == "" {
		return nil, errors.Wrap(ErrFlagRequired, "dir")
	}

	flags.PublicURL, err = derivePublicURL(flags.PublicURL, flags.Listen)
	if err != nil {
		return nil, err
	}

	log.Println(flags.PublicURL)

	if flags.QueryDelay <= time.Duration(0) {
		return nil, errors.Errorf("Flag %q must be higher than %q", "query-delay", time.Duration(0).String())
	}

	err = isDir(flags.Dir)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to assert flag %q", "dir")
	}

	return flags, nil
}

func isDir(p string) (error) {
	fm, err := os.Stat(p)
	if err != nil {
		return err
	}

	if ! fm.IsDir() {
		return errors.Errorf("Path %q is not a directory", p)
	}

	return nil
}

var ErrFlagRequired = errors.New("Missing required flag")
