package logger

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Setup настраивает глобальный логгер
func Setup(
	//LogFile - путь к файлу логов
	logFile string,
	//MaxSize - макс размер файла в Мб перед ротацией
	maxSizeMB int,
	//MaxBackups - макс кол-во старых файлов
	maxBackups int,
	//MaxAge - макс кол-во дней хранения старых файлов
	maxAgeDays int,
	//Compress - сжимать ли старые файлы
	compress bool,
	//AlsoToStdout - дублировать ли логи в консоль
	alsoToStdout bool,
) {
	// настройка ротации файла
	fileLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}

	// если надо дублировать в консоль, используем MultiWriter
	var writer io.Writer
	if alsoToStdout {
		writer = io.MultiWriter(os.Stdout, fileLogger)
	} else {
		writer = fileLogger
	}

	// установка глобального вывода логов
	log.SetOutput(writer)

	// настройка формата: дата, время, файл:строка
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Logger initialized")
}
