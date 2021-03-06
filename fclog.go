package fclog

import "os"
import "log"
import "sync"
import "fmt"
import "strconv"
import "time"
import "runtime"
import "strings"

var fcLog *FCLog = nil

func GetFcLog() *FCLog {
	if fcLog == nil {
		fcLog = new(FCLog)
	}

	return fcLog
}

type FCLog struct {
	logFile       *os.File
	bPrintConsole bool
	bWriteFile    bool
	logName       string
	logPath       string
	logFileName   string
	fileSizeByte  int64
	suffix        int
	lockConsole   sync.Mutex
	pid           string
	logLevel      int
}

const (
	LEVEL_ALL     = 0
	LEVEL_DEBUG   = 1
	LEVEL_INFO    = 2
	LEVEL_WARNING = 3
	LEVEL_ERROR   = 4
	LEVEL_NONE    = 5
)

func (l *FCLog) CheckLogsDir() error {
	_, err := os.Stat("./logs")
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return os.Mkdir("./logs", 0770)
	}

	return nil
}

func (l *FCLog) InitFCLog(bPrintConsole bool, bWriteFile bool, logName string, fileSizeByte int64, level int) error {

	err := l.CheckLogsDir()
	if err != nil {
		return err
	}

	l.bPrintConsole = bPrintConsole
	l.bWriteFile = bWriteFile
	l.logFile = nil
	l.pid = strconv.Itoa(os.Getpid())
	l.logPath = "./logs/" + logName + "." + l.pid
	l.logName = logName
	l.fileSizeByte = fileSizeByte
	l.suffix = 1
	l.logLevel = level

	if l.bWriteFile {
		l.logFile, err = os.Create(l.logPath + ".log")
		if err != nil {
			return err
		}

		l.logFileName = l.logPath + ".log"
	}

	return nil
}

func (l *FCLog) SetLogLevel(level int) {

	l.logLevel = level
}

func (l *FCLog) Debug(format string, v ...interface{}) {
	if l.logLevel <= LEVEL_DEBUG {
		l.Write(format, "DEBUG", v...)
	}
}

func (l *FCLog) Info(format string, v ...interface{}) {
	if l.logLevel <= LEVEL_INFO {
		l.Write(format, "INFO", v...)
	}
}
func (l *FCLog) Warning(format string, v ...interface{}) {
	if l.logLevel <= LEVEL_WARNING {
		l.Write(format, "WARN", v...)
	}
}
func (l *FCLog) Error(format string, v ...interface{}) {
	if l.logLevel <= LEVEL_ERROR {
		l.Write(format, "ERROR", v...)
	}
}

func (l *FCLog) Write(format string, level string, v ...interface{}) {

	l.lockConsole.Lock()

	defer l.lockConsole.Unlock()

	curTime := time.Now()
	_, file, line, _ := runtime.Caller(3)
	index := strings.LastIndex(file, "/")
	subFile := file[index+1 : len(file)]

	if l.bPrintConsole {

		levelColor := ""

		switch {
		case level == "DEBUG":
			levelColor = "\x1b[0;37m" + level + "\x1b[0m"
		case level == "INFO":
			levelColor = "\x1b[0;32m" + level + "\x1b[0m"
		case level == "WARN":
			levelColor = "\x1b[0;33m" + level + "\x1b[0m"
		case level == "ERROR":
			levelColor = "\x1b[0;31m" + level + "\x1b[0m"
		}

		fmt.Printf("%d-%02d-%02d %02d:%02d:%02d:%09d %s:%d %s ", curTime.Year(), curTime.Month(), curTime.Day(), curTime.Hour(), curTime.Minute(), curTime.Second(), curTime.Nanosecond(), subFile, line, levelColor)
		fmt.Printf(format, v...)
		fmt.Printf("\n")
	}

	if l.bWriteFile && l.logFile != nil {
		fileInfo, err := os.Stat(l.logFileName)
		if err == nil {
			if fileInfo.Size() >= l.fileSizeByte {
				l.logFile.Close()

				renamePath := l.logPath + "-" + strconv.Itoa(l.suffix) + ".log"
				err = os.Rename(l.logFileName, renamePath)
				if err != nil {
					log.Panic("Rename file error")
					return
				}

				l.suffix += 1

				l.logFile, err = os.Create(l.logFileName)
				if err != nil {
					log.Panic("Create log file error!")
					return
				}
			}
		}

		outputStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d:%09d %s:%d %s ", curTime.Year(), curTime.Month(), curTime.Day(), curTime.Hour(), curTime.Minute(), curTime.Second(), curTime.Nanosecond(), subFile, line, level)
		l.logFile.WriteString(outputStr)
		l.logFile.WriteString(fmt.Sprintf(format, v...))
		l.logFile.WriteString("\n")

	}
}

func Init(bPrintConsole bool, bWriteFile bool, logName string, fileSizeByte int64, level int) error {
	return GetFcLog().InitFCLog(bPrintConsole, bWriteFile, logName, fileSizeByte, level)
}

func DEBUG(format string, v ...interface{}) {
	GetFcLog().Debug(format, v...)
}

func INFO(format string, v ...interface{}) {
	GetFcLog().Info(format, v...)
}

func WARN(format string, v ...interface{}) {
	GetFcLog().Warning(format, v...)
}

func ERROR(format string, v ...interface{}) {
	GetFcLog().Error(format, v...)
}
