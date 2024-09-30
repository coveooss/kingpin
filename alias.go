package kingpin

import (
	"fmt"
	"strings"
)

type aliasMixin struct {
	aliases      map[string]aliasKind
	autoShortcut *bool // Only set if explicitly configured
}

type flagAlias struct {
	*FlagClause
	kind aliasKind
}

type aliasKind byte

const (
	aliasNone aliasKind = iota
	aliasName
	aliasNegative
	aliasShortcut
)

// Alias defines alias name that could be used instead of the long name.
func (f *FlagClause) Alias(aliases ...string) *FlagClause {
	if aliases == nil {
		f.aliases = nil // If supplied alias is nil, we clear the previously defined aliases
	}
	for _, alias := range aliases {
		f.addAlias(alias, aliasName)
	}
	return f
}

// AutoShortcut enables automatic shortcut for this flag (overriding the flag group setting).
func (f *FlagClause) AutoShortcut() *FlagClause { return f.setAutoShortcut(true) }

// NoAutoShortcut disables automatic shortcut for this flag (overriding the flag group setting).
func (f *FlagClause) NoAutoShortcut() *FlagClause { return f.setAutoShortcut(false) }

func (f *FlagClause) setAutoShortcut(value bool) *FlagClause {
	f.autoShortcut = &value
	return f
}

// Add an alias to the selected flag.
func (f *FlagClause) addAlias(alias string, kind aliasKind) error {
	if f.aliases == nil {
		f.aliases = make(map[string]aliasKind)
	}
	if current, exist := f.aliases[alias]; exist && current != kind {
		return fmt.Errorf("Alias %s already exist", alias)
	}
	f.aliases[alias] = kind
	return nil
}

// This function find the corresponding flag either by name or though aliases.
// If the resulted flag correspond to a negative alias (-no-boolOption), invert is set to true
func (fg *flagGroup) getFlagAlias(name string) (flag *FlagClause, invert bool, err error) {
	if err = fg.ensureAliases(); err != nil {
		return
	} else if flag = fg.long[name]; flag != nil {
		return
	} else if alias := fg.aliases[name]; alias.kind != aliasNone {
		flag = alias.FlagClause
		invert = alias.kind == aliasNegative
	}
	return
}

// Clear the aliases and regenerate it.
// Used when the ParseContext change to a sub command.
func (fg *flagGroup) resetAliases() error {
	fg.aliases = nil
	return fg.ensureAliases()
}

// Ensure that the aliases are evaluated.
// Called during the parsing since we do not know the nature of flags until we launch the parsing.
func (fg *flagGroup) ensureAliases() error {
	if fg.aliases != nil {
		return nil
	}
	// The alias map is not yet initialized, so we do it
	fg.aliases = make(map[string]flagAlias)

	for _, flag := range fg.flagOrder {
		if err := fg.addShortcut(flag.name, flag); err != nil {
			return err
		}
		if err := fg.addNegativeAlias(flag.name, flag, aliasNone); err != nil {
			return err
		}
		for alias, kind := range flag.aliases {
			if kind != aliasName {
				continue
			}
			if err := fg.addGroupAlias(alias, flag, kind); err != nil {
				return err
			}
			if err := fg.addShortcut(alias, flag); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add an alias to the current flag group and return an error if the alias conflict with another flag.
func (fg *flagGroup) addGroupAlias(name string, flag *FlagClause, kind aliasKind) error {
	if existing := fg.long[name]; existing != nil {
		return aliasErrorf("Alias %s on %s is already associated to flag %s", name, flag.name, existing.name)
	}
	if alias := fg.aliases[name]; alias.kind != aliasNone && (alias.kind != kind || alias.FlagClause != flag) {
		return aliasErrorf("Alias %s on %s is already associated to flag %s", name, flag.name, alias.name)
	}
	if err := fg.addNegativeAlias(name, flag, kind); err != nil {
		return err
	}

	fg.aliases[name] = flagAlias{FlagClause: flag, kind: kind}
	if err := flag.addAlias(name, kind); err != nil {
		return aliasErrorf("Unable to add alias %s to %s: %v", name, flag.name, err)
	}
	return nil
}

func (fg *flagGroup) addShortcut(name string, flag *FlagClause) error {
	if flag.autoShortcut == nil {
		flag.autoShortcut = &fg.autoShortcut
	}
	if !*flag.autoShortcut || flag.shorthand != 0 && !strings.Contains(name, "-") {
		// We do not add single letter shortcut for flag that already have a shorthand
		return nil
	}

	var shortcut string
	for _, word := range strings.Split(name, "-") {
		shortcut += string(word[0])
	}
	if err := fg.addGroupAlias(shortcut, flag, aliasShortcut); err != nil {
		return err
	}
	return fg.addNegativeAlias(shortcut, flag, aliasShortcut)
}

// If the flag is a boolFlag, add its negative counterpart.
func (fg *flagGroup) addNegativeAlias(name string, flag *FlagClause, kind aliasKind) error {
	if fb, isSwitch := flag.value.(boolFlag); kind != aliasNegative && isSwitch && fb.IsBoolFlag() {
		if err := fg.addGroupAlias("no-"+name, flag, aliasNegative); err != nil {
			return err
		}
		if len(name) <= 3 && !strings.Contains(name, "-") {
			// For short single word name, we also negative form simply prefixed by a n
			if err := fg.addGroupAlias("n"+name, flag, aliasNegative); err != nil {
				return err
			}
		}
	}
	return nil
}

type aliasError string

func (ae aliasError) Error() string { return string(ae) }
func aliasErrorf(format string, args ...interface{}) aliasError {
	return aliasError(fmt.Sprintf(format, args...))
}
