package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const VERSION = `0.2.1`

var debug = debugT(false)

type debugT bool

func (d debugT) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

// Run is main routine.
func (trdsql TRDSQL) Run(args []string) int {
	var (
		version bool
		odriver string
		odsn    string
		iltsv   bool
		inSep   string
		ihead   bool
		iskip   int
		query   string
		driver  string
		dsn     string
		odebug  bool
	)
	type Format int
	var (
		oltsv bool
		oat   bool
		omd   bool
		oraw  bool
		fjson bool
	)
	flags := flag.NewFlagSet("trdsql", flag.ContinueOnError)
	driver = "sqlite3"
	dsn = ""
	cfgfile := configOpen()
	cfg, _ := loadConfig(cfgfile)
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage: %s [OPTIONS] [SQL(SELECT...)]

Options:
`, os.Args[0])
		flags.PrintDefaults()
	}

	flags.StringVar(&cfg.Db, "db", cfg.Db, "Specify db name of the setting.")
	flags.StringVar(&odriver, "driver", "", "database driver.  [ "+strings.Join(sql.Drivers(), " | ")+" ]")
	flags.StringVar(&odsn, "dsn", "", "database connection option.")
	flags.BoolVar(&iltsv, "iltsv", false, "LTSV format for input.")
	flags.StringVar(&inSep, "id", ",", "Field delimiter for input.")
	flags.StringVar(&trdsql.outSep, "od", ",", "Field delimiter for output.")
	flags.BoolVar(&ihead, "ih", false, "The first line is interpreted as column names.")
	flags.BoolVar(&oltsv, "oltsv", false, "LTSV format for output.")
	flags.BoolVar(&oat, "oat", false, "ASCII Table format for output.")
	flags.BoolVar(&omd, "omd", false, "Mark Down format for output.")
	flags.BoolVar(&oraw, "oraw", false, "Raw format for output.")
	flags.BoolVar(&fjson, "ojson", false, "Json format for output.")
	flags.BoolVar(&trdsql.outHeader, "oh", false, "Output column name as header.")
	flags.IntVar(&iskip, "is", 0, "Skip header row.")
	flags.StringVar(&query, "q", "", "Read query from the provided filename.")
	flags.BoolVar(&version, "version", false, "display version information.")
	flags.BoolVar(&odebug, "debug", false, "debug print.")
	flags.Parse(args[1:])
	if version {
		fmt.Println(VERSION)
		return (0)
	}
	var sqlstr string
	if query != "" {
		bq, err := ioutil.ReadFile(query)
		if err != nil {
			log.Println("ERROR: ", err)
			return (1)
		}
		sqlstr = string(bq)
	} else {
		sqlstr = strings.Join(flags.Args(), " ")
	}
	if len(sqlstr) == 0 {
		flags.Usage()
		return (2)
	}
	if odebug {
		debug = true
	}
	if strings.HasSuffix(sqlstr, ";") {
		sqlstr = sqlstr[:len(sqlstr)-1]
	}

	if cfg.Db != "" {
		if cfg.Database[cfg.Db].Driver == "" {
			debug.Printf("ERROR: db[%s] does not found", cfg.Db)
		} else {
			driver = cfg.Database[cfg.Db].Driver
			dsn = cfg.Database[cfg.Db].Dsn
		}
	}
	if odriver != "" {
		driver = odriver
	}
	if odsn != "" {
		dsn = odsn
	}

	debug.Printf("driver: %s, dsn: %s", driver, dsn)
	db, err := Connect(driver, dsn)
	if err != nil {
		log.Println("ERROR: ", err)
		return 1
	}
	defer db.Disconnect()

	tablenames := sqlparse(sqlstr)
	if len(tablenames) == 0 {
		// without FROM clause. ex. SELECT 1+1;
		debug.Printf("table not found\n")
	}
	trdsql.iskip = iskip
	var r int
	if iltsv {
		trdsql.inSep = "\t"
		sqlstr, r = trdsql.ltsvReader(db, sqlstr, tablenames)
	} else {
		trdsql.inSep = inSep
		trdsql.ihead = ihead
		sqlstr, r = trdsql.csvReader(db, sqlstr, tablenames)
	}
	if r != 0 {
		return r
	}
	switch {
	case oltsv:
		r = trdsql.ltsvWrite(db, sqlstr)
	case fjson:
		r = trdsql.jsonWrite(db, sqlstr)
	case oraw:
		r = trdsql.rawWrite(db, sqlstr)
	case omd:
		trdsql.omd = true
		r = trdsql.twWrite(db, sqlstr)
	case oat:
		r = trdsql.twWrite(db, sqlstr)
	default:
		r = trdsql.csvWrite(db, sqlstr)
	}
	return r
}

func getSeparator(sepString string) (rune, error) {
	if sepString == "" {
		return 0, nil
	}
	sepRunes, err := strconv.Unquote(`'` + sepString + `'`)
	if err != nil {
		return ',', fmt.Errorf("ERROR getSeparator: %s:%s", err, sepString)
	}
	sepRune := ([]rune(sepRunes))[0]
	return sepRune, err
}

func tFileOpen(filename string) (*os.File, error) {
	if filename == "-" {
		return os.Stdin, nil
	}
	if filename[0] == '`' {
		filename = strings.Replace(filename, "`", "", 2)
	}
	if filename[0] == '"' {
		filename = strings.Replace(filename, "\"", "", 2)
	}
	return os.Open(filename)
}

func valString(v interface{}) string {
	var str string
	b, ok := v.([]byte)
	if ok {
		str = string(b)
	} else {
		if v == nil {
			str = ""
		} else {
			str = fmt.Sprint(v)
		}
	}
	return str
}
