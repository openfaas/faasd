package pkg

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LocalResolver provides hostname to IP look-up for faasd core services
type LocalResolver struct {
	Path  string
	Map   map[string]string
	Mutex *sync.RWMutex
}

// NewLocalResolver creates a new resolver for reading from a hosts file
func NewLocalResolver(path string) Resolver {
	return &LocalResolver{
		Path:  path,
		Mutex: &sync.RWMutex{},
		Map:   make(map[string]string),
	}
}

// Start polling the disk for the hosts file in Path
func (l *LocalResolver) Start() {
	var lastStat os.FileInfo

	for {
		rebuild := false
		if info, err := os.Stat(l.Path); err == nil {
			if lastStat == nil {
				rebuild = true
			} else {
				if !lastStat.ModTime().Equal(info.ModTime()) {
					rebuild = true
				}
			}
			lastStat = info
		}

		if rebuild {
			log.Printf("Resolver rebuilding map")
			l.rebuild()
		}
		time.Sleep(time.Second * 3)
	}
}

func (l *LocalResolver) rebuild() {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	fileData, fileErr := ioutil.ReadFile(l.Path)
	if fileErr != nil {
		log.Printf("resolver rebuild error: %s", fileErr.Error())
		return
	}

	lines := strings.Split(string(fileData), "\n")

	for _, line := range lines {
		index := strings.Index(line, "\t")

		if len(line) > 0 && index > -1 {
			ip := line[:index]
			host := line[index+1:]
			log.Printf("Resolver: %q=%q", host, ip)
			l.Map[host] = ip
		}
	}
}

// Get resolves a hostname to an IP, or timesout after the duration has passed
func (l *LocalResolver) Get(upstream string, got chan<- string, timeout time.Duration) {
	start := time.Now()
	for {
		if val := l.get(upstream); len(val) > 0 {
			got <- val
			break
		}

		if time.Now().After(start.Add(timeout)) {
			log.Printf("Timed out after %s getting host %q", timeout.String(), upstream)
			break
		}

		time.Sleep(time.Millisecond * 250)
	}
}

func (l *LocalResolver) get(upstream string) string {
	l.Mutex.RLock()
	defer l.Mutex.RUnlock()

	if val, ok := l.Map[upstream]; ok {
		return val
	}

	return ""
}
