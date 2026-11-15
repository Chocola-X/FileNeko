package cmd

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestTrustedProxyCLIConfiguration(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		want    []string
		invalid bool
	}{
		{want: []string{"10.0.0.1"}},
		{args: []string{"--trustedProxies=127.0.0.1/32,::1/128"}, want: []string{"127.0.0.1/32", "::1/128"}},
		{args: []string{"--trustedProxies="}, want: []string{}},
		{args: []string{"--trustedProxies=localhost"}, invalid: true},
	} {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		addConfigFlags(flags)
		if err := flags.Parse(tc.args); err != nil {
			t.Fatal(err)
		}
		set := &settings.Settings{AuthMethod: auth.MethodJSONAuth}
		server := &settings.Server{TrustedProxies: []string{"10.0.0.1"}}
		_, err := getSettings(flags, set, server, &auth.JSONAuth{}, false)
		if tc.invalid {
			if err == nil {
				t.Fatal("invalid proxy accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(server.TrustedProxies, tc.want) {
			t.Fatalf("got %v, want %v", server.TrustedProxies, tc.want)
		}
	}
}

func TestTrustedProxyRuntimeOverrideAndPersistence(t *testing.T) {
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Settings.SaveServer(&settings.Server{TrustedProxies: []string{"10.0.0.1"}}); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	server, err := getServerSettings(v, st)
	if err != nil || !reflect.DeepEqual(server.TrustedProxies, []string{"10.0.0.1"}) {
		t.Fatalf("saved trust lost: %v %v", server, err)
	}
	// Exercise a real environment string, not just Viper's array representation.
	t.Setenv("NEKO_TEST_TRUSTED_PROXIES", "127.0.0.1/32,::1/128")
	if err := v.BindEnv("trustedProxies", "NEKO_TEST_TRUSTED_PROXIES"); err != nil {
		t.Fatal(err)
	}
	server, err = getServerSettings(v, st)
	if err != nil || !reflect.DeepEqual(server.TrustedProxies, []string{"127.0.0.1/32", "::1/128"}) {
		t.Fatalf("environment list failed: %v %v", server, err)
	}
	v.Set("trustedProxies", []string{})
	server, err = getServerSettings(v, st)
	if err != nil || len(server.TrustedProxies) != 0 {
		t.Fatalf("explicit override failed: %v %v", server, err)
	}
	v.Set("trustedProxies", []string{"invalid"})
	if _, err := getServerSettings(v, st); err == nil {
		t.Fatal("invalid startup trust accepted")
	}
	if err := st.Settings.SaveServer(&settings.Server{TrustedProxies: []string{"invalid"}}); err == nil {
		t.Fatal("invalid trust persisted")
	}
	stored, err := st.Settings.GetServer()
	if err != nil || !reflect.DeepEqual(stored.TrustedProxies, []string{"10.0.0.1"}) {
		t.Fatal("runtime override or invalid save changed persistent trust")
	}
}
