module github.com/shenjl/rtsp-simple-server

go 1.17

require (
	github.com/aler9/gortsplib v0.0.0-20211130212324-870687d91d98
	github.com/asticode/go-astits v1.10.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.7
	github.com/gookit/color v1.4.2
	github.com/grafov/m3u8 v0.11.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/notedit/rtmp v0.0.2
	github.com/pion/rtp v1.6.2
	github.com/pion/sdp/v3 v3.0.2
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/fatih/color v1.13.0
	github.com/go-delve/delve v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.2.1 // indirect
	go.starlark.net v0.0.0-20211013185944-b0039bd2cfe3 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
)

replace github.com/notedit/rtmp => github.com/aler9/rtmp v0.0.0-20210403095203-3be4a5535927
