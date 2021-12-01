package core

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	gopath "path"
	"strings"
	"sync"

	"github.com/aler9/rtsp-simple-server/internal/conf"
	"github.com/aler9/rtsp-simple-server/internal/hikka"
	"github.com/aler9/rtsp-simple-server/internal/logger"
	"github.com/gin-gonic/gin"
)

type hikkaMuxerRequest struct {
	Dir  string
	File string
	Req  *http.Request
	Res  chan hikkaMuxerResponse
}

type hikkaMuxerResponse struct {
	Status int
	Header map[string]string
	Body   io.Reader
}

type hikkaServerParent interface {
	Log(logger.Level, string, ...interface{})
}

type hikkaServer struct {
	hikkaAlwaysRemux     bool
	hikkaSegmentCount    int
	hikkaSegmentDuration conf.StringDuration
	hikkaAllowOrigin     string
	readBufferCount      int
	pathManager          *pathManager
	metrics              *metrics
	parent               hikkaServerParent
	request              chan hikkaMuxerRequest
	ctx                  context.Context
	ctxCancel            func()
	wg                   sync.WaitGroup
	ln                   net.Listener
}

func newHikkaServer(
	parentCtx context.Context,
	address string,
	hikkaAlwaysRemux bool,
	hikkaSegmentCount int,
	hikkaSegmentDuration conf.StringDuration,
	hikkaAllowOrigin string,
	readBufferCount int,
	pathManager *pathManager,
	metrics *metrics,
	parent hikkaServerParent,
) (*hikkaServer, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	ctx, ctxCancel := context.WithCancel(parentCtx)

	s := &hikkaServer{
		hikkaAlwaysRemux:     hikkaAlwaysRemux,
		hikkaSegmentCount:    hikkaSegmentCount,
		hikkaSegmentDuration: hikkaSegmentDuration,
		hikkaAllowOrigin:     hikkaAllowOrigin,
		readBufferCount:      readBufferCount,
		pathManager:          pathManager,
		parent:               parent,
		metrics:              metrics,
		request:              make(chan hikkaMuxerRequest),
		ctx:                  ctx,
		ctxCancel:            ctxCancel,
		ln:                   ln,
	}

	s.log(logger.Info, "listener opened on "+address)

	s.wg.Add(1)
	go s.run()

	return s, nil
}

// Log is the main logging function.
func (s *hikkaServer) log(level logger.Level, format string, args ...interface{}) {
	s.parent.Log(level, "[hikka] "+format, append([]interface{}{}, args...)...)
}

func (s *hikkaServer) close() {
	s.ctxCancel()
	s.wg.Wait()
	s.log(logger.Info, "listener closed")
}

func (s *hikkaServer) run() {
	defer s.wg.Done()

	router := gin.New()
	// router.NoRoute(s.onRequest)

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.GET("/open/door/:ip", openDoor)

	hs := &http.Server{Handler: router}
	go hs.Serve(s.ln)

outer:
	for {
		select {
		case <-s.ctx.Done():
			break outer
		}
	}
	s.ctxCancel()

	hs.Shutdown(context.Background())

}

func openDoor(c *gin.Context) {
	ip := c.Param("ip")
	hikka.OpenDoor(ip, "admin", "Pccwc@m5")
	c.String(http.StatusOK, "Hello %s", ip)
}

func (s *hikkaServer) onRequest(ctx *gin.Context) {
	s.log(logger.Info, "[conn %v] %s %s", ctx.Request.RemoteAddr, ctx.Request.Method, ctx.Request.URL.Path)

	byts, _ := httputil.DumpRequest(ctx.Request, true)
	s.log(logger.Debug, "[conn %v] [c->s] %s", ctx.Request.RemoteAddr, string(byts))

	logw := &httpLogWriter{ResponseWriter: ctx.Writer}
	ctx.Writer = logw

	ctx.Writer.Header().Set("Server", "rtsp-simple-server")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", s.hikkaAllowOrigin)
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	switch ctx.Request.Method {
	case http.MethodGet:

	case http.MethodOptions:
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", ctx.Request.Header.Get("Access-Control-Request-Headers"))
		ctx.Writer.WriteHeader(http.StatusOK)
		return

	default:
		ctx.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	// remove leading prefix
	pa := ctx.Request.URL.Path[1:]

	switch pa {
	case "", "favicon.ico":
		ctx.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	dir, fname := func() (string, string) {
		if strings.HasSuffix(pa, ".ts") || strings.HasSuffix(pa, ".m3u8") {
			return gopath.Dir(pa), gopath.Base(pa)
		}
		return pa, ""
	}()

	if fname == "" && !strings.HasSuffix(dir, "/") {
		ctx.Writer.Header().Set("Location", "/"+dir+"/")
		ctx.Writer.WriteHeader(http.StatusMovedPermanently)
		return
	}

	dir = strings.TrimSuffix(dir, "/")

	cres := make(chan hikkaMuxerResponse)
	hreq := hikkaMuxerRequest{
		Dir:  dir,
		File: fname,
		Req:  ctx.Request,
		Res:  cres,
	}

	select {
	case s.request <- hreq:
		res := <-cres

		for k, v := range res.Header {
			ctx.Writer.Header().Set(k, v)
		}
		ctx.Writer.WriteHeader(res.Status)

		if res.Body != nil {
			io.Copy(ctx.Writer, res.Body)
		}

	case <-s.ctx.Done():
	}

	s.log(logger.Debug, "[conn %v] [s->c] %s", ctx.Request.RemoteAddr, logw.dump())
}
