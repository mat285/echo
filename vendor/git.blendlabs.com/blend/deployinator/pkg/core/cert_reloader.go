package core

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"sync"
	"time"

	exception "github.com/blend/go-sdk/exception"
	logger "github.com/blend/go-sdk/logger"
	"github.com/fsnotify/fsnotify"
)

// based on this StackOverflow answer https://stackoverflow.com/a/40883377

const (
	defaultCertReloadDelay = time.Second
)

// CertReloaderState is the state of CertReloader
type CertReloaderState int

const (
	// CertReloaderStateInitialized is the initial state
	CertReloaderStateInitialized CertReloaderState = iota
	// CertReloaderStateRunning is the running state, after the watcher is added
	CertReloaderStateRunning
	// CertReloaderStateStopped is the stopped state
	CertReloaderStateStopped
)

// FileWatcherInterface is an interface for fsnotify.Watcher
type FileWatcherInterface interface {
	Add(name string) error
	Remove(name string) error
	Close() error
	EventsChan() <-chan fsnotify.Event
	ErrorsChan() <-chan error
}

type watcher struct {
	*fsnotify.Watcher
}

func (w *watcher) EventsChan() <-chan fsnotify.Event {
	return w.Events
}

func (w *watcher) ErrorsChan() <-chan error {
	return w.Errors
}

// CertReloader reloads a cert key pair when there is a change, e.g. cert renewal
type CertReloader struct {
	log         *logger.Logger
	mutex       sync.RWMutex
	cert        *tls.Certificate
	certPath    string
	keyPath     string
	reloadDelay time.Duration
	reloadTimer *time.Timer
	watcher     FileWatcherInterface
	state       CertReloaderState
	stop        chan interface{}
}

// NewCertReloader creates a new CertReloader object with a reload delay
func NewCertReloader(log *logger.Logger, certPath, keyPath string) (*CertReloader, error) {
	result := &CertReloader{
		log:         log,
		certPath:    certPath,
		keyPath:     keyPath,
		reloadDelay: defaultCertReloadDelay,
		state:       CertReloaderStateInitialized,
		stop:        make(chan interface{}),
	}

	// load cert to make sure the current key pair is valid
	if err := result.reload(); err != nil {
		return nil, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, exception.New(err)
	}
	result.watcher = &watcher{w}
	return result, nil
}

func (cr *CertReloader) scheduleReload() {
	if cr.reloadTimer != nil {
		cr.reloadTimer.Stop()
	}
	cr.reloadTimer = time.AfterFunc(cr.reloadDelay, func() {
		cr.log.Infof("cert reloading")
		if err := cr.reload(); err != nil {
			cr.log.Warningf("cannot reload certificate: %v", err)
		} else {
			cr.log.Infof("cert reloaded")
		}

		// the watch of the removed file is automatically removed (at least for linux)
		// we need to ensure that we are still watching the file. `watcher.Add` is idempotent
		if err := cr.watcher.Add(cr.certPath); err != nil {
			cr.log.Errorf("cannot add watch for %s: %v", cr.certPath, err)
		} else {
			cr.log.Debugf("fsnotify: new watch set up")
		}
	})
}

func (cr *CertReloader) reload() error {
	cert, err := tls.LoadX509KeyPair(cr.certPath, cr.keyPath)
	if err != nil {
		return exception.New(err)
	}
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	cr.cert = &cert
	return nil
}

// GetCertificate gets the cached certificate, it blocks when the `cert` field is being updated
func (cr *CertReloader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	return cr.cert, nil
}

// State returns the current cert reloader state
func (cr *CertReloader) State() CertReloaderState {
	return cr.state
}

// Run watches the cert and triggers a reload on change
func (cr *CertReloader) Run() error {
	// we want to add the watcher and start polling for event right away (vs adding it in the constructor)
	if err := cr.watcher.Add(cr.certPath); err != nil {
		return exception.New(err)
	}
	defer cr.watcher.Remove(cr.certPath)
	cr.log.Infof("watching cert at %s", cr.certPath)

	cr.state = CertReloaderStateRunning
	defer func() {
		cr.state = CertReloaderStateStopped
	}()

	for {
		select {
		case event, ok := <-cr.watcher.EventsChan():
			if !ok {
				return nil
			}
			cr.log.Debugf("fsnotify event: %v", event)

			// note: behavior observed when kube updates mounted secret is chmod, remove, then the watch is lost
			// the behavior seems to vary across environment (e.g. kube, jenkins, mac)
			// if this causes any more issue, we should consider just doing poll + stat
			modified := fsnotify.Write | fsnotify.Remove | fsnotify.Chmod
			if event.Op&modified > 0 {
				// since we are watching only the cert, we may need to wait a bit for the key to get updated.
				cr.log.Infof("cert modified: %s, scheduling a reload after %s", event.Name, cr.reloadDelay)
				cr.scheduleReload()
			}
		case err, ok := <-cr.watcher.ErrorsChan():
			if !ok {
				return nil
			}
			cr.log.Errorf("fsnotify error: %v", err)
		case <-cr.stop:
			cr.log.Infof("stopped watching cert at %s", cr.certPath)
			return nil
		}
	}
}

// Stop stops watching the cert. don't call this more than once.
func (cr *CertReloader) Stop() error {
	close(cr.stop)
	return exception.New(cr.watcher.Close())
}

//NameToCertReloader does the same thing as what you would find in crypto/tls but for CertReloaders
type NameToCertReloader map[string]*CertReloader

//CertTuple couples a certificate with the files it is derived from
type CertTuple struct {
	CertPath    string
	KeyPath     string
	Certificate *tls.Certificate
}

//BuildNameToCertReloader does the same thing as crypto/tls BuildNameToCertificate but for cert_reloaders
func BuildNameToCertReloader(log *logger.Logger, certificates []*CertTuple) NameToCertReloader {
	mapping := make(NameToCertReloader)
	for _, meta := range certificates {
		x509Cert := meta.Certificate.Leaf
		if x509Cert == nil {
			var err error
			x509Cert, err = x509.ParseCertificate(meta.Certificate.Certificate[0])
			if err != nil {
				continue
			}
		}

		watcher, err := NewCertReloader(log, meta.CertPath, meta.KeyPath)

		if err != nil {
			log.SyncFatalExit(err)
		}
		if len(x509Cert.Subject.CommonName) > 0 {
			mapping[x509Cert.Subject.CommonName] = watcher
		}
		for _, san := range x509Cert.DNSNames {
			mapping[san] = watcher
		}
	}
	return mapping
}

//GetCertReloader retrieves the correct reloader given a client request
func GetCertReloader(clientHello *tls.ClientHelloInfo, certReloaderMapping NameToCertReloader) (*CertReloader, error) {
	if len(certReloaderMapping) == 0 {
		return nil, exception.New("No certificate reloaders were set")
	}

	name := strings.ToLower(clientHello.ServerName)
	for len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	if watcher, ok := certReloaderMapping[name]; ok {
		return watcher, nil
	}

	// could be a wildcard
	if watcher, ok := certReloaderMapping[WildcardNeededForDomain(name)]; ok {
		return watcher, nil
	}

	// Nothing was found
	return nil, exception.New("No valid certificate was found")
}
