package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_filterErrorOutput(t *testing.T) {
	type args struct {
		outStdErr string
	}
	tests := []struct {
		name      string
		args      args
		wantTrace string
		wantError string
	}{
		{
			name: "No TraceLevel lines",
			args: args{
				outStdErr: "Test line\nTestLine 2\n",
			},
			wantTrace: "",
			wantError: "Test line\nTestLine 2\n",
		},
		{
			name: "TraceLevel lines",
			args: args{
				outStdErr: "Test line\n++ TestLine 2\n+ Test3\n",
			},
			wantTrace: "++ TestLine 2\n+ Test3\n",
			wantError: "Test line\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotTrace, gotError := filterErrorOutput(tt.args.outStdErr)
			assert.EqualValues(t, tt.wantTrace, gotTrace, "Trace needs to be correct")
			assert.EqualValues(t, tt.wantError, gotError, "Errors need to be correct")
		})
	}
}
