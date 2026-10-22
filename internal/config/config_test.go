package config

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults",
			env:  map[string]string{"AUTH_DATABASE_URL": "postgres://x"},
			want: Config{Addr: ":8080", DatabaseURL: "postgres://x", CookieSecure: true, BcryptCost: 12, SessionTTL: 24 * time.Hour, RateBurst: 10},
		},
		{
			name: "overrides",
			env: map[string]string{
				"AUTH_DATABASE_URL": "postgres://x", "AUTH_ADDR": ":9", "AUTH_COOKIE_SECURE": "false",
				"AUTH_BCRYPT_COST": "10", "AUTH_SESSION_TTL": "1h", "AUTH_RATE_LIMIT_BURST": "50",
			},
			want: Config{Addr: ":9", DatabaseURL: "postgres://x", CookieSecure: false, BcryptCost: 10, SessionTTL: time.Hour, RateBurst: 50},
		},
		{name: "missing database url", env: map[string]string{}, wantErr: true},
		{name: "bad bool", env: map[string]string{"AUTH_DATABASE_URL": "x", "AUTH_COOKIE_SECURE": "maybe"}, wantErr: true},
		{name: "cost too low", env: map[string]string{"AUTH_DATABASE_URL": "x", "AUTH_BCRYPT_COST": "4"}, wantErr: true},
		{name: "bad burst", env: map[string]string{"AUTH_DATABASE_URL": "x", "AUTH_RATE_LIMIT_BURST": "0"}, wantErr: true},
		{name: "bad ttl", env: map[string]string{"AUTH_DATABASE_URL": "x", "AUTH_SESSION_TTL": "-1h"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(env(tt.env))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
