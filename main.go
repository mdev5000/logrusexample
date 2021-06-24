package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type NestedThing struct {
	NestedField1 int
	NestedThing2 string
}

type Thing struct {
	SomeField string
	Nested    *NestedThing
}

type LogEntry struct {
	Msg          string
	Err          string
	Stack        string
	Thing        string
	ThingEncoded string
}

func unglob(data []byte, i interface{}) error {
	enc := gob.NewDecoder(bytes.NewBuffer(data))
	return enc.Decode(i)
}

func gobIt(i interface{}) ([]byte, error) {
	gobb := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(gobb)
	err := enc.Encode(i)
	return gobb.Bytes(), err
}

func logIt(log *logrus.Logger, msg string, errAndStack error, thing *Thing) {
	thingEncoded, err := gobIt(thing)
	if err != nil {
		log.Errorf("error while encoding Thing: %s", err)
	}
	log.WithFields(logrus.Fields{
		"err":          errAndStack.Error(),
		"stack":        fmt.Sprintf("%+v", errAndStack),
		"thing":        spew.Sdump(thing),
		"thingEncoded": base64.StdEncoding.EncodeToString(thingEncoded),
	}).Error(msg)
}

func logWhenSomethingBadHappens(log *logrus.Logger, thing *Thing) {
	err := errors.New("something happened")
	logIt(log, "something bad happened", errors.WithStack(err), thing)
}

func main() {
	logOut := bytes.NewBuffer(nil)
	log := logrus.New()
	log.SetOutput(logOut)
	log.SetFormatter(&logrus.JSONFormatter{})

	t := Thing{
		SomeField: "a value",
		Nested: &NestedThing{
			NestedField1: 10,
			NestedThing2: "some value",
		},
	}

	// Log some stuff.
	logWhenSomethingBadHappens(log, &t)
	logWhenSomethingBadHappens(log, &t)


	// Print the log output.
	fmt.Println("LOG DATA: ---------------------------------------------------------")
	fmt.Println("")
	logBytes := logOut.Bytes()
	fmt.Println(string(logBytes))

	// Extract information from the first log entry.
	fmt.Println("ENTRY Info: ---------------------------------------------------------")
	var entryJson []byte
	lines := bytes.Split(logBytes, []byte("\n"))
	entryJson = lines[0]
	var logEntry LogEntry
	if err := json.Unmarshal(entryJson, &logEntry); err != nil {
		panic(err)
	}
	fmt.Println("")

	fmt.Println("Log Message:", logEntry.Msg)
	fmt.Println("")

	// Debug the string output of the Thing we passed.
	//
	// You could also use fmt.Sprintf("+%v", ...) or whatever you preference is, I just find that spew output is a bit
	// easier to read.
	fmt.Println("String dump of thing:")
	fmt.Println("")
	fmt.Println(logEntry.Thing)
	fmt.Println("")

	// You can decode the globed object we encoded and manipulate it via code.
	fmt.Println("Decode globed object:")
	fmt.Println("")
	globBytes, err := base64.StdEncoding.DecodeString(logEntry.ThingEncoded)
	if err != nil {
		panic(err)
	}
	var entryThing Thing
	if err := unglob(globBytes, &entryThing); err != nil {
		panic(err)
	}
	fmt.Println("NestedField1:", entryThing.Nested.NestedField1)
	fmt.Println("")
	fmt.Println("Nested:")
	spew.Dump(entryThing.Nested)

	// And print your saved stacktrace information.
	fmt.Println("")
	fmt.Println("Getting stacktrace data:")
	fmt.Println("")
	fmt.Println(logEntry.Stack)
}
