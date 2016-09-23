// Copyright 2016 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package generator deals with creating projects from template files.
package generator

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	// DefinitionFile is the file containing the generator definition within a generator.
	DefinitionFile = "generator.json"

	// PartialsDir is the directory containing partials within a generator.
	PartialsDir = "partials"

	// FilesDir is the directory containing files within a generator.
	FilesDir = "files"
)

// Definition contains properties for a template generator definition.
type Definition struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Variables   map[string]interface{} `json:"variables"`
	Inputs      InputMap               `json:"inputs"`
	Priorities  []string               `json:"priorities"`
}

// NewDefinitionFromFile loads a generator from a file.
// The file is treated as a template and is fed the given variables and functions.
// If no functions are given, DefaultDefinitionFuncs is used.
func NewDefinitionFromFile(path string, vars map[string]interface{}, funcs template.FuncMap) (*Definition, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tmpl := template.New("generator")
	if funcs == nil {
		tmpl.Funcs(DefaultDefinitionFuncs())
	} else {
		tmpl.Funcs(funcs)
	}
	if _, err := tmpl.Parse(string(b)); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, err
	}
	var gen Definition
	if err := json.Unmarshal(buf.Bytes(), &gen); err != nil {
		return nil, err
	}
	return &gen, nil
}

// DefaultDefinitionFuncs return the default function map used when parsing a generator definition.
// It adds the following functions:
// - now(format): returns a formatted representation of the current date
// - nowUTC(format): returns a formatted representation of the current UTC date
// - secret(length): returns a random secret string
func DefaultDefinitionFuncs() template.FuncMap {
	return template.FuncMap{
		"now":    now,
		"nowUTC": nowUTC,
		"secret": secret,
	}
}

// Options contains options for a generator.
type Options struct {
	// Variables for the definition file.
	DefVars map[string]interface{}

	// Functions for the definition file.
	DefFuncs template.FuncMap

	// Variables for the templates.
	TmplVars map[string]interface{}

	// Functions for the templates.
	TmplFuncs template.FuncMap

	// A reader for user input, default to stdin.
	Reader io.Reader
}

// Generator deals with parsing templates, handling user input, and outputing processed templates.
type Generator struct {
	opts     *Options
	src      string
	def      *Definition
	partials *template.Template
	files    *template.Template
	values   map[string]interface{}
	reader   *bufio.Reader
}

// NewFromDir create a new generator from a directory.
func NewFromDir(src string, opts *Options) (*Generator, error) {
	defFile := filepath.Join(src, DefinitionFile)
	funcs := extendFuncs(DefaultDefinitionFuncs(), opts.DefFuncs)
	def, err := NewDefinitionFromFile(defFile, opts.DefVars, funcs)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if opts.Reader != nil {
		reader = opts.Reader
	} else {
		reader = os.Stdin
	}
	return &Generator{
		opts:   opts,
		src:    src,
		def:    def,
		values: map[string]interface{}{},
		reader: bufio.NewReader(reader),
	}, nil
}

// DefaultTmplFuncs return the default function map used when parsing a template
// It adds the following functions:
// - ask(json): creates an input on-the-fly and returns its value
// - input(id): returns the value of an input
// - now(format): returns a formatted representation of the current date
// - nowUTC(format): returns a formatted representation of the current UTC date
// - partial(path, [vars]): executes the partial with given name and variables (path relative to partials folder)
// - secret(length): returns a random secret string
func (gen *Generator) DefaultTmplFuncs() template.FuncMap {
	return template.FuncMap{
		"ask":     gen.ask,
		"input":   gen.input,
		"now":     now,
		"nowUTC":  nowUTC,
		"partial": gen.execPartial,
		"secret":  secret,
	}
}

// Exec parses templates, handles user input, and outputs processed templates to given dir.
func (gen *Generator) Exec(dst string) error {
	if err := gen.parsePartials(); err != nil {
		return err
	}
	if err := gen.parseFiles(); err != nil {
		return err
	}
	if err := gen.generate(dst); err != nil {
		return err
	}
	return nil
}

func (gen *Generator) parsePartials() error {
	gen.partials = template.New("partials")
	dir := filepath.Join(gen.src, PartialsDir)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	gen.partials.Funcs(extendFuncs(gen.DefaultTmplFuncs(), gen.opts.TmplFuncs))
	if err := walkTmpl(dir, dir, gen.partials); err != nil {
		return err
	}
	return nil
}

