package configs

import (
	"reflect"
	"testing"
)

func Test_readConf(t *testing.T) {
	tests := []struct {
		name    string
		want    AllConfig
		wantErr bool
	}{
		{
			name: "Simple test",
			want: AllConfig{
				ServiceConfig: ServiceConfig{
					AuthKey:    "authenticated",
					UserIDKey:  "user_id",
					BcryptCost: 14,
					Port:       ":5000",
				},
				SessionStoreConfig: SessionStoreConfig{
					CookieHTTPOnly: true,
					CookieSecure:   false,
					Expiration:     5,
				},
				DBConfig: DBConfig{
					Host:     "host",
					Port:     "port",
					User:     "user",
					Password: "password",
					Dbname:   "dbname",
					Sslmode:  "sslmode"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigFromFile[AllConfig]("./test_fixtures/config_fixture.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("getConfigFromFile() \n\tgot = %#v, \n\twant %#v", got, &tt.want)
			}
		})
	}
}

func Test_DBTest(t *testing.T) {
	tests := []struct {
		name    string
		want    DBConfig
		wantErr bool
	}{
		{
			name: "Simple test",
			want: DBConfig{
				Host:     "host",
				Port:     "port",
				User:     "user",
				Password: "password",
				Dbname:   "dbname",
				Sslmode:  "sslmode",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigFromFile[DBConfig]("./test_fixtures/config_fixture.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("getConfigFromFile() \n\tgot = %#v, \n\twant %#v", got, &tt.want)
			}
		})
	}
}
