package logger

/**
 * User: Santiago Vidal
 * Date: 10/05/17
 * Time: 12:42
 */

import (
	"eco/services/config"
	"eco/services/net"
	"eco/services/persistence/db"
	"eco/services/session"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ian-kent/go-log/appenders"
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/log"
)

const logfile = "ecoserver.log"

func Init() {

	debug, _ := config.EcoCfg.Bool("debug")

	logg := log.Logger()
	r := appenders.RollingFile(logfile, true)
	r.MaxBackupIndex = 20
	r.MaxFileSize = 200 * 1024

	ecoFileAppender := &ecoLogAppender{
		nested:  r,
		logInfo: true,
	}

	if debug {

		ecoConsoleAppender := &ecoLogAppender{
			nested:  appenders.Console(),
			logInfo: false,
		}

		multi := appenders.Multiple(nil, ecoFileAppender, ecoConsoleAppender)
		logg.SetAppender(multi)
		logg.SetLevel(levels.TRACE)

	} else {
		logg.SetAppender(ecoFileAppender)
		logLevel, _ := config.EcoCfg.String("log.level")

		switch logLevel {

		case "info":
			logg.SetLevel(levels.INFO)
			break

		case "error":
			logg.SetLevel(levels.ERROR)
			break

		default:
			logg.SetLevel(levels.FATAL)

		}
	}

	path, _ := filepath.Abs(logfile)
	fmt.Fprintf(os.Stdout, "Loggin to %s\n", path)
}

func LogHandler(inner net.EcoApiHttp, name string) net.EcoApiHandler {

	return net.EcoApiHandler(func(sess *session.EcoSession) error {

		start := time.Now()
		err := inner.ServeHTTP(sess)
		if err != nil {
			return err
		}
		_, req := sess.GetHttp()
		timeElapsed := time.Since(start)

		go func(method string, uri string, ecoSession *session.EcoSession, ip string) {

			var user string
			if ecoSession != nil {
				user = ecoSession.EcoUser.Email
			} else {
				user = "unknown"
			}

			log.Printf(
				"%s\t%s\t%s\t%s user:%s origin:%s",
				method,
				uri,
				name,
				timeElapsed,
				user,
				ip,
			)

			if len(uri) > 8 && uri[0:8] == "/ecoapi/" {

				str := strings.Split(uri, "/")
				uri = "/" + str[1] + "/" + str[2]

				s := db.Redis.Session()
				defer s.Close()
				//c := persistence.GetRedisPool().Get()
				c := s.GetConn()

				_, err := c.Do("HINCRBY", "api_usage:"+uri, "count", 1)
				if err != nil {
					log.Error(err)
					return
				}

				_, err = c.Do("HINCRBY", "api_usage:"+uri, "time", uint64(timeElapsed))
				if err != nil {
					log.Error(err)
					return
				}
				c.Flush()
			}

		}(req.Method, req.RequestURI, sess, req.RemoteAddr)
		return nil

	})
}

type ecoLogAppender struct {
	nested  appenders.Appender
	logInfo bool
}

func (a *ecoLogAppender) Write(level levels.LogLevel, message string, args ...interface{}) {

	//en modo debug no siempre escribe logs de INFO.
	//por ejemplo en la consola de la IDE no lo hace, solo errores y warnings
	if level == levels.INFO && !a.logInfo {
		return //rechazado
	}
	a.nested.Write(level, time.Now().Format("2006/01/02 15:04:05")+" - "+message, args...)

}
func (a *ecoLogAppender) SetLayout(layout layout.Layout) {
	a.nested.SetLayout(layout)
}

func (a *ecoLogAppender) Layout() layout.Layout {
	return a.nested.Layout()
}
