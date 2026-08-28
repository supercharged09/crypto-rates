package logger

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// настройки логирования
type Config struct {
	//LogFile - путь к файлу логов
	LogFile string
	//MaxSize - макс размер файла в Мб перед ротацией
	MaxSize int
	//MaxBackups - макс кол-во старых файлов
	MaxBackups int
	//MaxAge - макс кол-во дней хранения старых файлов
	MaxAge int
	//Compress - сжимать ли старые файлы
	Compress bool
	//AlsoToStdout - дублировать ли логи в консоль
	AlsoToStdout bool
}

// DefaultConfig возвращает конфиг по дефолту
func DefaultConfig() Config {
	return Config{
		LogFile:      "logs/crypto-rates.log",
		MaxSize:      10, //10Mb
		MaxBackups:   5,  //храним 5 старых файлов
		MaxAge:       30, //удалять старые файлы старше 30 дней
		Compress:     true,
		AlsoToStdout: true,
	}
}

// Setup настраивает глобальный логгер
func Setup(cfg Config) {
	//настройка ротации файла
	fileLogger := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	//если надо дублировать в консоль, используем MultiWriter
	var writer io.Writer
	if cfg.AlsoToStdout {
		writer = io.MultiWriter(os.Stdout, fileLogger)
	} else {
		writer = fileLogger
	}

	//установка глобального вывода логов
	log.SetOutput(writer)

	//настройка формата: дата, время, файл:строка
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Logger initialized")
}
