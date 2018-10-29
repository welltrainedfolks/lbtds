// Flagger - arbitrary CLI flags parser.
//
// Copyright (c) 2017-2018, Stanislav N. aka pztrn.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject
// to the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package flagger

import (
	// stdlib
	"errors"
	"flag"
	"sync"
)

// Flagger implements (kinda) extended CLI parameters parser. As it
// available from CommonContext, these flags will be available to
// whole application.
//
// It uses reflection to determine what kind of variable we should
// parse or get.
type Flagger struct {
	// Flags that was added by user.
	flags      map[string]*Flag
	flagsMutex sync.Mutex

	// Flags that will be passed to flag module.
	flagsBool   map[string]*bool
	flagsInt    map[string]*int
	flagsString map[string]*string
}

// Adds flag to list of flags we will pass to ``flag`` package.
func (f *Flagger) AddFlag(flag *Flag) error {
	_, present := f.flags[flag.Name]
	if present {
		logger.Fatal("Cannot add flag '" + flag.Name + "' - already added!")
		return errors.New("Cannot add flag '" + flag.Name + "' - already added!")
	}

	f.flags[flag.Name] = flag
	return nil
}

// This function returns boolean value for flag with given name.
// Returns bool value for flag and nil as error on success
// and false bool plus error with text on error.
func (f *Flagger) GetBoolValue(name string) (bool, error) {
	fl, present := f.flagsBool[name]
	if !present {
		return false, errors.New("No such flag: " + name)
	}
	return (*fl), nil
}

// This function returns integer value for flag with given name.
// Returns integer on success and 0 on error.
func (f *Flagger) GetIntValue(name string) (int, error) {
	fl, present := f.flagsInt[name]
	if !present {
		return 0, errors.New("No such flag: " + name)
	}
	return (*fl), nil
}

// This function returns string value for flag with given name.
// Returns string on success or empty string on error.
func (f *Flagger) GetStringValue(name string) (string, error) {
	fl, present := f.flagsString[name]
	if !present {
		return "", errors.New("No such flag: " + name)
	}
	return (*fl), nil
}

// Flagger initialization.
func (f *Flagger) Initialize() {
	logger.Print("Initializing CLI parameters parser...")

	f.flags = make(map[string]*Flag)

	f.flagsBool = make(map[string]*bool)
	f.flagsInt = make(map[string]*int)
	f.flagsString = make(map[string]*string)
}

// This function adds flags from flags map to flag package and parse
// them. They can be obtained later by calling GetTYPEValue(name),
// where TYPE is one of Bool, Int, String.
func (f *Flagger) Parse() {
	for name, fl := range f.flags {
		if fl.Type == "bool" {
			fdef := fl.DefaultValue.(bool)
			f.flagsBool[name] = &fdef
			flag.BoolVar(&fdef, name, fdef, fl.Description)
		} else if fl.Type == "int" {
			fdef := fl.DefaultValue.(int)
			f.flagsInt[name] = &fdef
			flag.IntVar(&fdef, name, fdef, fl.Description)
		} else if fl.Type == "string" {
			fdef := fl.DefaultValue.(string)
			f.flagsString[name] = &fdef
			flag.StringVar(&fdef, name, fdef, fl.Description)
		}
	}

	logger.Print("Parsing CLI parameters...")
	flag.Parse()
}
