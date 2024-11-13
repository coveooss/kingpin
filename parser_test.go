package kingpin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserExpandFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()

	_, err = app.Parse([]string{"@" + f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "hello", *arg0)
	assert.Equal(t, "world", *arg1)
}

func TestParserExpandFromFileLeadingArg(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()

	_, err = app.Parse([]string{"prefix", "@" + f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "prefix", *arg0)
	assert.Equal(t, "hello", *arg1)
	assert.Equal(t, "world", *arg2)
}

func TestParserExpandFromFileTrailingArg(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()

	_, err = app.Parse([]string{"@" + f.Name(), "suffix"})
	assert.NoError(t, err)
	assert.Equal(t, "hello", *arg0)
	assert.Equal(t, "world", *arg1)
	assert.Equal(t, "suffix", *arg2)
}

func TestParserExpandFromFileMultipleSurroundingArgs(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()
	arg3 := app.Arg("arg3", "").String()

	_, err = app.Parse([]string{"prefix", "@" + f.Name(), "suffix"})
	assert.NoError(t, err)
	assert.Equal(t, "prefix", *arg0)
	assert.Equal(t, "hello", *arg1)
	assert.Equal(t, "world", *arg2)
	assert.Equal(t, "suffix", *arg3)
}

func TestParserExpandFromFileMultipleFlags(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("--flag1=f1\n--flag2=f2\n")
	f.Close()

	app := New("test", "")
	flag0 := app.Flag("flag0", "").String()
	flag1 := app.Flag("flag1", "").String()
	flag2 := app.Flag("flag2", "").String()
	flag3 := app.Flag("flag3", "").String()

	_, err = app.Parse([]string{"--flag0=f0", "@" + f.Name(), "--flag3=f3"})
	assert.NoError(t, err)
	assert.Equal(t, "f0", *flag0)
	assert.Equal(t, "f1", *flag1)
	assert.Equal(t, "f2", *flag2)
	assert.Equal(t, "f3", *flag3)
}

func TestParseContextPush(t *testing.T) {
	app := New("test", "")
	app.Command("foo", "").Command("bar", "")
	c := tokenize([]string{"foo", "bar"}, false)
	a := c.Next()
	assert.Equal(t, TokenArg, a.Type)
	b := c.Next()
	assert.Equal(t, TokenArg, b.Type)
	c.Push(b)
	c.Push(a)
	a = c.Next()
	assert.Equal(t, "foo", a.Value)
	b = c.Next()
	assert.Equal(t, "bar", b.Value)
}

func TestAppParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		unmanaged   []string
		elementsLen int
	}{
		{
			name:      "Single then double dash flags",
			args:      []string{"foo", "-single-dash", "--double-dash"},
			unmanaged: []string{"-single-dash", "--double-dash"},
		},
		{
			name:      "Two single dash flags",
			args:      []string{"foo", "--", "-short-flag", "-verylongshort-flag"},
			unmanaged: []string{"-short-flag", "-verylongshort-flag"},
		},
		{
			name:      "Double then single dash flags",
			args:      []string{"foo", "--double-dash", "-single-dash"},
			unmanaged: []string{"--double-dash", "-single-dash"},
		},
		{
			name:      "Verbose var flags",
			args:      []string{"foo", "-v", "-var"},
			unmanaged: []string{"-var"},
		},
		{
			name:      "Unmanaged var",
			args:      []string{"foo", "-var"},
			unmanaged: []string{"-var"},
		},
		{
			name:      "Long flag as short flag",
			args:      []string{"foo", "-test", "-verbose-level", "-another-flag"},
			unmanaged: []string{"-test", "-verbose-level", "-another-flag"},
		},
		{
			name:      "Long flag as short flag with value",
			args:      []string{"foo", "-test=123", "-verbose-level", "-another-flag"},
			unmanaged: []string{"-test=123", "-verbose-level", "-another-flag"},
		},
		{
			name:      "Long flag as short flag with negative value",
			args:      []string{"foo", "-test=-123", "-verbose-level", "-another-flag"},
			unmanaged: []string{"-test=-123", "-verbose-level", "-another-flag"},
		},
		{
			name:        "Short pseudo long flags",
			args:        []string{"foo", "-this_is_a_very_long-flag", "-this is not really a flag"},
			unmanaged:   []string{"-this_is_a_very_long-flag", "-this is not really a flag"},
			elementsLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New("test", "")
			app.allowUnmanaged = true
			app.Command("foo", "")
			app.Flag("verbose-level", "").Short('v').Alias("verbose").Bool()
			app.Flag("aflag", "").Short('a').Bool()

			ctx, err := app.ParseContext(tt.args)
			assert.Nil(t, err)
			if tt.unmanaged != nil {
				assert.Equal(t, tt.unmanaged, app.Unmanaged)
			}
			if tt.elementsLen > 0 {
				assert.Len(t, ctx.Elements, tt.elementsLen)
			}
		})
	}
}
