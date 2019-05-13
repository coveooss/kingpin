package kingpin

import (
	"fmt"
)

type aliasMixin struct {
	aliases map[string]aliasKind
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
)

// Alias defines alias name that could be used instead of the long name.
func (f *FlagClause) Alias(aliases ...string) *FlagClause {
	if aliases == nil {
		// If supplied alias is nil, we clear the previously defined aliases
		f.aliases = nil
	}
	for _, alias := range aliases {
		f.addAlias(alias, aliasName)
	}
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

type aliasGroupMixin struct{}

// This function find the corresponding flag either by name or though aliases.
// If the resulted flag correspond to a negative alias (-no-boolOption), invert is set to true
func (f *flagGroup) getFlagAlias(name string) (flag *FlagClause, invert bool, err error) {
	if err = f.ensureAliases(); err != nil {
		return
	} else if flag = f.long[name]; flag != nil {
		return
	} else if alias := f.aliases[name]; alias.kind != aliasNone {
		flag = alias.FlagClause
		invert = alias.kind == aliasNegative
	}
	return
}

// Clear the aliases and regenerate it.
// Used when the ParseContext change to a sub command.
func (f *flagGroup) resetAliases() error {
	f.aliases = nil
	return f.ensureAliases()
}

// Ensure that the aliases are evaluated.
// Called during the parsing since we do not know the nature of flags until we launch the parsing.
func (f *flagGroup) ensureAliases() error {
	if f.aliases != nil {
		return nil
	}
	// The alias map is not yet initialized, so we do it
	f.aliases = make(map[string]flagAlias)

	for _, flag := range f.flagOrder {
		if err := f.addNegativeAlias(flag.name, flag, aliasNone); err != nil {
			return err
		}
		for alias, kind := range flag.aliases {
			if err := f.addAlias(alias, flag, kind); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add an alias to the current flag group and return an error if the alias conflict with another flag.
func (f *flagGroup) addAlias(name string, flag *FlagClause, kind aliasKind) error {
	if existing := f.long[name]; existing != nil {
		return aliasErrorf("Alias %s on %s is already associated to flag %s", name, flag.name, existing.name)
	}
	if alias := f.aliases[name]; alias.kind != aliasNone && (alias.kind != kind || alias.FlagClause != flag) {
		return aliasErrorf("Alias %s on %s is already associated to flag %s", name, flag.name, alias.name)
	}
	if err := f.addNegativeAlias(name, flag, kind); err != nil {
		return err
	}

	f.aliases[name] = flagAlias{FlagClause: flag, kind: kind}
	return nil
}

// If the flag is a boolFlag, add its negative counterpart.
func (f *flagGroup) addNegativeAlias(name string, flag *FlagClause, kind aliasKind) error {
	switch kind {
	case aliasNegative:
		return nil
	}
	if fb, isSwitch := flag.value.(boolFlag); isSwitch && fb.IsBoolFlag() {
		if err := f.addAlias("no-"+name, flag, aliasNegative); err != nil {
			return err
		}
	}
	return nil
}

type aliasError string

func (ae aliasError) Error() string { return string(ae) }
func aliasErrorf(format string, args ...interface{}) aliasError {
	return aliasError(fmt.Sprintf(format, args...))
}
