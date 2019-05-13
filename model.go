package kingpin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Data model for Kingpin command-line structure.

var (
	ignoreInCount = map[string]bool{
		"help":                   true,
		"help-long":              true,
		"help-man":               true,
		"completion-bash":        true,
		"completion-script-bash": true,
		"completion-script-zsh":  true,
	}
)

// FlagGroupModel represents a read only value of a flagGroup.
type FlagGroupModel struct {
	Flags []*FlagModel
}

// FlagSummary returns a summary string for all flags in a flag group.
func (f *FlagGroupModel) FlagSummary() string {
	out := []string{}
	count := 0

	for _, flag := range f.Flags {
		if !ignoreInCount[flag.Name] {
			count++
		}

		if flag.Required {
			if flag.IsBoolFlag() {
				out = append(out, fmt.Sprintf("--[no-]%s", flag.Name))
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.Name, flag.FormatPlaceHolder()))
			}
		}
	}
	if count != len(out) {
		out = append(out, "[<flags>]")
	}
	return strings.Join(out, " ")
}

// FlagModel represents a read only value of an FlagClause.
type FlagModel struct {
	Name            string
	Help            string
	Short           rune
	Default         []string
	Envar           string
	Aliases         []string
	NegativeAliases []string
	PlaceHolder     string
	Required        bool
	Hidden          bool
	Value           Value
}

func (f *FlagModel) String() string {
	return f.Value.String()
}

// IsBoolFlag determines if the current FlagModel is a switch.
func (f *FlagModel) IsBoolFlag() bool {
	if fl, ok := f.Value.(boolFlag); ok {
		return fl.IsBoolFlag()
	}
	return false
}

// FormatPlaceHolder returns a string representing the place holder for the value associated to the flag.
func (f *FlagModel) FormatPlaceHolder() string {
	if f.PlaceHolder != "" {
		return f.PlaceHolder
	}
	if len(f.Default) > 0 {
		ellipsis := ""
		if len(f.Default) > 1 {
			ellipsis = "..."
		}
		if _, ok := f.Value.(*stringValue); ok {
			return strconv.Quote(f.Default[0]) + ellipsis
		}
		return f.Default[0] + ellipsis
	}
	return strings.ToUpper(f.Name)
}

// ArgGroupModel returns a read only value of an argument group.
type ArgGroupModel struct {
	Args []*ArgModel
}

// ArgSummary returns the summary string representing the arguments.
func (a *ArgGroupModel) ArgSummary() string {
	depth := 0
	out := []string{}
	for _, arg := range a.Args {
		h := "<" + arg.Name + ">"
		if !arg.Required {
			h = "[" + h
			depth++
		}
		out = append(out, h)
	}
	out[len(out)-1] = out[len(out)-1] + strings.Repeat("]", depth)
	return strings.Join(out, " ")
}

// ArgModel represents a read only value of an argument clause.
type ArgModel struct {
	Name     string
	Help     string
	Default  []string
	Envar    string
	Required bool
	Value    Value
}

func (a *ArgModel) String() string {
	return a.Value.String()
}

// CmdGroupModel represents a read only value of a command group.
type CmdGroupModel struct {
	Commands []*CmdModel
}

// FlattenedCommands returns the list of command model (handling recursive sub command definition).
func (c *CmdGroupModel) FlattenedCommands() (out []*CmdModel) {
	for _, cmd := range c.Commands {
		if len(cmd.Commands) == 0 {
			out = append(out, cmd)
		}
		out = append(out, cmd.FlattenedCommands()...)
	}
	return
}

// CmdModel represents a read only value of an command.
type CmdModel struct {
	Name        string
	Aliases     []string
	Help        string
	FullCommand string
	Depth       int
	Hidden      bool
	Default     bool
	*FlagGroupModel
	*ArgGroupModel
	*CmdGroupModel
}

func (c *CmdModel) String() string {
	return c.FullCommand
}

// ApplicationModel represents a read only value of an application.
type ApplicationModel struct {
	Name    string
	Help    string
	Version string
	Author  string
	*ArgGroupModel
	*CmdGroupModel
	*FlagGroupModel
}

// Model returns a read only value of an application.
func (a *Application) Model() *ApplicationModel {
	return &ApplicationModel{
		Name:           a.Name,
		Help:           a.Help,
		Version:        a.version,
		Author:         a.author,
		FlagGroupModel: a.flagGroup.Model(),
		ArgGroupModel:  a.argGroup.Model(),
		CmdGroupModel:  a.cmdGroup.Model(),
	}
}

func (a *argGroup) Model() *ArgGroupModel {
	m := &ArgGroupModel{}
	for _, arg := range a.args {
		m.Args = append(m.Args, arg.Model())
	}
	return m
}

// Model returns a read only value of an argument clause.
func (a *ArgClause) Model() *ArgModel {
	return &ArgModel{
		Name:     a.name,
		Help:     a.help,
		Default:  a.defaultValues,
		Envar:    a.envar,
		Required: a.required,
		Value:    a.value,
	}
}

func (f *flagGroup) Model() *FlagGroupModel {
	m := &FlagGroupModel{}
	for _, fl := range f.flagOrder {
		m.Flags = append(m.Flags, fl.Model())
	}
	return m
}

// Model returns a read only value of a Flag.
func (f *FlagClause) Model() *FlagModel {
	var aliases, negatives []string
	for alias, kind := range f.aliases {
		if kind == aliasNegative {
			negatives = append(negatives, alias)
		} else {
			aliases = append(aliases, alias)
		}
	}
	sort.Strings(aliases)
	sort.Strings(negatives)
	return &FlagModel{
		Name:            f.name,
		Help:            f.help,
		Short:           rune(f.shorthand),
		Default:         f.defaultValues,
		Envar:           f.envar,
		PlaceHolder:     f.placeholder,
		Aliases:         aliases,
		NegativeAliases: negatives,
		Required:        f.required,
		Hidden:          f.hidden,
		Value:           f.value,
	}
}

func (c *cmdGroup) Model() *CmdGroupModel {
	m := &CmdGroupModel{}
	for _, cm := range c.commandOrder {
		m.Commands = append(m.Commands, cm.Model())
	}
	return m
}

// Model returns a read only value of a Command.
func (c *CmdClause) Model() *CmdModel {
	depth := 0
	for i := c; i != nil; i = i.parent {
		depth++
	}
	return &CmdModel{
		Name:           c.name,
		Aliases:        c.aliases,
		Help:           c.help,
		Depth:          depth,
		Hidden:         c.hidden,
		Default:        c.isDefault,
		FullCommand:    c.FullCommand(),
		FlagGroupModel: c.flagGroup.Model(),
		ArgGroupModel:  c.argGroup.Model(),
		CmdGroupModel:  c.cmdGroup.Model(),
	}
}
