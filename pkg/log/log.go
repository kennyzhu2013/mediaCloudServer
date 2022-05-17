package log

import log "common/log/log"

var Logger = log.GetLogger()

type Config struct {
	Level   int32
	Dir     string
	Path    string
	FileNum int
}

// trans Logger struct.
func Init(serverId string, c Config) {
	// init log here.
	log.InitLogger(
		log.WithLevel(log.Level(c.Level)),
		log.WithFields(log.Fields{
			0: {0: "MediaServer:" + serverId},
		}),
		log.WithOutput(
			log.NewOutput(
				log.OutputDir(c.Dir), log.OutputName(c.Path),
				//log.NewAsyncOptions(log.EnableAsync(false)),
			),
		),
	)

	if c.FileNum > 0 {
		log.DefaultFileMaxNum = c.FileNum
	}

	Logger = log.GetLogger()
	log.Debugf("log init.")
}

func SetLevel(level int) {
	log.SetDefaultOption(log.WithLevel(log.Level(level)))
}
