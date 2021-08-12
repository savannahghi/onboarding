package extension

import (
	"reflect"
	"testing"

	"github.com/savannahghi/firebasetools"
)

func TestNewBaseExtensionImpl(t *testing.T) {

	fc := &firebasetools.MockFirebaseClient{}
	_, err := fc.InitFirebase()
	if err != nil {
		t.Errorf("Failed to init Firebase: %v", err)
	}

	type args struct {
		fc firebasetools.IFirebaseClient
	}
	tests := []struct {
		name string
		args args
		want BaseExtension
	}{
		{
			name: "Happy test",
			args: args{
				fc: &firebasetools.MockFirebaseClient{},
			},
			want: &BaseExtensionImpl{
				fc: fc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBaseExtensionImpl(tt.args.fc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBaseExtensionImpl() = %v, want %v", got, tt.want)
			}
		})
	}
}
