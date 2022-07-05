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
			got, err := GetConfigFromFile[AllConfig]("./test_fixtures/config_fixture.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfigFromFile() \n\tgot = %#v, \n\twant %#v", got, &tt.want)
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
			got, err := GetConfigFromFile[DBConfig]("./test_fixtures/config_fixture.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfigFromFile() \n\tgot = %#v, \n\twant %#v", got, &tt.want)
			}
		})
	}
}
