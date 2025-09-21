package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// ; Config 配置文件结构
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Output   OutputConfig   `yaml:"output"`
}

// ; DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	DSN      string `yaml:"dsn"`
}

// ; OutputConfig 输出配置
type OutputConfig struct {
	Dir     string `yaml:"dir"`
	PkgName string `yaml:"pkgname"`
}

// ; 默认配置
var defaultConfig = Config{
	Database: DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "p@ssw0rd",
		DBName:   "dbname",
		DSN:      "",
	},
	Output: OutputConfig{
		Dir:     "./outputs",
		PkgName: "gorm2model",
	},
}

// ; 默认配置文件内容模板
const defaultConfigTemplate = `# genGormStruct 默认配置文件
# 命令行 flag 优先级高于配置文件

database:
  host: "127.0.0.1"      # 数据库主机地址
  port: "3306"           # 数据库端口
  user: "root"           # 数据库用户名
  password: "p@ssw0rd"   # 数据库密码
  dbname: "dbname"       # 数据库名称
  # dsn: ""              # 自定义 DSN（如果设置，将忽略上述配置）
  ##; 示例@{dsn}: user:pass@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

output:
  dir: "./outputs"       # 输出目录
  pkgname: "gorm2model"  # 生成的 go 包名
  ##; 生成的 go 包名，用于组织生成的包（默认值：gorm2model）
`

func main() {
	//; 检查子命令
	if len(os.Args) > 1 && os.Args[1] == "init-config" {
		if err := initConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	//; 定义 flag
	helpFlag := flag.Bool("h", false, "Display this help message")
	configFile := flag.String("c", "", "config file path (yaml format), sample: config.yaml")

	//; 数据库相关 flag
	password := flag.String("password", "", "database password, : p@ssw0rd")
	user := flag.String("user", "", "database user")
	host := flag.String("host", "", "database host")
	port := flag.String("port", "", "database port")
	dbname := flag.String("dbname", "", "database name")
	dsn := flag.String("dsn", "", "database uri, default: '', override other config if set !! (format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local)")

	//; 输出相关 flag
	output := flag.String("output", "", "output directory for generated gorm code")
	pkgname := flag.String("pkgname", "", "go package name for generated gorm code")

	//; 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  init-config    Generate default config file (config.yaml)\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  -c string\n\tconfig file path (yaml format)\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -c config.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -host 127.0.0.1 -port 3306 -user root -password secret -dbname mydb\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsn \"user:pass@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local\"\n", os.Args[0])
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	//; 加载配置文件（如果有）
	cfg := defaultConfig
	if *configFile != "" {
		loadedCfg, err := loadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(1)
		}
		cfg = *loadedCfg
		fmt.Fprintf(os.Stdout, "Loaded config from: %s\n", *configFile)
	}

	//; 命令行 flag 优先级高于配置文件
	if *password != "" {
		cfg.Database.Password = *password
	} else if cfg.Database.Password == "" {
		cfg.Database.Password = defaultConfig.Database.Password
	}

	if *user != "" {
		cfg.Database.User = *user
	} else if cfg.Database.User == "" {
		cfg.Database.User = defaultConfig.Database.User
	}

	if *host != "" {
		cfg.Database.Host = *host
	} else if cfg.Database.Host == "" {
		cfg.Database.Host = defaultConfig.Database.Host
	}

	if *port != "" {
		cfg.Database.Port = *port
	} else if cfg.Database.Port == "" {
		cfg.Database.Port = defaultConfig.Database.Port
	}

	if *dbname != "" {
		cfg.Database.DBName = *dbname
	} else if cfg.Database.DBName == "" {
		cfg.Database.DBName = defaultConfig.Database.DBName
	}

	if *output != "" {
		cfg.Output.Dir = *output
	} else if cfg.Output.Dir == "" {
		cfg.Output.Dir = defaultConfig.Output.Dir
	}

	if *pkgname != "" {
		cfg.Output.PkgName = *pkgname
	} else if cfg.Output.PkgName == "" {
		cfg.Output.PkgName = defaultConfig.Output.PkgName
	}

	//; DSN 处理：flag > 配置文件 > 自动构建
	dsnQuery := "charset=utf8mb4&parseTime=True&loc=Local"
	finalDSN := cfg.Database.DSN
	if *dsn != "" {
		finalDSN = *dsn
	}

	if finalDSN == "" {
		encodedPassword := url.QueryEscape(cfg.Database.Password)
		finalDSN = cfg.Database.User + ":" + encodedPassword + "@tcp(" + cfg.Database.Host + ":" + cfg.Database.Port + ")/" + cfg.Database.DBName + "?" + dsnQuery
		fmt.Fprintf(os.Stdout, "format dsn : %s\n", finalDSN)
	} else {
		fmt.Fprintf(os.Stdout, "custom dsn : %s\n", finalDSN)
	}

	//;; if not exist, create output dir
	if _, err := os.Stat(cfg.Output.Dir); os.IsNotExist(err) {
		os.MkdirAll(cfg.Output.Dir, 0755)
	}
	absPath, _ := filepath.Abs(cfg.Output.Dir)
	fmt.Fprintf(os.Stdout, "output dir : %s (%s) \n", cfg.Output.Dir, absPath)

	//;;
	g := gen.NewGenerator(gen.Config{
		OutPath:      cfg.Output.Dir,
		OutFile:      "",
		ModelPkgPath: filepath.Join(cfg.Output.Dir, cfg.Output.PkgName),
		WithUnitTest: true,
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	gormdb, _ := gorm.Open(mysql.Open(finalDSN), &gorm.Config{})
	g.UseDB(gormdb) ////;reuse your gorm db

	////; list all tables
	tables, _ := gormdb.Migrator().GetTables()
	fmt.Fprintf(os.Stdout, "tables : %v \n", tables)
	cnt := 0
	for _, table := range tables {
		//;; skip some tables with prefix
		if strings.Contains(table, ".") || strings.HasSuffix(table, "_tmp") || strings.HasSuffix(table, "_bak") {
			fmt.Fprintf(os.Stdout, "skip table : %s \n", table)
		} else {
			cnt = cnt + 1
			fmt.Fprintf(os.Stdout, "todo table : %s \n", table)
			g.ApplyBasic(g.GenerateModel(table))
		}
	}
	fmt.Fprintf(os.Stdout, "cnt : %d \n", cnt)
	////;Generate the code
	g.Execute()
	fmt.Fprintf(os.Stdout, "out : %s/%s \n", cfg.Output.Dir, cfg.Output.PkgName)
	fmt.Println("done")
}

// ; initConfig 生成默认配置文件
func initConfig() error {
	configPath := "config.yaml"

	//; 检查文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file '%s' already exists", configPath)
	}

	//; 写入默认配置
	err := os.WriteFile(configPath, []byte(defaultConfigTemplate), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Generated default config file: %s\n", configPath)
	fmt.Fprintf(os.Stdout, "Please edit the file with your database credentials before running.\n")
	return nil
}

// ; loadConfig 从文件加载配置
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
