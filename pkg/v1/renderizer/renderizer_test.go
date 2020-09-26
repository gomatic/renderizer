package renderizer

import (
	"reflect"
	"testing"
)

func TestRender(t *testing.T) {
	type args struct {
		settings Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Renderizer(&tt.args.settings).Render(); (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_Render(t *testing.T) {
	tests := []struct {
		name     string
		settings *Options
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.settings.Render(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Render() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_typer(t *testing.T) {
	type args struct {
		d string
	}
	tests := []struct {
		name       string
		settings   Options
		args       args
		wantResult interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.settings.typer(tt.args.d); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Options.typer() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_retypeSingleElementSlice(t *testing.T) {
	type args struct {
		config *retyperConfig
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retypeSingleElementSlice(tt.args.config)
		})
	}
}

func TestOptions_retyping(t *testing.T) {
	type args struct {
		source map[string]interface{}
		config retyperConfig
	}
	tests := []struct {
		name     string
		settings Options
		args     args
		want     map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.settings.retyping(tt.args.source, tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.retyping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_Retyper(t *testing.T) {
	type args struct {
		source  map[string]interface{}
		options []retyperOptions
	}
	tests := []struct {
		name     string
		settings Options
		args     args
		want     map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.settings.Retyper(tt.args.source, tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.Retyper() = %v, want %v", got, tt.want)
			}
		})
	}
}