func (gen *Generator) parseFiles() error {
	gen.files = template.New("files")
	dir := filepath.Join(gen.src, FilesDir)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	gen.files.Funcs(extendFuncs(gen.DefaultTmplFuncs(), gen.opts.TmplFuncs))
	if err := walkTmpl(dir, dir, gen.files); err != nil {
		return err
	}
	return nil
}

func (gen *Generator) execPartial(name string, opts ...interface{}) (string, error) {
	l := len(opts)
	if l > 1 {
		return "", errors.New("too many arguments")
	}
	var vars interface{}
	if l == 1 {
		vars = opts[0]
	}
	var buf bytes.Buffer
	if err := gen.partials.ExecuteTemplate(&buf, name, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type tmplDesc struct {
	tmpl     *template.Template
	priority int
}

type tmplDescSlice []tmplDesc

func (s tmplDescSlice) Len() int {
	return len(s)
}

func (s tmplDescSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s tmplDescSlice) Less(i, j int) bool {
	s1, s2 := s[i], s[j]
	p1, p2 := s1.priority, s2.priority
	if p1 == p2 {
		return s1.tmpl.Name() < s2.tmpl.Name()
	}
	if p1 == -1 {
		return false
	}
	if p2 == -1 {
		return true
	}
	return p1 < p2
}

func (gen *Generator) generate(dst string) error {
	var descs tmplDescSlice
	for _, tmpl := range gen.files.Templates() {
		name := tmpl.Name()
		priority := -1
		for i, v := range gen.def.Priorities {
			if v == name {
				priority = i
				break
			}
		}
		descs = append(descs, tmplDesc{
			tmpl:     tmpl,
			priority: priority,
		})
	}
	sort.Sort(descs)

	for _, desc := range descs {
		tmpl := desc.tmpl
		name := tmpl.Name()
		in := filepath.Join(gen.src, FilesDir, name)
		info, err := os.Stat(in)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, name)
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		vars := map[string]interface{}{}
		for k, v := range gen.opts.TmplVars {
			vars[k] = v
		}
		for k, v := range gen.def.Variables {
			vars[k] = v
		}
		if err := tmpl.Execute(f, vars); err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) input(id string) (interface{}, error) {
	val, ok := gen.values[id]
	if ok {
		return val, nil
	}
	for k, in := range gen.def.Inputs {
		if k == id {
			val, err := gen.read(in)
			if err != nil {
				return nil, err
			}
			gen.values[id] = val
			return val, nil
		}
	}
	return nil, fmt.Errorf("undefined input %q", id)
}

func (gen *Generator) ask(input string) (interface{}, error) {
	in, err := UnmarshalJSONInput([]byte(input))
	if err != nil {
		return nil, err
	}
	return gen.read(in)
}

func (gen *Generator) read(in Input) (interface{}, error) {
	fmt.Print(in.Msg())
	for {
		fmt.Print("? ")
		str, err := gen.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		str = strings.TrimSpace(str)
		if err := in.Set(str); err != nil {
			fmt.Println(err)
			continue
		}
		return in.Get(), nil
	}
}

func walkTmpl(base, dir string, tmpl *template.Template) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := walkTmpl(base, file, tmpl); err != nil {
				return err
			}
			continue
		}
		name, err := filepath.Rel(base, file)
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		t := tmpl.New(name)
		if _, err := t.Parse(string(b)); err != nil {
			return err
		}
	}
	return nil
}

func now(format string) string {
	return time.Now().Format(format)
}

func nowUTC(format string) string {
	return time.Now().UTC().Format(format)
}

func extendFuncs(maps ...template.FuncMap) template.FuncMap {
	funcs := template.FuncMap{}
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				funcs[k] = v
			}
		}
	}
	return funcs
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&()[]*+-,./:;<=>?^_{}~")

func secret(n int) (string, error) {
	if n < 0 {
		return "", errors.New("size must not be negative")
	}
	r := make([]rune, n)
	max := big.NewInt(int64(len(letters)))
	for i := range r {
		j, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		r[i] = letters[int(j.Int64())]
	}
	return string(r), nil
}
