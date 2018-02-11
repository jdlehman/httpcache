package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/satori/go.uuid"
)

var (
	proxyCache cache
	port       string
	origin     string
)

type cache struct {
	dat map[string]string
}

func newCache() cache {
	c := cache{make(map[string]string)}
	return c
}

func (c cache) exists(s string) bool {
	_, b := c.dat[s]
	return b
}

func (c cache) fname(s string) string {
	f, _ := c.dat[s]
	return f
}

func (c *cache) clean() {
	for _, fname := range c.dat {
		os.Remove(fname)
	}
}

func (c *cache) ingest(key string) error {
	fmt.Printf("ingesting %s\n", key)

	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	ext := filepath.Ext(u.EscapedPath())
	fname := fmt.Sprintf("%s%s", uuid.NewV4().String(), ext)

	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	url := fmt.Sprintf("%s/%s", origin, key)
	fmt.Printf("pulling %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	c.dat[key] = fname

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%v?%v", strings.Replace(r.URL.Path, "/", "", 1), r.URL.RawQuery)
	exists := proxyCache.exists(key)
	if exists == false {
		err := proxyCache.ingest(key)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	http.ServeFile(w, r, proxyCache.fname(key))
}

func init() {
	proxyCache = newCache()
	flag.StringVar(&port, "port", "8080", "port that the caching layer runs on")
	flag.StringVar(&origin, "origin", "http://localhost:3000", "origin to pull cache misses from")
	flag.Parse()

	// cleanup files on exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		proxyCache.clean()
		os.Exit(0)
	}()
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(handler))

	log.Printf("Caching requests to port %s. Cache misses will pull from %s...\n", port, origin)
	http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
