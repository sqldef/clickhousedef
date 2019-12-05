package main

import (
	"fmt"
	"log"
	"os"

	"github.com/howeyc/gopass"
	"github.com/jessevdk/go-flags"
	"github.com/k0kubun/sqldef"
	"github.com/k0kubun/sqldef/adapter"
	"github.com/k0kubun/sqldef/schema"
	"github.com/sqldef/clickhousedef/adapter/clickhouse"
)

// Return parsed options and schema filename
// TODO: Support `sqldef schema.sql -opt val...`
func parseOptions(args []string) (adapter.Config, *sqldef.Options) {
	var opts struct {
		User     string `short:"U" long:"user" description:"ClickHouse user name" value-name:"username"`
		Password string `short:"W" long:"password" description:"ClickHouse user password, overridden by $CHPASS"`
		Host     string `short:"h" long:"host" description:"Host or socket directory to connect to the ClickHouse server" value-name:"hostname" default:"127.0.0.1"`
		Port     uint   `short:"p" long:"port" description:"Port used for the connection" value-name:"port" default:"9000"`
		Prompt   bool   `long:"password-prompt" description:"Force ClickHouse user password prompt"`
		File     string `short:"f" long:"file" description:"Read schema SQL from the file, rather than stdin" value-name:"filename" default:"-"`
		DryRun   bool   `long:"dry-run" description:"Don't run DDLs but just show them"`
		Export   bool   `long:"export" description:"Just dump the current schema to stdout"`
		SkipDrop bool   `long:"skip-drop" description:"Skip destructive changes such as DROP"`
		Help     bool   `long:"help" description:"Show this help"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Usage = "[option...] db_name"
	args, err := parser.ParseArgs(args)
	if err != nil {
		log.Fatal(err)
	}

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if len(args) == 0 {
		fmt.Print("No database is specified!\n\n")
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	} else if len(args) > 1 {
		fmt.Printf("Multiple databases are given: %v\n\n", args)
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	database := args[0]

	options := sqldef.Options{
		SqlFile:  opts.File,
		DryRun:   opts.DryRun,
		Export:   opts.Export,
		SkipDrop: opts.SkipDrop,
	}

	password, ok := os.LookupEnv("CHPASS")
	if !ok {
		password = opts.Password
	}

	if opts.Prompt {
		fmt.Printf("Enter Password: ")
		pass, err := gopass.GetPasswd()
		if err != nil {
			log.Fatal(err)
		}
		password = string(pass)
	}

	config := adapter.Config{
		DbName:   database,
		User:     opts.User,
		Password: password,
		Host:     opts.Host,
		Port:     int(opts.Port),
	}
	if _, err := os.Stat(config.Host); !os.IsNotExist(err) {
		config.Socket = config.Host
	}
	return config, &options
}

func main() {
	config, options := parseOptions(os.Args[1:])

	database, err := clickhouse.NewDatabase(config)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	sqldef.Run(schema.GeneratorModeMysql, database, options)
}
