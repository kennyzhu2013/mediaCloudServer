module xmediaEmu

go 1.16

require (
	common v0.0.0
	github.com/bitly/go-simplejson v0.5.0
	github.com/chromedp/chromedp v0.8.1
	github.com/gin-gonic/gin v1.7.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/gorilla/websocket v1.4.2
	github.com/hajimehoshi/ebiten/v2 v2.3.2
	github.com/hajimehoshi/go-mp3 v0.3.3
	github.com/hajimehoshi/oto/v2 v2.1.0
	github.com/jfreymuth/oggvorbis v1.0.3
	github.com/pterm/pterm v0.12.33
	github.com/satori/go.uuid v1.2.0
	golang.org/x/image v0.0.0-20220321031419-a8550c1d254a
	gopkg.in/yaml.v2 v2.3.0
)

replace common => ../go-common
