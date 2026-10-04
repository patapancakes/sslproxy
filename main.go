package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/patapancakes/sslspoof"
	"go.yaml.in/yaml/v3"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("sslproxy by Pancakes (pancakes@mooglepowered.com)")
	fmt.Println()

	hostsPath := flag.String("hosts", "hosts.yml", "path to hosts file")
	flag.Parse()

	f, err := os.Open(*hostsPath)
	if err != nil {
		panic(err)
	}
	var hosts []struct{ Host, Listen, Backend string }
	err = yaml.NewDecoder(f).Decode(&hosts)
	if err != nil {
		panic(err)
	}
	f.Close()

	var eg errgroup.Group
	for _, host := range hosts {
		fmt.Printf("[%s] %s -> %s\n", host.Host, host.Listen, host.Backend)

		eg.Go(func() error {
			l, err := sslspoof.NewListener(host.Listen, host.Host, false)
			if err != nil {
				return err
			}

			defer l.Close()

			for {
				client, err := l.Accept()
				if err != nil {
					return err
				}

				backend, err := net.Dial("tcp", host.Backend)
				if err != nil {
					return err
				}

				go func() {
					go io.Copy(client, backend)
					io.Copy(backend, client)

					client.Close()
					backend.Close()
				}()
			}
		})
	}

	fmt.Println()
	fmt.Println("Now taking connections...")

	err = eg.Wait()
	if err != nil {
		panic(err)
	}
}
