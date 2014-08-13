package debug

import (
	"fmt"
	"github.com/zenithar/aktarus/state"
	"html/template"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
)

var memTemplate string = `<html><head><title>Memstats</title></head><body>
<h1>Memstats</h1>
<h2>General</h2>
<dl>
<dt>Alloc</dt>
<dd>{{.Alloc}}</dd>
<dt>TotalAlloc</dt>
<dd>{{.TotalAlloc}}</dd>
<dt>Sys</dt>
<dd>{{.Sys}}</dd>
<dt>Lookups</dt>
<dd>{{.Lookups}}</dd>
<dt>Mallocs</dt>
<dd>{{.Mallocs}}</dd>
<dt>Frees</dt>
<dd>{{.Frees}}</dd>
</dl>
<h2>Heap</h2>
<dl>
<dt>HeapAlloc</dt>
<dd>{{.HeapAlloc}}</dd>
<dt>HeapSys</dt>
<dd>{{.HeapSys}}</dd>
<dt>HeapIdle</dt>
<dd>{{.HeapIdle}}</dd>
<dt>HeapInuse</dt>
<dd>{{.HeapInuse}}</dd>
<dt>HeapReleased</dt>
<dd>{{.HeapReleased}}</dd>
<dt>HeapObjects</dt>
<dd>{{.HeapObjects}}</dd>
</dl>
<h2>Low-level</h2>
<dl>
<dt>StackInuse</dt>
<dd>{{.StackInuse}}</dd>
<dt>StackSys</dt>
<dd>{{.StackSys}}</dd>
<dt>MSpanInuse</dt>
<dd>{{.MSpanInuse}}</dd>
<dt>MSpanSys</dt>
<dd>{{.MSpanSys}}</dd>
<dt>MCacheInuse</dt>
<dd>{{.MCacheInuse}}</dd>
<dt>MCacheSys</dt>
<dd>{{.MCacheSys}}</dd>
<dt>BuckHashSys</dt>
<dd>{{.BuckHashSys}}</dd>
<dt>GCSys</dt>
<dd>{{.GCSys}}</dd>
<dt>OtherSys</dt>
<dd>{{.OtherSys}}</dd>
</dl>
<h2>GC</h2>
<dl>
<dt>NextGC</dt>
<dd>{{.NextGC}}</dd>
<dt>LastGC</dt>
<dd>{{.LastGC}}</dd>
<dt>PauseTotalNs</dt>
<dd>{{.PauseTotalNs}}</dd>
<dt>NumGC</dt>
<dd>{{.NumGC}}</dd>
<dt>EnableGC</dt>
<dd>{{.EnableGC}}</dd>
<dt>DebugGC</dt>
<dd>{{.DebugGC}}</dd>
</dl>
</body></html>
`

var indexTemplate string = `
<html><head><title>Debug</title></head><body>
<ul>
	<li><a href="/debug/states">State tracking</a></li>
	<li><a href="/debug/memstats">Memstats</a></li>
	<li><a href="/debug/pprof">Profiling</a></li>
	<li><a href="/debug/stop">Stop debug server</a></li>
</ul>
</body></html>
`

var stateTemplate string = `
<html><head><title>State Debug</title></head><body>
<h1>Channels</h1>
<dl>
	{{range $key, $val := .}}
		<dt><h2>{{$key}} Nicks</h2></dt>
		{{range $nick, $priv := $val}}
		<dd>
			<strong>{{$nick}}</strong> - <em>
			{{if $priv.Owner}}Owner{{end}}
			{{if $priv.Admin}}Admin{{end}}
			{{if $priv.Op}}Op{{end}}
			{{if $priv.HalfOp}}HalfOp{{end}}
			{{if $priv.Voice}}Voice{{end}}</em>
		</dd>
		{{end}}
	{{end}}
</dl>
</body></html>
`

var botState *state.StateTracker

func SetState(s *state.StateTracker) {
	botState = s
}

func printRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	tmpl, _ := template.New("memstats").Parse(memTemplate)
	// Error checking elided
	tmpl.Execute(w, m)
}

func printState(w http.ResponseWriter, r *http.Request) {
	if botState != nil {
		stateMapping := make(map[string]map[string]*state.ChannelPrivileges)

		for _, channel := range botState.Channels() {
			c := botState.GetChannel(channel)
			if c != nil {
				stateMapping[channel] = c.Nicks
			}
		}

		tmpl, _ := template.New("states").Parse(stateTemplate)
		// Error checking elided
		tmpl.Execute(w, stateMapping)
	} else {
		fmt.Fprint(w, "No state available")
	}
}

func printDebugIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, indexTemplate)
}

var running bool
var port string
var listen net.Listener
var lock sync.Mutex
var logger *log.Logger
var waiter sync.WaitGroup
var status string

func init() {
	logger = log.New(os.Stdout, "[debug] ", log.LstdFlags)
	status = "shut down"
}

func PrintStack() {
	buf := make([]byte, 10240) // 10kb buffer
	runtime.Stack(buf, true)
	logger.Printf("\n\n\n\nStacktrace:\n%s\n\n\n\n", buf)
}

type conn struct {
	net.Conn
	once      sync.Once
	waitGroup *sync.WaitGroup
}

// Close closes the connection and notifies the listener that accepted it.
func (c *conn) Close() (err error) {
	err = c.Conn.Close()
	c.once.Do(c.waitGroup.Done)
	return
}

type listener struct {
	net.Listener
	waitGroup *sync.WaitGroup
	conns     []*conn
}

// Accept waits for, accounts for, and returns the next connection to the
// listener.
func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if nil != err {
		return
	}
	l.waitGroup.Add(1)
	tmp := &conn{
		Conn:      c,
		waitGroup: l.waitGroup,
	}
	c = tmp
	l.conns = append(l.conns, tmp)
	return
}

// Close closes the listener.  It does not wait for all connections accepted
// through the listener to be closed.
func (l *listener) Close() (err error) {
	err = l.Listener.Close()
	logger.Printf("%#v", l)
	for _, c := range l.conns {
		c.Close()
	}
	return
}

func StartDebugServer() (string, bool) {
	lock.Lock()
	defer lock.Unlock()

	if running {
		return port, true
	}

	status = "starting"
	running = true
	l, _ := net.Listen("tcp", ":0")
	port = l.Addr().String()
	logger.Printf("Started debug server on port %s", port)

	listen = &listener{
		Listener:  l,
		waitGroup: &waiter,
	}

	go func() {
		http.HandleFunc("/debug/", printDebugIndex)
		http.HandleFunc("/debug/memstats/", printRuntimeInfo)
		http.HandleFunc("/debug/states/", printState)
		http.HandleFunc("/debug/stop/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			StopDebugServer()
		}))
		http.Serve(listen, nil)
	}()

	status = "started"
	return port, false
}

func DebugServerStatus() string {
	if port == "" {
		return status
	}
	return status + " (" + port + ")"
}

func StopDebugServer() {
	lock.Lock()
	defer lock.Unlock()

	if status != "started" {
		return
	}

	logger.Println("Shutting down debug server")
	status = "shutting down"
	if listen != nil {
		listen.Close()
	}
	waiter.Wait()
	logger.Println("Shut down debug server")
	status = "shut down"
	listen = nil
	port = ""
	running = false
}
