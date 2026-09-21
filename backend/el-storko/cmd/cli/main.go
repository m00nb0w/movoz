package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"el-storko/internal/clicmd"
)

func apiBaseURL() string {
	if v := os.Getenv("EL_STORKO_API_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		runAdd(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  el-storko-cli add "<title>" [--epic <id>] [--description "..."]
  el-storko-cli list [--source ...] [--status ...] [--epic <id>]`)
}

func runAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	epic := fs.Int64("epic", 0, "parent epic id")
	description := fs.String("description", "", "task description")
	fs.Parse(args)

	positional := fs.Args()
	if len(positional) < 1 {
		fmt.Println("error: add requires a title")
		os.Exit(1)
	}
	title := positional[0]

	var epicID *int64
	if fs.Lookup("epic").Value.String() != "0" {
		epicID = epic
	}

	item, err := clicmd.Add(apiBaseURL(), title, epicID, *description)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("Created task #%d: %s\n", item.ID, item.Title)
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	source := fs.String("source", "", "filter by source")
	status := fs.String("status", "", "filter by status")
	epic := fs.Int64("epic", 0, "filter by epic id")
	fs.Parse(args)

	opts := clicmd.ListOptions{Source: *source, Status: *status}
	if fs.Lookup("epic").Value.String() != "0" {
		opts.EpicID = epic
	}

	items, err := clicmd.List(apiBaseURL(), opts)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	for _, item := range items {
		parent := "-"
		if item.ParentID != nil {
			parent = strconv.FormatInt(*item.ParentID, 10)
		}
		fmt.Printf("#%d [%s] %-6s %-12s parent=%s %s\n", item.ID, item.Type, item.Status, item.Source, parent, item.Title)
	}
}
