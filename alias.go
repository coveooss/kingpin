package kingpin

import (
	"fmt"
	"strings"
)

type aliasMixin struct {
	aliases      []string
	autoShortcut *bool // Only set if explicitly configured
}

// Alias defines alias name that could be used instead of the long name.
func (f *FlagClause) Alias(aliases ...string) *FlagClause {
	f.aliases = aliases
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

type aliasGroupMixin struct{}

func (f *flagGroup) getFlagAlias(name string) (flag *FlagClause, invert bool, err error) {
	if err = f.ensureAliases(); err != nil {
		return
	} else if flag = f.long[name]; flag != nil {
		return
	} else if flag = f.aliases[name]; flag != nil {
		return
	} else if flag, invert = f.getNegativeFlag(name, "no-"); flag != nil {
		// This is a boolean flag and the positive name exists
		return
	} else if !strings.Contains(name, "-") {
		// When there is no - in the name, we can simply prefix the name with n to get the negative value
		if flag, invert = f.getNegativeFlag(name, "n"); flag != nil {
			// This is a boolean flag without - in the name and the positive name exists
			return
		}
	}
	return
}

func (f *flagGroup) getNegativeFlag(name, prefix string) (*FlagClause, bool) {
	if strings.HasPrefix(name, prefix) {
		if flag, _, _ := f.getFlagAlias(name[len(prefix):]); flag != nil {
			if fb, isSwitch := flag.value.(boolFlag); isSwitch && fb.IsBoolFlag() {
				return flag, true
			}
		}
	}
	return nil, false
}

func (f *flagGroup) ensureAliases() error {
	if f.aliases != nil {
		return nil
	}
	// The alias map is not yet initialized, so we do it
	f.aliases = make(map[string]*FlagClause)

	for _, flag := range f.flagOrder {
		if err := f.addShortcut(flag, flag.name); err != nil {
			return err
		}
		for _, alias := range flag.aliases {
			if err := f.addAlias(flag, alias); err != nil {
				return err
			}
			if err := f.addShortcut(flag, alias); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *flagGroup) addShortcut(flag *FlagClause, name string) error {
	if flag.autoShortcut == nil {
		flag.autoShortcut = &f.autoShortcut
	}
	if !*flag.autoShortcut {
		return nil
	}

	var shortcut string
	for _, word := range strings.Split(name, "-") {
		shortcut += string(word[0])
	}
	return f.addAlias(flag, shortcut)
}

func (f *flagGroup) addAlias(flag *FlagClause, name string) error {
	if existing := f.long[name]; existing != nil {
		return fmt.Errorf("Alias %s on %s is already associated to flag %s", name, flag.name, existing.name)
	}
	if existing := f.aliases[name]; existing != nil && existing != flag {
		return fmt.Errorf("Alias %s on %s is already associated to flag %s", name, flag.name, existing.name)
	}
	f.aliases[name] = flag
	return nil
}
