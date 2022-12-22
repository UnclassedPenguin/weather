//-------------------------------------------------------------------------------
//-------------------------------------------------------------------------------
//  
// Tyler(UnclassedPenguin) Weather 2022
//  
// Author: Tyler(UnclassedPenguin)
//    URL: https://unclassed.ca
// GitHub: https://github.com/UnclassedPenguin/weather/
//   Desc: A program to keep track of weather.
//
//-------------------------------------------------------------------------------
//-------------------------------------------------------------------------------

package main

import (
  "os"
  "fmt"
  "time"
  "flag"
  "strconv"
  "strings"
  "os/exec"
  "io/ioutil"
  "database/sql"
  "path/filepath"
  "gopkg.in/yaml.v2"
  _ "github.com/mattn/go-sqlite3"
  c "github.com/unclassedpenguin/weather/config"
  d "github.com/unclassedpenguin/weather/database"
  f "github.com/unclassedpenguin/weather/functions"
)

// Main Function
func main() {

  // Flags
  var (
    info         bool
    list         bool
    test         bool
    add          bool
    del          bool
    push         bool
    pull         bool
    status       bool
    version      bool
    debug        bool
    dateNewToOld bool
    dateOldToNew bool
    currentMonth bool
    today        bool
    showSql      bool
    number       int
    high         int
    low          int
    notes        string
    year         string
    month        string
    day          string
    date         string
    dateFrom     string
    custom       string
  )

  flag.BoolVar(        &info,        "i", false,
    "Prints some information you might need to remember.")
  flag.BoolVar(        &list,        "l", false,
    "Prints the Database to terminal. Optionally you can use -g, -s, -r, -y, -m, -date...")
  flag.BoolVar(        &test,        "t", false,
    "If set, uses the test database.")
  flag.BoolVar(         &add,        "a", false,
    "Adds a record to the database. If set, requires -high and -low.")
  flag.BoolVar(         &del,        "d", false,
    "Deletes a record from the database. If set, requires -n (id number of entry to delete),\n" +
    "or -g (animal group to delete).")
  flag.BoolVar(        &push,     "push", false,
    "Pushes the databases with git.")
  flag.BoolVar(        &pull,     "pull", false,
    "Pulls the databases with git.")
  flag.BoolVar(      &status,   "status", false,
    "Checks the git status on project.")
  flag.BoolVar(     &version,        "v", false,
    "Print the version number and exit.")
  flag.BoolVar(       &debug,    "debug", false,
    "Execute function for debugging.")
  flag.BoolVar(&dateNewToOld,     "desc", false,
    "List entires in descending order. Requires -l")
  flag.BoolVar(&dateOldToNew,      "asc", false,
    "List entries in ascending order. Requires -l")
  flag.BoolVar(&currentMonth,        "m", false,
    "List only current month.")
  flag.BoolVar(     &showSql,      "sql", false,
    "Show SQL query when listing.")
  flag.BoolVar(       &today,    "today", false,
    "Show entries only for current day.")

  flag.StringVar(      &year,     "year",    "",
    "Year to list from database. Can be a single year(ie 2019) or a range (ie 2019-2022)")
  flag.StringVar(     &month,    "month",    "",
    "Month to list from database. Can be a single month(ie 09) or a range (ie 09-12). \nSingle " +
    "digit months require a leading 0.")
  flag.StringVar(       &day,      "day",    "",
    "Day to list from database. Can be a single day(ie 19) or a range (ie 09-30)")
  flag.StringVar(      &date,     "date",    "",
    "The date to put into the database, if not today. yyyy-mm-dd")
  flag.StringVar(  &dateFrom,     "from",    "",
    "List from specified date to current date. Date must be YYYY-MM-DD requires -l")
  flag.StringVar(    &custom,        "c",    "",
    "Custom SQL request. Requires -l. Example:\nweather -t -l -c \"SELECT * FROM weather WHERE " +
    "strftime('%d', date) BETWEEN '01' AND '03'\"")
  flag.StringVar( &notes, "note", "",
    "A note for the day")

    // Set to -100 to know if it has not been changed.
  flag.IntVar(      &high,     "high",     -100,
    "High temperature degrees Celcius")
  flag.IntVar(      &low,     "low",     -100,
    "Low temperature degrees Celcius")
  flag.IntVar(       &number,        "n",     0,
    "Used with -d, the id of the record to delete .")

  // This changes the help/usage info when -h is used.
  flag.Usage = func() {
      w := flag.CommandLine.Output() // may be os.Stderr - but not necessarily
      description := "Description of %s:\n\n" +
       "This is a program to use to keep track of weather.\n" +
       "Usage:\n\n" +
       "weather [-t] [-l [-g group] [-s | -r] [-year year] [-month month] [-day day] " +
       "[-a [-date YYYY-MM-DD] -g group [-s | -r] -n num] [-d [-n num || -g group]]\n\n" +
       "Available arguments:\n"
      fmt.Fprintf(w, description, os.Args[0])
      flag.PrintDefaults()
      //fmt.Fprintf(w, "...custom postamble ... \n")
  }

  // Parse the flags :p
  flag.Parse()

  // Handles cmd line flag -i 
  // Prints info and exits
  if info {
    f.PrintInfo()
  }

  // Handles cmd line flag -v 
  // Prints version and exits
  if version {
    f.PrintVersion()
  }

  if debug {
    f.DebugFunction()
  }

  // Variable to hold the date
  var timeStr string
  var todaysDate string

  t := time.Now()
  todaysDate = t.Format("2006-01-02")

  // Get either Current Date or a date entered as a command line option
  if date == "" {
    timeStr = todaysDate
  } else {
    timeStr = date
  }

  // Use regexp to check date to make sure it is a valid yyyy-mm-dd date
  if !f.CheckDate(timeStr) {
    fmt.Println("Error:")
    fmt.Println("\nIt seems your date isn't the proper format. Please enter date as YYYY-MM-DD ie 2022-01-12\n")
    os.Exit(1)
  }

  // Read Config file and setup databases
  home, _ := os.UserHomeDir()
  configFile, err := ioutil.ReadFile(filepath.Join(home, ".config/weather/config.yaml"))
  if err != nil {
    fmt.Println("Error reading config file:\n", err)
    os.Exit(1)
  }

  var configData c.Configuration
  err = yaml.Unmarshal(configFile, &configData)
  if err != nil {
    fmt.Println("Error Unmarshal-ling yaml config file:\n", err)
  }

  // I use this directory in the git section near the end
  dbDir := configData.DatabaseDir

  // Variable for databases. One for real, and one to test
  // things with, that has garbage data in it.
  var (
    realDb string
    testDb string
  )

  // This sets the database based on the config file
  realDb = configData.RealDatabase
  testDb = configData.TestDatabase

  // Change dir to database directory
  // This is needed so a database isn't created where you execute from 
  // (I have the executable soft linked to to a command in ~/.bin)
  // Keeps the database in the database directory
  err = os.Chdir(dbDir)
  if err != nil {
    fmt.Println("Error changing to directory:\n", err)
    os.Exit(1)
  }

  // Var that holds the current working database.
  var databaseToUse string

  // Says whether to use the test database or the real database. 
  // Set with -t 
  if test {
    databaseToUse = testDb
  } else {
    databaseToUse = realDb
  }

  // Creates database if it hasn't been created yet.
  d.CreateDatabase(databaseToUse)

  // Initialize database
  db, err := sql.Open("sqlite3", databaseToUse)
    if err != nil {
      fmt.Println("Error initializing database")
      os.Exit(1)
    }

  // Creates the table initially. "IF NOT EXISTS"
  d.CreateTable(db)


  // Handles the command line way to add record
  if add && high != -100 && low != -100 {
    fmt.Println("Date        : ", timeStr)
    fmt.Println("Low Temp    : ", low)
    fmt.Println("High Temp   : ", high)
    fmt.Println("Notes       : ", notes)
    fmt.Println("")
    fmt.Println("Adding record...")
    d.AddRecord(db, timeStr, low, high, notes)
    fmt.Println("Record added!")
    f.Exit(db, 0)
  } else if add {
    fmt.Println("Requires -low and -high! Try again, or try -h for help.")
    f.Exit(db, 1)
  }

  // Handles the command line way to delete record
  if del {
    if number != 0 {
      fmt.Print("Deleting record ", number , "...\n")
      str := fmt.Sprint("DELETE FROM weather WHERE id=" + strconv.Itoa(number))
      d.DeleteRecord(db, str)
      fmt.Println("Record deleted!")
      f.Exit(db, 0)
    } else {
      fmt.Println("Requires -n (ID number of record to delete)! Try again, or try -h for help.")
      f.Exit(db, 1)
    }
  }

  // Handles command line way to list records. 
  // It checks all the flags, and if they have been used, it adds them to "recordStrings". 
  // At the end, it takes all of those strings and creates a database query and then
  // sends that query to the fetchRecord function. 
  if list {
    if custom != "" {
      fmt.Println("Date: ", timeStr)
      record, err := db.Query(custom)
      d.FetchRecord(db, record, err)
      f.Exit(db, 0)
    }

    // recordStrings collects the sql phrases for each different flag. 
    var recordStrings []string

    // Used to order by date
    var dateOrder string

    // This is the beginning of all queries to the database. I always want every column 
    // returned. So if no options are set, this is sent to fetchRecords all by itself.
    // Otherwise, everything else is added onto this string.
    baseString := "SELECT * FROM weather"

    if date != "" {
      dateString := fmt.Sprint("date='"+date+"'")
      recordStrings = append(recordStrings, dateString)
    }

    if today {
      todayString := fmt.Sprint("strftime('%Y-%m-%d', date)='" + todaysDate + "'")
      recordStrings = append(recordStrings, todayString)
    }

    // year is -year flag
    if year != "" {
      contains := strings.Contains(year, "-")

      // This handles if you have a range of years. must be written as i.e. 2010-2015
      if contains {
        years := strings.Split(year, "-")
        // Lets the user know that the year must be 4 digits, instead of just returning an empty database.
        if !f.CheckYear(years[0]) || !f.CheckYear(years[1]) {
          fmt.Println("\nYour year appears to be entered wrong. Make sure year contains exactly 4 digits. ie 2022\n")
          f.Exit(db, 1)
        }
        yearString := "(strftime('%Y', date) between '" + string(years[0]) + "' and '" + string(years[1]) + "')"
        recordStrings = append(recordStrings, yearString)
      // This handles single year 
      } else {
        // Lets the user know that the year must be 4 digits, instead of just returning an empty database.
        if !f.CheckYear(year) {
          fmt.Println("\nYour year appears to be entered wrong. Make sure year contains exactly 4 digits. ie 2022\n")
          f.Exit(db, 1)
        }
        yearString := fmt.Sprint("strftime('%Y', date)='" + year + "'")
        recordStrings = append(recordStrings, yearString)
      }
    }


    if currentMonth && month == "" {
      splitDate := strings.Split(todaysDate, "-")
      currentMonthString := "strftime('%m', date)='" + splitDate[1] + "'"
      recordStrings = append(recordStrings, currentMonthString)
    } else if currentMonth && month != "" {
      fmt.Println("Can't use -m and -month together!")
      f.Exit(db, 1)
    }

    // month is -month flag
    if month != "" {
      contains := strings.Contains(month, "-")

      // This handles if you have a range of months. must be written as i.e. 05-10
      if contains {
        months := strings.Split(month, "-")
        // Lets the user know that the month requires a leading 0, instead of just returning an empty database.
        // CheckMonth checks if it is a valid month, and returns true. So if !false (true), it warns the user
        // the month is incorrect.
        if !f.CheckMonth(months[0]) || !f.CheckMonth(months[1]) {
          fmt.Println("\nYour month appears to be wrong. Make sure each month is exactly 2 digits, and between 01-12. If it's a single digit month, add a leading zero, ie 05.\n")
          f.Exit(db, 1)
        }
        monthString := "(strftime('%m', date) between '" + string(months[0]) + "' and '" + string(months[1]) + "')"
        recordStrings = append(recordStrings, monthString)
      // This handles single month
      } else {
        // Lets the user know that the month requires a leading 0, instead of just returning an empty database.
        //if len(month) != 2 {
        if !f.CheckMonth(month) {
          fmt.Println("\nYour month appears to be wrong. Make sure each month is exactly 2 digits, and between 01-12. If it's a single digit month, add a leading zero, ie 05.\n")
          f.Exit(db, 1)
        }
        monthString := fmt.Sprint("strftime('%m', date)='" + month + "'")
        recordStrings = append(recordStrings, monthString)
      }
    }

    // day is -day flag
    if day != "" {
      contains := strings.Contains(day, "-")

      // This handles if you have a range of days. must be written as i.e. 05-10
      if contains {
        days := strings.Split(day, "-")
        // Lets the user know that the day requires a leading 0, instead of just returning an empty database.
        if !f.CheckDay(days[0]) || !f.CheckDay(days[1]) {
          fmt.Println("\nYour day appears to be wrong. Make sure each day is exactly 2 digits, and between 01-31. If it's a single digit day, add a leading zero, ie 05.\n")
          f.Exit(db, 1)
        }
        dayString := "(strftime('%d', date) between '" + string(days[0]) + "' and '" + string(days[1]) + "')"
        recordStrings = append(recordStrings, dayString)
      // This handles single day
      } else {
        // Lets the user know that the day requires a leading 0, instead of just returning an empty database.
        if !f.CheckDay(day) {
          fmt.Println("\nYour day appears to be wrong. Make sure each day is exactly 2 digits, and between 01-31. If it's a single digit day, add a leading zero, ie 05.\n")
          f.Exit(db, 1)
        }
        dayString := fmt.Sprint("strftime('%d', date)='" + day + "'")
        recordStrings = append(recordStrings, dayString)
      }
    }

    // Select from this date to current date.
    if dateFrom != "" {
      if !f.CheckDate(dateFrom) {
        fmt.Println("Error:")
        fmt.Println("\nIt seems your date isn't the proper format. Please enter date as YYYY-MM-DD ie 2022-01-12\n")
        f.Exit(db, 1)
      }

      dateFromString := "(strftime('%Y-%m-%d', date) between '" + dateFrom + "' and '" + timeStr + "')"
      recordStrings = append(recordStrings, dateFromString)
    }

    // Oders by date either ascending or descending 
    if dateOldToNew && !dateNewToOld{
      dateOrder = " ORDER BY date ASC"
    } else if dateNewToOld && !dateOldToNew{
      dateOrder = " ORDER BY date DESC"
    } else if dateNewToOld && dateOldToNew {
      fmt.Println("Error:\nYou can't use both dateoton and datentoo. Conflict order by ascending and descending.")
      f.Exit(db, 1)
    } else {
      dateOrder = ""
    }

    // This is the area that puts the sql phrase together and sends it to the fetchRecords
    // function.
    // I set it up to pay attention to three scenarios: No additional phrase, 1 additional
    // phrase, or more than one additional phrase.
    // The phrases are stored in the slice recordStrings
    // If no additional phrases were set, ie no flags were used, sends only the baseString,
    // which returns the entire database.
    // Additionally I have added the ability to order by date, either ascending or desending.
    // I needed this because sometimes you add a record after the date, and they appear
    // out of order. To do it, the program checks if the flag is set to order by date,
    // and then just tags it on to the end of the sql query. If the flags aren't set,
    // it just tags on an empty string so nothing changes.
    if len(recordStrings) == 0 {
      fmt.Println("Date: ", timeStr)
      fullString := fmt.Sprint(baseString + dateOrder)
      if showSql {
        fmt.Println("SQL Query:", fullString)
      }
      record, err := db.Query(fullString)
      d.FetchRecord(db, record, err)
      f.Exit(db, 0)
    // If there is one additional phrase, it appends WHERE and the phrase to base string,
    } else if len(recordStrings) == 1 {
      fmt.Println("Date: ", timeStr)
      fullString := fmt.Sprint(baseString + " WHERE " + recordStrings[0] + dateOrder)
      if showSql {
        fmt.Println("SQL Query:", fullString)
      }
      record, err := db.Query(fullString)
      d.FetchRecord(db, record, err)
      f.Exit(db, 0)
    // If there are more than one phrase to add, first it combines them with AND, and 
    // then adds that to baseString, with the connecting WHERE as well.
    } else if len(recordStrings) > 1 {
      fmt.Println("Date: ", timeStr)
      combineStrings := strings.Join(recordStrings, " AND ")
      fullString := fmt.Sprint(baseString + " WHERE " + combineStrings + dateOrder)
      if showSql {
        fmt.Println("SQL Query:", fullString)
      }
      record, err := db.Query(fullString)
      d.FetchRecord(db, record, err)
      f.Exit(db, 0)
    }
  }

  // Handles the github push command.
  if push {
    // git add --all
    cmd, stdout := exec.Command("git", "add", "--all"), new(strings.Builder)
    cmd.Dir = dbDir
    cmd.Stdout = stdout
    err := cmd.Run()
    if err != nil {
      fmt.Println("Error executing git add --all:\n", err)
      f.Exit(db, 1)
    }
    fmt.Println(stdout.String())

    // git commit -m 'update weather database'
    cmd, stdout = exec.Command("git", "commit", "-m", "'update weather database'"), new(strings.Builder)
    cmd.Dir = dbDir
    cmd.Stdout = stdout
    err = cmd.Run()
    if err != nil {
      fmt.Println("Error executing git commit -m 'update weather database':\n", err)
      f.Exit(db, 1)
    }
    fmt.Println(stdout.String())

    // git push
    cmd, stdout, stderr := exec.Command("git", "push"), new(strings.Builder), new(strings.Builder)
    cmd.Dir = dbDir
    cmd.Stdout = stdout
    cmd.Stderr = stderr
    err = cmd.Run()
    if err != nil {
      fmt.Println("Error executing git push:\n", err)
      f.Exit(db, 1)
    }
    fmt.Println(stdout.String())
    fmt.Println(stderr.String())

    // Unsatisfactory confirmation message
    fmt.Println("You probably pushed it to git...")
    // Exit
    f.Exit(db, 0)
  }

  // Handles the github pull command.
  if pull {
    // git pull 
    cmd, stdout := exec.Command("git", "pull"), new(strings.Builder)
    cmd.Dir = dbDir
    cmd.Stdout = stdout
    err := cmd.Run()
    if err != nil {
      fmt.Println("Error executing git pull:\n", err)
      f.Exit(db, 1)
    }
    fmt.Println(stdout.String())

    // exit
    f.Exit(db, 0)
  }

  // Handles the github status command.
  if status {
    // git status 
    cmd, stdout := exec.Command("git", "status"), new(strings.Builder)
    cmd.Dir = dbDir
    cmd.Stdout = stdout
    err := cmd.Run()
    if err != nil {
      fmt.Println("Error executing git status:\n", err)
      f.Exit(db, 1)
    }
    fmt.Println(stdout.String())

    // exit
    f.Exit(db, 0)
  }

  // This runs if no arguments are specified.
  fmt.Printf("%s: Try running with -h for usage\n", os.Args[0])
}
