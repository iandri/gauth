package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/iandri/gauth/gauth"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	jsOutput = flag.Bool("j", false, "json output")
	withPass = flag.String("p", "", "password")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Output struct {
	Account  string `json:"account"`
	Prev     string `json:"prev"`
	Curr     string `json:"curr"`
	Next     string `json:"next"`
	Progress string `json:"progress"`
}

func main() {
	flag.Parse()
	cfgPath := os.Getenv("GAUTH_CONFIG")
	if cfgPath == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		cfgPath = path.Join(user.HomeDir, ".config/gauth.csv")
	}

	cfgContent, err := gauth.LoadConfigFile(cfgPath, getPassword)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	urls, err := gauth.ParseConfig(cfgContent)
	if err != nil {
		log.Fatalf("Decoding configuration file: %v", err)
	}

	_, progress := gauth.IndexNow() // TODO: do this per-code

	var out []Output
	for _, url := range urls {
		var output Output
		prev, curr, next, err := gauth.Codes(url)
		if err != nil {
			log.Fatalf("Generating codes for %q: %v", url.Account, err)
		}
		output.Account = url.Account
		output.Prev = prev
		output.Curr = curr
		output.Next = next
		output.Progress = fmt.Sprintf("%d/%d", progress+1, 30)
		out = append(out, output)
	}

	if *withPass != "" {
		var full string
		if progress <= 20 {
			full = fmt.Sprintf("%s%s", *withPass, out[0].Curr)
		} else {
			full = fmt.Sprintf("%s%s", *withPass, out[0].Next)
		}

		fmt.Println(full)
		return
	}
	if *jsOutput {
		js, err := json.Marshal(out)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(js))
		return
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
	fmt.Fprintln(tw, "\tprev\tcurr\tnext")
	for _, o := range out {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", o.Account, o.Prev, o.Curr, o.Next)
	}
	tw.Flush()
	fmt.Printf("[%-29s]\n", strings.Repeat("=", progress))
}

func getPassword() ([]byte, error) {
	fmt.Printf("Encryption password: ")
	defer fmt.Println()
	return terminal.ReadPassword(int(syscall.Stdin))
}
