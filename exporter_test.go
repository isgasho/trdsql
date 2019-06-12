package trdsql

import (
	"reflect"
	"testing"
)

func TestNewExporter(t *testing.T) {
	type args struct {
		outFormat Format
	}
	tests := []struct {
		name string
		args args
		want Format
	}{
		{
			name: "test1",
			args: args{outFormat: CSV},
			want: CSV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeOpts := NewWriteOpts()
			writeOpts.OutFormat = CSV
			if got := NewExporter(writeOpts, NewWriter(writeOpts)); !reflect.DeepEqual(got.WriteOpts.OutFormat, tt.want) {
				t.Errorf("NewExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteFormat_Export(t *testing.T) {
	type fields struct {
		driver string
		dsn    string
	}
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "testErr",
			fields:  fields{driver: "sqlite3", dsn: ""},
			args:    args{query: "SELECT 1 "},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Connect(tt.fields.driver, tt.fields.dsn)
			if err != nil {
				t.Fatal("Connect error")
			}
			e := NewExporter(NewWriteOpts(), nil)
			if err := e.Export(db, tt.args.query); (err != nil) != tt.wantErr {
				t.Errorf("WriteFormat.Export() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValString(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{v: "test"},
			want: "test",
		},
		{
			name: "testNil",
			args: args{v: nil},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValString(tt.args.v); got != tt.want {
				t.Errorf("ValString() = %v, want %v", got, tt.want)
			}
		})
	}
}