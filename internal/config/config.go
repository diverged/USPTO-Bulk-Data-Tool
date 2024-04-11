package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type TuningConfig struct {
	MaxConcurrentZips int
	BufferSize        int
}

type OutputConfig struct {
	TextFormat         string
	ParquetCompression string
}

type LoggerConfig struct {
	LogMode  string
	LogLevel string
	LogPath  string
}

type DevConfig struct {
	CleanOutput      bool
	ParserReturnsRaw bool
}

type Config struct {
	InputDir    string
	OutputDir   string
	OutputMode  string
	RunTime     time.Time
	CleanOutput bool

	TuningConfig TuningConfig

	OutputConfig OutputConfig

	LoggerConfig LoggerConfig

	DevConfig DevConfig
}

func LoadConfig(cliArgConfigPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	if cliArgConfigPath != "" {
		viper.SetConfigFile(cliArgConfigPath)
	} else {
		fmt.Println("No config path provided, using default path")
		viper.AddConfigPath("./")
	}

	viper.SetDefault("output.textformatting", "innerxml")
	viper.SetDefault("output.parquetcompression", "snappy")

	viper.SetDefault("logging.logmode", "prod")
	viper.SetDefault("logging.loglevel", "warn")
	viper.SetDefault("logging.logdirectory", "data/logfiles")

	viper.SetDefault("tuning.maxconcurrentzips", 0)
	viper.SetDefault("tuning.channelbuffersize", 100)

	viper.SetDefault("dev.cleanoutput", false)
	viper.SetDefault("dev.parserreturnsraw", false)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	return &Config{
		InputDir:   viper.GetString("required.inputdirectory"),
		OutputDir:  viper.GetString("required.outputdirectory"),
		OutputMode: viper.GetString("required.outputmode"),
		RunTime:    time.Now(),

		TuningConfig: TuningConfig{
			MaxConcurrentZips: viper.GetInt("tuning.maxconcurrentzips"),
			BufferSize:        viper.GetInt("tuning.channelbuffersize"),
		},

		OutputConfig: OutputConfig{
			TextFormat:         viper.GetString("output.textformatting"),
			ParquetCompression: viper.GetString("out.parquetcompression"),
		},

		LoggerConfig: LoggerConfig{
			LogMode:  viper.GetString("logging.logmode"),
			LogLevel: viper.GetString("logging.loglevel"),
			LogPath:  viper.GetString("logging.logdirectory"),
		},

		DevConfig: DevConfig{
			CleanOutput:      viper.GetBool("dev.cleanoutput"),
			ParserReturnsRaw: viper.GetBool("dev.parserreturnsraw"),
		},
	}, nil
}
