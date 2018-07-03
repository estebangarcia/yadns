package main

import (
	"github.com/miekg/dns"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"net"
	"strings"
)

const root = "198.41.0.4:53"


func main() {
	go serve(":5454", "udp")
	dns.HandleFunc(".", handleRequestRoot)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}

func serve(addr string, net string) {
	server := &dns.Server{Addr: addr, Net: net}
	if err := server.ListenAndServe(); err != nil {
		fmt.Errorf("Error trying to listen on %s", addr)
	}
}

func handleRequestRoot(w dns.ResponseWriter, r *dns.Msg) {
	fmt.Println(r.Question)
	m := resolve(r, root)
	if m != nil {
		w.WriteMsg(m)
	}

}

func handleRequestAmazonaws(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	fmt.Println(r.Question)
	m.SetReply(r)

	ip := net.ParseIP("127.0.0.1")

	m.Answer = append(m.Answer, &dns.A{
		Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
		A:   ip,
	})

	//fmt.Println(m.String())

	w.WriteMsg(m)
}

func resolve(m *dns.Msg, addr string) (*dns.Msg) {
	fmt.Printf("going to query for: %s to addr %s\n", m.Question[0].Name, addr)

	c := new(dns.Client)
	in, _, err := c.Exchange(m, addr)
	if err != nil {
		fmt.Printf("Err: %s", err)
		return nil
	}

	fmt.Println(in)
	fmt.Println("")

	if len(in.Answer) == 0 {
		next := ""
		if len(in.Ns) > 0 && in.Ns[0].Header().Rrtype != 6 {
			splitted := strings.Split(in.Ns[0].String(), "\t")

			fmt.Println(splitted)

			toQuery := splitted[len(splitted)-1]

			if toQuery[len(toQuery)-1] == '.' {
				next = toQuery[:len(toQuery)-1]
			}

		} else {
			return in
		}

		return resolve(m, next + ":53")
	}

	return in
}